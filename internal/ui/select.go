package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Bold(true)
	cursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))
	disabledStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	titleStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true)
)

type SelectOption struct {
	Label    string
	Value    string
	Disabled bool
}

type selectModel struct {
	title    string
	options  []SelectOption
	cursor   int
	selected int
	quitting bool
	canceled bool
}

func (m selectModel) Init() tea.Cmd {
	return nil
}

func (m selectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c", "esc"))):
			m.canceled = true
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.options) - 1
			}
			// Skip disabled items
			checked := 0
			for m.options[m.cursor].Disabled {
				m.cursor--
				if m.cursor < 0 {
					m.cursor = len(m.options) - 1
				}
				checked++
				if checked >= len(m.options) {
					break // All options are disabled
				}
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			m.cursor++
			if m.cursor >= len(m.options) {
				m.cursor = 0
			}
			// Skip disabled items
			checked := 0
			for m.options[m.cursor].Disabled {
				m.cursor++
				if m.cursor >= len(m.options) {
					m.cursor = 0
				}
				checked++
				if checked >= len(m.options) {
					break // All options are disabled
				}
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if !m.options[m.cursor].Disabled {
				m.selected = m.cursor
				m.quitting = true
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m selectModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render("? ") + m.title + "\n\n")

	for i, option := range m.options {
		cursor := "  "
		if m.cursor == i {
			cursor = cursorStyle.Render("❯ ")
		}

		label := option.Label
		if option.Disabled {
			label = disabledStyle.Render(label)
		} else if m.cursor == i {
			label = selectedStyle.Render(label)
		}

		b.WriteString(cursor + label + "\n")
	}

	return b.String()
}

// Select shows an interactive selection menu and returns the selected value
func Select(title string, options []SelectOption) (string, error) {
	// Find first non-disabled option
	cursor := 0
	for i, opt := range options {
		if !opt.Disabled {
			cursor = i
			break
		}
	}

	m := selectModel{
		title:   title,
		options: options,
		cursor:  cursor,
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("failed to run select: %w", err)
	}

	result := finalModel.(selectModel)
	if result.canceled {
		return "", fmt.Errorf("selection canceled")
	}

	return result.options[result.selected].Value, nil
}
