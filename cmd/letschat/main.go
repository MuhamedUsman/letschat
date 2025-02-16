package main

import (
	"flag"
	"github.com/MuhamedUsman/letschat/internal/client"
	"github.com/MuhamedUsman/letschat/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmittmann/tint"
	zone "github.com/lrstanley/bubblezone"
	"log/slog"
	"os"
)

func main() {

	var key int
	flag.IntVar(&key, "usr", 1, "User to login for testing")
	flag.Parse()

	slogger := slog.New(tint.NewHandler(os.Stderr, nil))

	// using it as initialization, if err occurs, we halt the application on startup rather than having issues while the
	// app is running
	if err := client.Init(key); err != nil {
		slogger.Error(err.Error())
		os.Exit(1)
	}

	f, err := tea.LogToFile("Letschat.log", "Letschat")

	slog.SetDefault(slog.New(slog.NewTextHandler(f, &slog.HandlerOptions{AddSource: true})))
	if err != nil {
		slogger.Error(err.Error())
		os.Exit(1)
	}
	defer f.Close()

	zone.NewGlobal()
	_ = lipgloss.DefaultRenderer().HasDarkBackground()
	_, err = tea.NewProgram(
		tui.InitialTabContainerModel(),
		tea.WithAltScreen(),
		tea.WithMouseAllMotion(),
		tea.WithoutBracketedPaste(),
		tea.WithReportFocus(),
	).Run()
	if err != nil {
		slogger.Error(err.Error())
		os.Exit(1)
	}
}
