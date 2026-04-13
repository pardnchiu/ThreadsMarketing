package tui

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/pardnchiu/ThreadsMarketing/internal/threads"
	utils_keychain "github.com/pardnchiu/go-utils/filesystem/keychain"
)

func verifyLogin() {
	token := utils_keychain.Get("access_token")
	if token == "" {
		writeLog("[auth] please login first")
		return
	}

	writeLog("[auth] token exist, verifying")

	ts := utils_keychain.Get("expires_at")
	if ts == "" {
		rewriteLog("[auth] missing expiry, refreshing to populate")
		tryRefresh(token)
		return
	}
	sec, parseErr := strconv.ParseInt(ts, 10, 64)
	if parseErr != nil {
		rewriteLog("[auth] invalid expiry, refreshing")
		tryRefresh(token)
		return
	}
	remaining := time.Until(time.Unix(sec, 0))
	if remaining <= 0 {
		rewriteLog("[auth] token expired, refreshing")
		tryRefresh(token)
		return
	}
	if remaining < refreshAheadHr {
		rewriteLog(fmt.Sprintf("[auth] token expires in %s, refreshing", formatExpiresIn(remaining)))
		tryRefresh(token)
		return
	}

	userID, username, err := threads.Verify(context.Background(), token)
	if err != nil {
		if errors.Is(err, threads.ErrUnauthorized) {
			rewriteLog("[auth] token unauthorized, refreshing")
			if !tryRefresh(token) {
				return
			}
			return
		}
		rewriteLog(fmt.Sprintf("[auth] failed to verify: %v", err))
		return
	}

	name := userID
	if username != "" {
		name = "@" + username
	}
	rewriteLog(fmt.Sprintf("[auth] account: %s, expires in %s", name, formatExpiresIn(remaining)))
}

func tryRefresh(accessToken string) bool {
	result, err := threads.Refresh(context.Background(), accessToken)
	if err != nil {
		logout()
		rewriteLog(fmt.Sprintf("[auth] refresh failed: %v", err))
		writeLog("[auth] please login again")
		return false
	}
	if err := persistToken(result, ""); err != nil {
		rewriteLog(fmt.Sprintf("[auth] failed to save token: %v", err))
		return false
	}
	rewriteLog(fmt.Sprintf("[auth] account: %s, expires in %d days", displayName(result), result.ExpiresIn/86400))
	return true
}

func formatExpiresIn(d time.Duration) string {
	if d < 0 {
		return "expired"
	}
	days := int(d / (24 * time.Hour))
	hours := int(d%(24*time.Hour)) / int(time.Hour)
	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	return fmt.Sprintf("%dh", hours)
}
