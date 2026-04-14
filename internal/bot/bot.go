package bot

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/pardnchiu/ThreadsMarketing/internal/database/postgresql"
	"github.com/pardnchiu/go-utils/rod"
)

const (
	contentTypeNews = "news"
	maxFetchRetries = 3
)

type Logger func(msg string)

type Bot struct {
	db         *sql.DB
	log        Logger
	downloadMu sync.Mutex
	failMu     sync.Mutex
	failCount  map[string]int
}

func New(db *sql.DB, log Logger) *Bot {
	if log == nil {
		log = func(string) {}
	}
	return &Bot{db: db, log: log, failCount: make(map[string]int)}
}

func (b *Bot) Close() {
	rod.Close()
}

func (b *Bot) CollectURLs(ctx context.Context) {
	urls := b.gatherURLs(ctx)

	added := 0
	for _, href := range urls {
		if ctx.Err() != nil {
			return
		}
		inserted, err := postgresql.InsertPendingURL(ctx, b.db, href, contentTypeNews)
		if err != nil {
			b.log(fmt.Sprintf("[collector] insert fail %s: %v", href, err))
			continue
		}
		if inserted {
			added++
		}
	}
	b.log(fmt.Sprintf("[collector] scanned=%d new=%d", len(urls), added))
}

func (b *Bot) Download(ctx context.Context) {
	if !b.downloadMu.TryLock() {
		return
	}
	defer b.downloadMu.Unlock()

	urls, err := postgresql.ListPendingURLs(ctx, b.db)
	if err != nil {
		b.log(fmt.Sprintf("[downloader] list fail: %v", err))
		return
	}
	if len(urls) == 0 {
		return
	}

	b.log(fmt.Sprintf("[downloader] %d pending", len(urls)))
	ok, fail := 0, 0
	for _, href := range urls {
		if ctx.Err() != nil {
			return
		}
		content, err := rod.Fetch(ctx, href, &rod.FetchOption{Output: rod.OutputMarkdown})
		if err != nil {
			fail++
			n := b.bumpFail(href)
			b.log(fmt.Sprintf("[downloader] fetch fail (%d/%d) %s: %v", n, maxFetchRetries, href, err))
			if n >= maxFetchRetries {
				if derr := postgresql.DismissURL(ctx, b.db, href); derr != nil {
					b.log(fmt.Sprintf("[downloader] dismiss fail %s: %v", href, derr))
				} else {
					b.log(fmt.Sprintf("[downloader] dismissed %s after %d failures", href, n))
					b.clearFail(href)
				}
			}
			continue
		}
		if err := postgresql.MarkDownloaded(ctx, b.db, href, content); err != nil {
			fail++
			b.log(fmt.Sprintf("[downloader] mark fail %s: %v", href, err))
			continue
		}
		b.clearFail(href)
		ok++
	}
	b.log(fmt.Sprintf("[downloader] done ok=%d fail=%d", ok, fail))
}

func (b *Bot) bumpFail(href string) int {
	b.failMu.Lock()
	defer b.failMu.Unlock()
	b.failCount[href]++
	return b.failCount[href]
}

func (b *Bot) clearFail(href string) {
	b.failMu.Lock()
	defer b.failMu.Unlock()
	delete(b.failCount, href)
}

func (b *Bot) gatherURLs(ctx context.Context) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0, 128)

	for _, u := range fetchRSS(ctx, defaultFeeds) {
		if _, ok := seen[u]; ok {
			continue
		}
		seen[u] = struct{}{}
		out = append(out, u)
	}
	for _, u := range fetchHN(ctx) {
		if _, ok := seen[u]; ok {
			continue
		}
		seen[u] = struct{}{}
		out = append(out, u)
	}
	return out
}
