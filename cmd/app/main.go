package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/common-nighthawk/go-figure"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	utils_keychain "github.com/pardnchiu/go-utils/filesystem/keychain"

	"github.com/pardnchiu/ThreadsMarketing/internal/threads"
)

type loginFlow int

const (
	flowNone loginFlow = iota
	flowAppID
	flowAppSecret
	flowShortToken
)

const serviceName = "ThreadsMarketing"

var (
	configDir = filepath.Join(os.Getenv("HOME"), ".config", serviceName)

	appOnce       sync.Once
	app           *tview.Application
	dashboardView *tview.TextView
	inputView     *tview.InputField

	authLoginFlow loginFlow
	authAppID     string
	authAppSecret string
)

func main() {
	appOnce.Do(func() {
		app = tview.NewApplication()

		dashboardView = tview.NewTextView().
			SetDynamicColors(true).
			SetScrollable(true).
			SetChangedFunc(func() {
				app.QueueUpdateDraw(func() {})
			})
		dashboardView.SetBorder(true).
			SetBorderColor(tcell.ColorWhite)

		inputView = tview.NewInputField()
		inputView.SetFieldBackgroundColor(tcell.ColorBlack).
			SetDoneFunc(commandActions)
		inputView.SetBorder(true).
			SetBorderColor(tcell.ColorWhite)

		app.SetAfterDrawFunc(func(screen tcell.Screen) {
			app.SetAfterDrawFunc(nil)
			setDefault()
			go verifyLogin()
		})
	})

	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(dashboardView, 0, 1, false).
		AddItem(inputView, 3, 0, true)

	app.SetRoot(layout, true)
	if err := app.Run(); err != nil {
		panic(err)
	}
}

func setDefault() {
	dashboardView.SetText("")

	_, _, width, _ := dashboardView.GetInnerRect()
	seperate := strings.Repeat("─", width/2)

	var sb strings.Builder
	ascii := figure.NewFigure("ThreadsMarketing", "thick", true)
	sb.WriteString(ascii.String())
	sb.WriteString(seperate + "\n")

	dashboardView.SetText(sb.String())
}

func commandActions(key tcell.Key) {
	if key != tcell.KeyEnter {
		return
	}

	text := strings.TrimSpace(inputView.GetText())
	if text == "" {
		return
	}
	inputView.SetText("")

	if authLoginFlow != flowNone {
		login(text)
		return
	}

	text = strings.ToLower(text)
	switch text {
	case "exit", "quit", "q":
		app.Stop()
		return

	case "clear", "c":
		setDefault()
		return

	case "login":
		authLoginFlow = flowAppID
		writeLog("[login] app id:")
		return
	}

	writeLog("[command] unknown: " + text)
}

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

			pairs := [][2]string{
				{"access_token", result.AccessToken},
				{"app_secret", secret},
				{"user_id", result.UserID},
			}
			for _, p := range pairs {
				if err := utils_keychain.Set(serviceName, configDir, p[0], p[1]); err != nil {
					rewriteLog(fmt.Sprintf("[login] failed to save token: %v", err))
					return
				}
			}

			days := result.ExpiresIn / 86400
			name := result.UserID
			if result.Username != "" {
				name = "@" + result.Username
			}
			rewriteLog(fmt.Sprintf("[login] account: %s, expires in %d", name, days))
		}()
	}
}

func verifyLogin() {
	token := utils_keychain.Get(serviceName, configDir, "access_token")
	if token == "" {
		writeLog("[auth] please login first")
		return
	}

	writeLog("[auth] token exist, verifying")

	userID, username, err := threads.Verify(context.Background(), token)
	if err != nil {
		utils_keychain.Delete(serviceName, configDir, "access_token")
		rewriteLog(fmt.Sprintf("[auth] failed to verify: %v", err))
		writeLog("[auth] please login again")
		return
	}

	name := userID
	if username != "" {
		name = "@" + username
	}
	rewriteLog(fmt.Sprintf("[auth] account: %s", name))
}

func writeLog(msg string) {
	dashboardView.Write([]byte(msg + "\n"))
}

func rewriteLog(msg string) {
	app.QueueUpdateDraw(func() {
		text := dashboardView.GetText(true)
		if i := strings.LastIndex(strings.TrimRight(text, "\n"), "\n"); i >= 0 {
			text = text[:i+1]
		}
		dashboardView.SetText(text + msg + "\n")
	})
}
