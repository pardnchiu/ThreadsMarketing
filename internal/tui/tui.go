package tui

import (
	"strings"
	"sync"
	"time"

	"database/sql"

	"github.com/common-nighthawk/go-figure"
	"github.com/gdamore/tcell/v2"
	"github.com/pardnchiu/ThreadsMarketing/internal/scheduler"
	"github.com/pardnchiu/ThreadsMarketing/internal/threads"
	"github.com/rivo/tview"
)

type loginFlow int

const (
	flowNone loginFlow = iota
	flowAppID
	flowAppSecret
	flowShortToken
)

const (
	refreshAheadHr = 5 * 24 * time.Hour
)

var (
	appOnce       sync.Once
	app           *tview.Application
	dashboardView *tview.TextView
	inputView     *tview.InputField

	authLoginFlow loginFlow
	authAppID     string
	authAppSecret string
)

func New(db *sql.DB) *tview.Application {
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

		app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyTab {
				if app.GetFocus() == inputView {
					app.SetFocus(dashboardView)
				} else {
					app.SetFocus(inputView)
				}
				return nil
			}
			return event
		})

		app.SetAfterDrawFunc(func(screen tcell.Screen) {
			app.SetAfterDrawFunc(nil)
			setDefault()
			go verifyLogin()
			go func() {
				if err := scheduler.Start(db, writeLog, rewriteLog); err != nil {
					writeLog("[scheduler] " + err.Error())
				}
			}()
		})
	})

	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(dashboardView, 0, 1, false).
		AddItem(inputView, 3, 0, true)

	app.SetRoot(layout, true)
	return app
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

func displayName(r *threads.Result) string {
	if r.Username != "" {
		return "@" + r.Username
	}
	return r.UserID
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
