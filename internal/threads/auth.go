package threads

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	utils_http "github.com/pardnchiu/go-utils/http"
)

const graphBase = "https://graph.threads.net"

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrRateLimited  = errors.New("rate limited")
	ErrServer       = errors.New("server error")
)

type token struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type user struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type Result struct {
	AccessToken string
	ExpiresIn   int
	UserID      string
	Username    string
}

func classify(code int) error {
	switch {
	case code == http.StatusOK:
		return nil
	case code == http.StatusUnauthorized:
		return ErrUnauthorized
	case code == http.StatusTooManyRequests:
		return ErrRateLimited
	case code >= 500:
		return fmt.Errorf("%w: %d", ErrServer, code)
	default:
		return fmt.Errorf("http %d", code)
	}
}

func Verify(ctx context.Context, accessToken string) (userID, username string, err error) {
	link := fmt.Sprintf("%s/me?access_token=%s", graphBase, url.QueryEscape(accessToken))
	u, code, err := utils_http.GET[user](ctx, nil, link, nil)
	if err != nil {
		return "", "", fmt.Errorf("utils_http.GET: %w", err)
	}
	if err := classify(code); err != nil {
		return "", "", err
	}
	if u.ID == "" {
		return "", "", fmt.Errorf("empty user ID")
	}
	return u.ID, u.Username, nil
}

func Exchange(ctx context.Context, appID, appSecret, shortToken string) (*Result, error) {
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
	if err := classify(code); err != nil {
		return nil, err
	}
	if t.AccessToken == "" {
		return nil, fmt.Errorf("empty access_token")
	}

	userID, username, err := Verify(ctx, t.AccessToken)
	if err != nil {
		return nil, err
	}

	return &Result{
		AccessToken: t.AccessToken,
		ExpiresIn:   t.ExpiresIn,
		UserID:      userID,
		Username:    username,
	}, nil
}

func Refresh(ctx context.Context, accessToken string) (*Result, error) {
	link := fmt.Sprintf("%s/refresh_access_token?grant_type=th_refresh_token&access_token=%s",
		graphBase, url.QueryEscape(accessToken))

	t, code, err := utils_http.GET[token](ctx, nil, link, nil)
	if err != nil {
		return nil, fmt.Errorf("utils_http.GET: %w", err)
	}
	if err := classify(code); err != nil {
		return nil, err
	}
	if t.AccessToken == "" {
		return nil, fmt.Errorf("empty access_token")
	}

	userID, username, err := Verify(ctx, t.AccessToken)
	if err != nil {
		return nil, err
	}

	return &Result{
		AccessToken: t.AccessToken,
		ExpiresIn:   t.ExpiresIn,
		UserID:      userID,
		Username:    username,
	}, nil
}
