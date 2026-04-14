package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	hnTopStoriesURL = "https://hacker-news.firebaseio.com/v0/topstories.json"
	hnItemURLFmt    = "https://hacker-news.firebaseio.com/v0/item/%d.json"
	hnMaxStories    = 30
)

type hnItem struct {
	URL   string `json:"url"`
	Title string `json:"title"`
	Type  string `json:"type"`
}

func fetchHN(ctx context.Context) []string {
	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, hnTopStoriesURL, nil)
	if err != nil {
		return nil
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	data, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil
	}

	var ids []int64
	if err := json.Unmarshal(data, &ids); err != nil {
		return nil
	}
	if len(ids) > hnMaxStories {
		ids = ids[:hnMaxStories]
	}

	out := make([]string, 0, len(ids))
	for _, id := range ids {
		itemURL := fmt.Sprintf(hnItemURLFmt, id)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, itemURL, nil)
		if err != nil {
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			continue
		}
		var item hnItem
		if err := json.Unmarshal(body, &item); err != nil {
			continue
		}
		link := strings.TrimSpace(item.URL)
		if link == "" {
			continue
		}
		out = append(out, link)
	}
	return out
}
