package menu

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Underline(true)
	itemStyle  = lipgloss.NewStyle().PaddingLeft(2)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("214")).Bold(true)
)

type Model struct {
	cursor   int
	items    []string
	choice   string
	quitting bool
}

func NewModel() Model {
	return Model{
		items: []string{"Wilson's Revenge", "WFC Land Creator", "WFC City Generator", "Tactical Colony", "Atlas Warlord", "Atlas Defense", "Exit"},
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter", " ":
			m.choice = m.items[m.cursor]
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\n\n  " + titleStyle.Render(" ATLAS GAMES ARCHIVE ") + "\n\n")

	for i, item := range m.items {
		if m.cursor == i {
			sb.WriteString(selectedItemStyle.Render("> " + item) + "\n")
		} else {
			sb.WriteString(itemStyle.Render("  " + item) + "\n")
		}
	}

	sb.WriteString("\n  [↑/↓] Navigate  [Enter] Select  [Q] Quit\n")
	return sb.String()
}

func (m Model) Choice() string {
	return m.choice
}
