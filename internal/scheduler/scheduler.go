package scheduler

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand/v2"
	"sync/atomic"
	"time"

	scheduler "github.com/pardnchiu/go-scheduler"

	"github.com/pardnchiu/ThreadsMarketing/internal/llm"
)

const (
	minInterval = 5 * time.Minute
	maxInterval = 10 * time.Minute
)

type Logger func(msg string)

func Start(db *sql.DB, writeLog, rewriteLog Logger) error {
	gen := llm.New()

	location, _ := time.LoadLocation("Asia/Taipei")
	cron, err := scheduler.New(scheduler.Config{Location: location})
	if err != nil {
		return fmt.Errorf("scheduler init: %w", err)
	}

	var next atomic.Int64
	next.Store(time.Now().Unix())

	_, err = cron.Add("@every 30s", func() {
		if time.Now().Unix() < next.Load() {
			return
		}

		rewriteLog("[gen] generating post")
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		style := llm.PickStyle(rand.IntN)
		text, err := gen.Generate(ctx, style, "")
		if err != nil {
			rewriteLog(fmt.Sprintf("[gen] failed: %v", err))
		} else {
			rewriteLog(fmt.Sprintf("[gen] ─── post (%s) ───\n%s\n[gen] ─── end ───", style, text))
		}

		delta := minInterval + time.Duration(rand.Int64N(int64(maxInterval-minInterval)))
		next.Store(time.Now().Add(delta).Unix())
		writeLog(fmt.Sprintf("[gen] next run in %s", formatNext(delta)))
	}, "generate_post")

	if err != nil {
		return fmt.Errorf("scheduler add: %w", err)
	}

	_ = db
	cron.Start()
	return nil
}

func formatNext(d time.Duration) string {
	m := int(d / time.Minute)
	s := int(d%time.Minute) / int(time.Second)
	return fmt.Sprintf("%dm%02ds", m, s)
}
