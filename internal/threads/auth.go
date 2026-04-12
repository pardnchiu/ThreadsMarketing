package threads

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	utils_http "github.com/pardnchiu/go-utils/http"
)

const graphBase = "https://graph.threads.net"

type token struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type user struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type result struct {
	AccessToken string
	ExpiresIn   int
	UserID      string
	Username    string
}

func Verify(ctx context.Context, accessToken string) (userID, username string, err error) {
	link := fmt.Sprintf("%s/me?access_token=%s", graphBase, url.QueryEscape(accessToken))
	u, code, err := utils_http.GET[user](ctx, nil, link, nil)
	if err != nil {
		return "", "", fmt.Errorf("utils_http.GET: %w", err)
	}
	if code == http.StatusUnauthorized {
		return "", "", fmt.Errorf("utils_http.GET: 401")
	}
	if code != http.StatusOK {
		return "", "", fmt.Errorf("utils_http.GET: %d", code)
	}
	if u.ID == "" {
		return "", "", fmt.Errorf("utils_http.GET: empty user ID")
	}
	return u.ID, u.Username, nil
}

func Exchange(ctx context.Context, appID, appSecret, shortToken string) (*result, error) {
	link := fmt.Sprintf("%s/access_token?grant_type=th_exchange_token&client_id=%s&client_secret=%s&access_token=%s",
		graphBase,
		url.QueryEscape(appID),
		url.QueryEscape(appSecret),
		url.QueryEscape(shortToken),
	)

	t, code, err := utils_http.GET[token](ctx, nil, link, nil)
	if err != nil {
		return nil, fmt.Errorf("utils_http.GET: %w", err)
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("utils_http.GET: %d", code)
	}
	if t.AccessToken == "" {
		return nil, fmt.Errorf("utils_http.GET: empty access_token in response")
	}

	meLink := fmt.Sprintf("%s/me?access_token=%s", graphBase, url.QueryEscape(t.AccessToken))
	u, code, err := utils_http.GET[user](ctx, nil, meLink, nil)
	if err != nil {
		return nil, fmt.Errorf("utils_http.GET: %w", err)
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("utils_http.GET: %d", code)
	}
	if u.ID == "" {
		return nil, fmt.Errorf("utils_http.GET: empty user ID in response")
	}

	return &result{
		AccessToken: t.AccessToken,
		ExpiresIn:   t.ExpiresIn,
		UserID:      u.ID,
		Username:    u.Username,
	}, nil
}
