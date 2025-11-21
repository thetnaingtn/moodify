package main

import (
	"log"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct{}

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(teaMsg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := teaMsg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
	}
	return m, nil
}

func newModel() *model {
	return &model{}
}

func (m *model) View() string {
	return "Hello World!"
}

func main() {
	m := newModel()

	if _, err := tea.LogToFile(filepath.Join("./", "moodify.log"), ""); err != nil {
		panic(err)
	}

	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
