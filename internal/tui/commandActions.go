package tui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	utils_keychain "github.com/pardnchiu/go-utils/filesystem/keychain"
)

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
		if utils_keychain.Get("access_token") != "" {
			writeLog("[login] already logged in, use `logout` first")
			return
		}
		authLoginFlow = flowAppID
		writeLog("[login] app id:")
		return

	case "logout":
		logout()
		return
	}

	writeLog("[command] unknown: " + text)
}
