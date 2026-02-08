package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type inputMode int

const (
	modeInput inputMode = iota
	modeSelect
)

type listItem struct {
	title string
	value string
}

func (i listItem) Title() string       { return i.title }
func (i listItem) Description() string { return "" }
func (i listItem) FilterValue() string { return i.title }

type inputModel struct {
	title      string
	textInput  textinput.Model
	list       list.Model
	mode       inputMode
	options    []SelectOption
	quitting   bool
	canceled   bool
	useOptions bool
}

func (m inputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c", "esc"))):
			m.canceled = true
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
			if m.useOptions {
				if m.mode == modeInput {
					m.mode = modeSelect
				} else {
					m.mode = modeInput
				}
			}
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			m.quitting = true
			return m, tea.Quit
		}
	}

	if m.mode == modeInput {
		m.textInput, cmd = m.textInput.Update(msg)
	} else {
		m.list, cmd = m.list.Update(msg)
	}

	return m, cmd
}

func (m inputModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render("? ") + m.title + "\n\n")

	if m.useOptions {
		if m.mode == modeInput {
			b.WriteString("  " + selectedStyle.Render("[Manual Input]") + "\n")
			b.WriteString("  " + m.textInput.View() + "\n\n")
			b.WriteString("  " + disabledStyle.Render("─────────────────────") + "\n")
			b.WriteString("  " + Muted.Sprint("(Press Tab to select from origin)") + "\n")
		} else {
			b.WriteString("  " + disabledStyle.Render("[Manual Input]") + "\n")
			b.WriteString("  " + disabledStyle.Render(m.textInput.Value()) + "\n\n")
			b.WriteString("  " + selectedStyle.Render("─────────────────────") + "\n")
			b.WriteString("  " + selectedStyle.Render("Select from origin:") + "\n\n")
			b.WriteString(m.list.View())
			b.WriteString("\n  " + Muted.Sprint("(Press Tab to switch to manual input)") + "\n")
		}
	} else {
		b.WriteString("  " + m.textInput.View() + "\n")
	}

	return b.String()
}

// Input shows an interactive text input
func Input(title, placeholder string) (string, error) {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 40

	m := inputModel{
		title:      title,
		textInput:  ti,
		mode:       modeInput,
		useOptions: false,
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("failed to run input: %w", err)
	}

	result := finalModel.(inputModel)
	if result.canceled {
		return "", fmt.Errorf("input canceled")
	}

	return strings.TrimSpace(result.textInput.Value()), nil
}

// InputWithOptions shows an interactive input with selectable options
func InputWithOptions(title, placeholder string, options []SelectOption) (string, error) {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 40

	// Convert options to list items
	items := make([]list.Item, len(options))
	for i, opt := range options {
		items[i] = listItem{title: opt.Label, value: opt.Value}
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = selectedStyle
	delegate.Styles.SelectedDesc = selectedStyle
	delegate.SetHeight(1)

	l := list.New(items, delegate, 50, 10)
	l.Title = ""
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.SetShowTitle(false)
	l.Styles.Title = lipgloss.NewStyle()

	m := inputModel{
		title:      title,
		textInput:  ti,
		list:       l,
		mode:       modeInput,
		options:    options,
		useOptions: true,
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("failed to run input: %w", err)
	}

	result := finalModel.(inputModel)
	if result.canceled {
		return "", fmt.Errorf("input canceled")
	}

	if result.mode == modeInput {
		return strings.TrimSpace(result.textInput.Value()), nil
	}

	// Get selected item from list
	selected := result.list.SelectedItem()
	if selected != nil {
		if item, ok := selected.(listItem); ok {
			return item.value, nil
		}
	}

	return "", nil
}
