package main

import (
	"github.com/M0hammadUsman/letschat/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"log"
)

func main() {
	f, err := tea.LogToFile("tea.log", "log")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if _, err := tea.NewProgram(tui.InitialModel(), tea.WithAltScreen()).Run(); err != nil {
		log.Fatal(err)
	}
}
