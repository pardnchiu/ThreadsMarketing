package tui

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/pardnchiu/ThreadsMarketing/internal/threads"
	utils_keychain "github.com/pardnchiu/go-utils/filesystem/keychain"
)

func login(input string) {
	switch authLoginFlow {
	case flowAppID:
		authAppID = input
		authLoginFlow = flowAppSecret
		inputView.SetMaskCharacter('*')
		rewriteLog("[login] app secret:")

	case flowAppSecret:
		authAppSecret = input
		authLoginFlow = flowShortToken
		rewriteLog("[login] short-lived token: (u can get this from gui test in developer tools)")

	case flowShortToken:
		shortToken := input
		inputView.SetMaskCharacter(0)
		authLoginFlow = flowNone
		rewriteLog("[login] get Long-live token")

		appID := authAppID
		secret := authAppSecret
		authAppID = ""
		authAppSecret = ""

		go func() {
			result, err := threads.Exchange(context.Background(), appID, secret, shortToken)
			if err != nil {
				rewriteLog(fmt.Sprintf("[login] failed to login: %v", err))
				return
			}
			if err := persistToken(result, secret); err != nil {
				rewriteLog(fmt.Sprintf("[login] failed to save token: %v", err))
				return
			}
			rewriteLog(fmt.Sprintf("[login] account: %s, expires in %d days", displayName(result), result.ExpiresIn/86400))
		}()
	}
}

func persistToken(r *threads.Result, appSecret string) error {
	expiresAt := time.Now().Add(time.Duration(r.ExpiresIn) * time.Second).Unix()
	pairs := [][2]string{
		{"access_token", r.AccessToken},
		{"user_id", r.UserID},
		{"expires_at", strconv.FormatInt(expiresAt, 10)},
	}
	if appSecret != "" {
		pairs = append(pairs, [2]string{"app_secret", appSecret})
	}
	for _, p := range pairs {
		if err := utils_keychain.Set(p[0], p[1]); err != nil {
			return err
		}
	}
	return nil
}

func logout() {
	for _, k := range []string{"access_token", "app_secret", "user_id", "expires_at"} {
		utils_keychain.Delete(k)
	}
	writeLog("[logout] done")
}
