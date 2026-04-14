package bot

import (
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var excludedPathSegments = map[string]struct{}{
	"av":       {},
	"video":    {},
	"videos":   {},
	"audio":    {},
	"audios":   {},
	"sound":    {},
	"sounds":   {},
	"podcast":  {},
	"podcasts": {},
	"live":     {},
}

func isMediaURL(href string) bool {
	u, err := url.Parse(href)
	if err != nil {
		return false
	}
	for seg := range strings.SplitSeq(u.Path, "/") {
		if _, ok := excludedPathSegments[strings.ToLower(seg)]; ok {
			return true
		}
	}
	return false
}

type rssFeed struct {
	XMLName xml.Name `xml:"rss"`
	Channel struct {
		Item []struct {
			Link string `xml:"link"`
		} `xml:"item"`
	} `xml:"channel"`
}

var defaultFeeds = []string{
	// BBC
	"https://feeds.bbci.co.uk/news/rss.xml",
	"https://feeds.bbci.co.uk/news/business/rss.xml",
	"https://feeds.bbci.co.uk/news/entertainment_and_arts/rss.xml",
	"https://feeds.bbci.co.uk/news/health/rss.xml",
	"https://feeds.bbci.co.uk/news/science_and_environment/rss.xml",
	"https://feeds.bbci.co.uk/news/technology/rss.xml",
	"https://feeds.bbci.co.uk/news/world/rss.xml",
	"https://feeds.bbci.co.uk/zhongwen/trad/rss.xml",

	// The Guardian
	"https://www.theguardian.com/world/rss",
	"https://www.theguardian.com/science/rss",
	"https://www.theguardian.com/politics/rss",
	"https://www.theguardian.com/uk/business/rss",
	"https://www.theguardian.com/uk/technology/rss",
	"https://www.theguardian.com/uk/environment/rss",
	"https://www.theguardian.com/uk/money/rss",

	// 台灣
	"https://news.ltn.com.tw/rss/all.xml",
	"https://udn.com/rssfeed/news/2/6638?ch=news",
	"https://feeds.feedburner.com/ettoday/news",
	"https://tw.appledaily.com/rss",

	// HN
	"https://hnrss.org/frontpage",
}

func fetchRSS(ctx context.Context, feeds []string) []string {
	client := &http.Client{Timeout: 30 * time.Second}
	seen := make(map[string]struct{})
	out := make([]string, 0, 64)

	for _, feed := range feeds {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, feed, nil)
		if err != nil {
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		data, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			continue
		}

		var rss rssFeed
		if err := xml.Unmarshal(data, &rss); err != nil {
			continue
		}
		for _, item := range rss.Channel.Item {
			link := strings.TrimSpace(item.Link)
			if link == "" {
				continue
			}
			if isMediaURL(link) {
				continue
			}
			if _, ok := seen[link]; ok {
				continue
			}
			seen[link] = struct{}{}
			out = append(out, link)
		}
	}
	return out
}
