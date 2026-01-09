package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/donghun/wm/internal/config"
	"github.com/donghun/wm/internal/detect"
)

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

type InitModel struct {
	step       int
	inputs     []textinput.Model
	focusIndex int
	syncFiles  []string
	detection  detect.DetectionResult
	repoName   string
	repoRoot   string
	done       bool
	result     *config.Config
}

func NewInitModel(repoRoot, repoName string, detection detect.DetectionResult) InitModel {
	m := InitModel{
		step:      0,
		inputs:    make([]textinput.Model, 2),
		detection: detection,
		repoName:  repoName,
		repoRoot:  repoRoot,
		syncFiles: []string{".env"},
	}

	// Base directory input
	var t textinput.Model
	t = textinput.New()
	t.Placeholder = "../wm_{repo}"
	t.SetValue("../wm_{repo}")
	t.Focus()
	t.CharLimit = 256
	t.Width = 40
	m.inputs[0] = t

	// Sync files input
	t = textinput.New()
	t.Placeholder = ".env, apps/*/.env"
	t.SetValue(".env")
	t.CharLimit = 512
	t.Width = 40
	m.inputs[1] = t

	return m
}

func (m InitModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m InitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "tab", "shift+tab", "up", "down":
			s := msg.String()
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex >= len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs) - 1
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := range m.inputs {
				if i == m.focusIndex {
					cmds[i] = m.inputs[i].Focus()
				} else {
					m.inputs[i].Blur()
				}
			}

			return m, tea.Batch(cmds...)

		case "enter":
			m.done = true
			m.result = m.buildConfig()
			return m, tea.Quit
		}
	}

	// Handle input updates
	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m *InitModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

func (m InitModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("WM Init"))
	b.WriteString("\n\n")

	// Detection info
	if m.detection.PackageManager != "" {
		b.WriteString(fmt.Sprintf("Detected: %s", m.detection.PackageManager))
		if m.detection.IsMonorepo {
			b.WriteString(" (monorepo)")
		}
		b.WriteString("\n\n")
	}

	// Base directory
	b.WriteString("Worktree base directory:\n")
	b.WriteString(m.inputs[0].View())
	b.WriteString("\n\n")

	// Sync files
	b.WriteString("Files to sync (comma-separated):\n")
	b.WriteString(m.inputs[1].View())
	b.WriteString("\n\n")

	// Help
	b.WriteString(helpStyle.Render("tab: next field | enter: save | esc: cancel"))

	return b.String()
}

func (m InitModel) buildConfig() *config.Config {
	cfg := config.NewConfig()
	cfg.Worktree.BaseDir = m.inputs[0].Value()

	// Parse sync files
	syncStr := m.inputs[1].Value()
	if syncStr != "" {
		parts := strings.Split(syncStr, ",")
		cfg.Sync = make([]config.SyncItem, len(parts))
		for i, part := range parts {
			path := strings.TrimSpace(part)
			cfg.Sync[i] = config.SyncItem{
				Src:  path,
				Dst:  path,
				Mode: "copy",
				When: "always",
			}
		}
	}

	// Add detected install command
	if m.detection.InstallCommand != "" {
		cfg.Tasks.PostInstall = config.PostInstallConfig{
			Mode:     "background",
			Commands: []string{m.detection.InstallCommand},
		}
	}

	return cfg
}

func (m InitModel) Result() *config.Config {
	return m.result
}

func (m InitModel) Done() bool {
	return m.done
}

// RunInitTUI runs the init TUI and returns the config
func RunInitTUI(repoRoot, repoName string, detection detect.DetectionResult) (*config.Config, error) {
	model := NewInitModel(repoRoot, repoName, detection)
	p := tea.NewProgram(model)

	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	m := finalModel.(InitModel)
	if !m.Done() {
		return nil, fmt.Errorf("cancelled")
	}

	return m.Result(), nil
}
