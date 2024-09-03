package main

import (
	"github.com/M0hammadUsman/letschat/internal/client"
	"github.com/M0hammadUsman/letschat/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
	"log"
)

func main() {
	f, err := tea.LogToFile("tea.log", "log")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	// using it as initialization if err occurs we halt the application on startup rather than having issues while the
	// app is running
	if err = client.Init(); err != nil {
		log.Fatal(err)
	}
	zone.NewGlobal()
	_, err = tea.NewProgram(tui.InitialTabContainerModel(), tea.WithAltScreen(), tea.WithMouseAllMotion()).Run()
	if err != nil {
		log.Fatal(err)
	}
}
