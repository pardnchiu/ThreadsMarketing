package llm

import (
	"context"
	_ "embed"
	"fmt"

	utils_http "github.com/pardnchiu/go-utils/http"
)

const messagesURL = "https://api.anthropic.com/v1/messages"

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type response struct {
	Content    []contentBlock `json:"content"`
	StopReason string         `json:"stop_reason"`
	Model      string         `json:"model"`
}

//go:embed dna.md
var dna string

const defaultUserPrompt = `依照上方人設 DNA，生成一篇 Threads 貼文。

硬性要求：
- 直接輸出貼文純文字，不要任何前言、解釋、markdown 格式、引號包覆
- <= 500 字元（Threads 限制）
- 從 4 種風格（opinion / news / meme / qa）中自行選擇一種最適合今日心境的
- 遵守第 5 節格式約束、第 6 節禁止清單、第 8 節禁止詞

只輸出貼文本體。`

type Client struct {
	apiKey string
	model  string
}

func New(apiKey, model string) *Client {
	if model == "" {
		model = "claude-sonnet-4-6"
	}
	return &Client{apiKey: apiKey, model: model}
}

func (c *Client) Generate(ctx context.Context) (string, error) {
	body := map[string]any{
		"model":      c.model,
		"max_tokens": 1024,
		"system":     dna,
		"messages": []message{
			{Role: "user", Content: defaultUserPrompt},
		},
	}
	header := map[string]string{
		"x-api-key":         c.apiKey,
		"anthropic-version": "2023-06-01",
	}

	resp, code, err := utils_http.POST[response](ctx, nil, messagesURL, header, body, "json")
	if err != nil {
		return "", fmt.Errorf("anthropic POST: %w", err)
	}
	if code < 200 || code >= 300 {
		return "", fmt.Errorf("anthropic http %d", code)
	}
	if len(resp.Content) == 0 {
		return "", fmt.Errorf("empty response")
	}
	return resp.Content[0].Text, nil
}
