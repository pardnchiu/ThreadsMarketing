package main

import (
	"strings"
	"sync"

	"github.com/common-nighthawk/go-figure"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	appOnce       sync.Once
	app           *tview.Application
	dashboardView *tview.TextView
	inputView     *tview.InputField
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
	})

	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(dashboardView, 0, 1, false).
		AddItem(inputView, 3, 0, true)

	app.SetRoot(layout, true)

	_, _, width, _ := dashboardView.GetInnerRect()
	seperate := strings.Repeat("─", width)

	var sb strings.Builder
	ascii := figure.NewFigure("ThreadsMarketing", "thick", true)
	sb.WriteString(ascii.String())
	sb.WriteString(seperate + "\n")

	dashboardView.SetText(sb.String())

	if err := app.Run(); err != nil {
		panic(err)
	}
}

func commandActions(key tcell.Key) {
	if key != tcell.KeyEnter {
		return
	}

	text := strings.TrimSpace(inputView.GetText())
	if text == "" {
		return
	}

	text = strings.ToLower(text)
	if text == "exit" || text == "quit" || text == "q" {
		app.Stop()
		return
	}

	dashboardView.Write([]byte(text + "\n"))
	inputView.SetText("")
}
