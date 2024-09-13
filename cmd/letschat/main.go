package main

import (
	"github.com/M0hammadUsman/letschat/internal/client"
	"github.com/M0hammadUsman/letschat/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmittmann/tint"
	zone "github.com/lrstanley/bubblezone"
	"log/slog"
	"os"
)

func main() {
	slogger := slog.New(tint.NewHandler(os.Stderr, nil))
	// using it as initialization if err occurs we halt the application on startup rather than having issues while the
	// app is running
	if err := client.Init(); err != nil {
		slogger.Error(err.Error())
		os.Exit(1)
	}
	f, err := tea.LogToFile("Letschat.log", "Letschat")
	if err != nil {
		slogger.Error(err.Error())
		os.Exit(1)
	}
	defer f.Close()

	zone.NewGlobal()
	_, err = tea.NewProgram(tui.InitialTabContainerModel(), tea.WithAltScreen(), tea.WithMouseAllMotion()).Run()
	if err != nil {
		slogger.Error(err.Error())
		os.Exit(1)
	}
}
