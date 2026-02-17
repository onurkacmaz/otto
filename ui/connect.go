package ui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"otto/db"
)

const (
	fieldDriver = iota
	fieldHost
	fieldPort
	fieldUser
	fieldPassword
	fieldDBName
	fieldCount
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF6F61")).
			MarginBottom(1)

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA"))

	activeInputStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF6F61")).
				Bold(true)

	buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000")).
			Background(lipgloss.Color("#FF6F61")).
			Padding(0, 3).
			MarginTop(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			MarginTop(1)

	historySelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF6F61")).
				Bold(true)

	historyItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#AAAAAA"))
)

type ConnectedMsg struct {
	DB  db.DB
	Cfg db.Config
}

type connectErrMsg struct {
	err error
}

type ConnectModel struct {
	inputs          []textinput.Model
	focused         int
	err             error
	connecting      bool
	width           int
	height          int
	history         []db.Config
	selectedHistory int
	showHistory     bool
	driver          db.Driver
}

func NewConnectModel() ConnectModel {
	inputs := make([]textinput.Model, fieldCount)

	for i := range inputs {
		t := textinput.New()
		t.CharLimit = 64

		switch i {
		case fieldDriver:
			t.Placeholder = "postgres"
			t.Focus()
		case fieldHost:
			t.Placeholder = "localhost"
		case fieldPort:
			t.Placeholder = "5432"
		case fieldUser:
			t.Placeholder = "postgres"
		case fieldPassword:
			t.Placeholder = "••••••••"
			t.EchoMode = textinput.EchoPassword
		case fieldDBName:
			t.Placeholder = "mydb"
		}

		inputs[i] = t
	}

	history := db.LoadHistory()

	return ConnectModel{
		inputs:          inputs,
		focused:         0,
		history:         history,
		selectedHistory: -1,
		showHistory:     len(history) > 0,
		driver:          db.DriverPostgres,
	}
}

func (m ConnectModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *ConnectModel) applyDriverDefaults() {
	if m.driver == db.DriverMySQL {
		m.inputs[fieldPort].Placeholder = "3306"
		m.inputs[fieldUser].Placeholder = "root"
	} else {
		m.inputs[fieldPort].Placeholder = "5432"
		m.inputs[fieldUser].Placeholder = "postgres"
	}
}

func (m *ConnectModel) toggleDriver() {
	if m.driver == db.DriverPostgres {
		m.driver = db.DriverMySQL
		m.inputs[fieldDriver].SetValue("mysql")
	} else {
		m.driver = db.DriverPostgres
		m.inputs[fieldDriver].SetValue("postgres")
	}
	m.applyDriverDefaults()
}

func (m *ConnectModel) connectToHistory(idx int) tea.Cmd {
	if idx < 0 || idx >= len(m.history) {
		return nil
	}
	cfg := m.history[idx]
	m.driver = cfg.Driver
	if m.driver == "" {
		m.driver = db.DriverPostgres
	}
	m.inputs[fieldDriver].SetValue(string(m.driver))
	m.inputs[fieldHost].SetValue(cfg.Host)
	m.inputs[fieldPort].SetValue(cfg.Port)
	m.inputs[fieldUser].SetValue(cfg.User)
	m.inputs[fieldPassword].SetValue(cfg.Password)
	m.inputs[fieldDBName].SetValue(cfg.DBName)
	m.showHistory = false
	m.connecting = true
	m.err = nil
	return func() tea.Msg {
		conn, err := db.Connect(context.Background(), cfg)
		if err != nil {
			return connectErrMsg{err: err}
		}
		return ConnectedMsg{DB: conn, Cfg: cfg}
	}
}

func (m ConnectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		if m.showHistory {
			switch msg.Type {
			case tea.KeyDown:
				if m.selectedHistory < len(m.history)-1 {
					m.selectedHistory++
				}
				return m, nil
			case tea.KeyUp:
				if m.selectedHistory > 0 {
					m.selectedHistory--
				}
				return m, nil
			case tea.KeyEnter:
				if m.selectedHistory >= 0 {
					cmd := m.connectToHistory(m.selectedHistory)
					return m, cmd
				}
				return m, nil
			case tea.KeyTab:
				m.showHistory = false
				return m, nil
			case tea.KeyEsc:
				m.showHistory = false
				return m, nil
			}
			if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 {
				r := msg.Runes[0]
				if r >= '1' && r <= '9' {
					idx := int(r - '1')
					if idx < len(m.history) {
						cmd := m.connectToHistory(idx)
						return m, cmd
					}
					return m, nil
				}
			}
			return m, nil
		}

		switch msg.Type {
		case tea.KeyDown:
			m.inputs[m.focused].Blur()
			m.focused = (m.focused + 1) % fieldCount
			m.inputs[m.focused].Focus()
			return m, nil

		case tea.KeyUp:
			m.inputs[m.focused].Blur()
			m.focused = (m.focused - 1 + fieldCount) % fieldCount
			m.inputs[m.focused].Focus()
			return m, nil

		case tea.KeyTab:
			if m.focused == fieldDriver {
				m.toggleDriver()
				return m, nil
			}
			if len(m.history) > 0 {
				m.showHistory = true
				if m.selectedHistory < 0 {
					m.selectedHistory = 0
				}
				return m, nil
			}

		case tea.KeyEnter:
			if m.connecting {
				return m, nil
			}
			m.connecting = true
			m.err = nil
			cfg := db.Config{
				Driver:   m.driver,
				Host:     m.inputs[fieldHost].Value(),
				Port:     m.inputs[fieldPort].Value(),
				User:     m.inputs[fieldUser].Value(),
				Password: m.inputs[fieldPassword].Value(),
				DBName:   m.inputs[fieldDBName].Value(),
			}
			return m, func() tea.Msg {
				conn, err := db.Connect(context.Background(), cfg)
				if err != nil {
					return connectErrMsg{err: err}
				}
				return ConnectedMsg{DB: conn, Cfg: cfg}
			}
		}

	case connectErrMsg:
		m.err = msg.err
		m.connecting = false
		return m, nil
	}

	if m.focused == fieldDriver {
		return m, nil
	}

	var cmd tea.Cmd
	m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
	return m, cmd
}

func (m ConnectModel) View() string {
	labels := []string{"Driver", "Host", "Port", "User", "Password", "Database"}

	driverIcon := "🐘"
	driverName := "PostgreSQL"
	if m.driver == db.DriverMySQL {
		driverIcon = "🐬"
		driverName = "MySQL"
	}
	s := titleStyle.Render(driverIcon+" "+driverName+" Connection") + "\n\n"

	if m.showHistory && len(m.history) > 0 {
		s += lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FAFAFA")).Render("Recent Connections") + "\n"
		max := len(m.history)
		if max > 9 {
			max = 9
		}
		for i := 0; i < max; i++ {
			name := db.DisplayName(m.history[i])
			prefix := fmt.Sprintf(" %d ", i+1)
			if i == m.selectedHistory {
				s += historySelectedStyle.Render("▸"+prefix+name) + "\n"
			} else {
				s += historyItemStyle.Render(" "+prefix+name) + "\n"
			}
		}
		s += "\n" + helpStyle.Render("↑/↓: select • Enter/1-9: connect • Tab/Esc: back to form")

		w := m.width
		h := m.height
		if w == 0 {
			w = 80
		}
		if h == 0 {
			h = 24
		}
		return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, s)
	}

	for i, input := range m.inputs {
		label := labels[i]
		if i == fieldDriver {
			driverVal := "postgres"
			if m.driver == db.DriverMySQL {
				driverVal = "mysql"
			}
			if i == m.focused {
				s += activeInputStyle.Render("▸ "+label+": "+driverVal+" (Tab to switch)") + "\n"
			} else {
				s += inputStyle.Render("  "+label+": "+driverVal) + "\n"
			}
			continue
		}
		if i == m.focused {
			s += activeInputStyle.Render("▸ "+label+": ") + input.View() + "\n"
		} else {
			s += inputStyle.Render("  "+label+": ") + input.View() + "\n"
		}
	}

	if m.err != nil {
		errMsg := m.err.Error()
		maxErrLen := 60
		if len(errMsg) > maxErrLen {
			errMsg = errMsg[:maxErrLen] + "..."
		}
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
		s += "\n" + errStyle.Render("Error: "+errMsg) + "\n"
	} else if m.connecting {
		s += "\n" + helpStyle.Render("Connecting...") + "\n"
	} else {
		s += "\n" + buttonStyle.Render("Connect") + "\n"
	}

	helpText := "↑/↓: navigate • Enter: connect • Ctrl+C: quit"
	if len(m.history) > 0 {
		helpText += " • Tab: history"
	}
	s += helpStyle.Render(helpText)

	w := m.width
	h := m.height
	if w == 0 {
		w = 80
	}
	if h == 0 {
		h = 24
	}
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, s)
}
