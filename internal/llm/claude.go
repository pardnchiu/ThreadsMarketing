package llm

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"strings"
	"sync"

	utils_http "github.com/pardnchiu/go-utils/http"
)

//go:embed dna.md
var dna string

type Style string

const (
	StyleOpinion Style = "opinion"
	StyleNews    Style = "news"
	StyleMeme    Style = "meme"
	StyleQA      Style = "qa"
)

func (s Style) isNews() bool { return s == StyleNews }

const defaultUserPrompt = `依照上方人設 DNA，生成一篇 Threads 貼文。

硬性要求：
- 直接輸出貼文純文字，不要任何前言、解釋、markdown 格式、引號包覆
- <= 500 字元（Threads 限制）
- 遵守第 5 節格式約束、第 6 節禁止清單、第 8 節禁止詞

`

const injectKnowledge = `本篇風格：%s（A 類純知識分享）
僅使用既有知識撰寫，不要引用版本號或發布日期。不要使用任何搜尋類關鍵字。

只輸出貼文本體。`

const injectNews = `本篇風格：news（B 類最新資訊分享）
請透過 web search 搜尋網路取得最新資訊，建議關鍵字：%s release notes、%s changelog、%s latest version、HackerNews %s。
根據搜尋結果撰寫，版本號、日期、feature 以搜尋結果為準。

只輸出貼文本體。`

type sendResponse struct {
	Text      string `json:"text"`
	SessionID string `json:"session_id"`
}

const (
	SessionGenerate = "generate"
	sessionDir      = "internal/llm/sessions"
)

type Client struct {
	endpoint string
	mu       sync.Mutex
}

func New() *Client {
	ep := os.Getenv("AGENVOY_URL")
	if ep == "" {
		ep = "http://localhost:17989"
	}
	return &Client{endpoint: strings.TrimRight(ep, "/") + "/v1/send"}
}

func sessionPath(key string) string {
	return sessionDir + "/" + key
}

func readSession(key string) string {
	data, err := os.ReadFile(sessionPath(key))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func writeSession(key, id string) error {
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return err
	}
	tmp := sessionPath(key) + ".tmp"
	if err := os.WriteFile(tmp, []byte(id), 0644); err != nil {
		return err
	}
	return os.Rename(tmp, sessionPath(key))
}

func (c *Client) send(ctx context.Context, sessionKey, system, content string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	body := map[string]any{
		"content":       content,
		"sse":           false,
		"system_prompt": system,
		"persist":       true,
	}
	if sid := readSession(sessionKey); sid != "" {
		body["session_id"] = sid
	}
	resp, code, err := utils_http.POST[sendResponse](ctx, nil, c.endpoint, nil, body, "json")
	if err != nil {
		return "", fmt.Errorf("agenvoy POST: %w", err)
	}
	if code < 200 || code >= 300 {
		return "", fmt.Errorf("agenvoy http %d", code)
	}
	if resp.SessionID != "" && resp.SessionID != readSession(sessionKey) {
		_ = writeSession(sessionKey, resp.SessionID)
	}
	return strings.TrimSpace(resp.Text), nil
}

func buildUserPrompt(style Style, topic string) string {
	if style.isNews() {
		t := strings.TrimSpace(topic)
		if t == "" {
			t = "Go backend / database / devops"
		}
		return defaultUserPrompt + fmt.Sprintf(injectNews, t, t, t, t)
	}
	return defaultUserPrompt + fmt.Sprintf(injectKnowledge, style)
}

func PickStyle(r func(n int) int) Style {
	switch n := r(10); {
	case n < 4:
		return StyleOpinion
	case n < 7:
		return StyleNews
	case n < 9:
		return StyleMeme
	default:
		return StyleQA
	}
}

func (c *Client) Generate(ctx context.Context, style Style, topic string) (string, error) {
	return c.send(ctx, SessionGenerate, dna, buildUserPrompt(style, topic))
}
