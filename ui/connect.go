package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"otto/db"
)

const (
	fieldName = iota
	fieldDriver
	fieldHost
	fieldPort
	fieldUser
	fieldPassword
	fieldDBName
	fieldCount
)

var (
	accentColor = lipgloss.Color("#FF6F61")
	mutedColor  = lipgloss.Color("#555566")
	dimColor    = lipgloss.Color("#30363D")
	textColor   = lipgloss.Color("#DADADA")
	darkColor   = lipgloss.Color("#0D1117")
	greenColor  = lipgloss.Color("#3FB950")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accentColor).
			MarginBottom(1)

	inputStyle = lipgloss.NewStyle().
			Foreground(textColor)

	activeInputStyle = lipgloss.NewStyle().
				Foreground(accentColor).
				Bold(true)

	buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000")).
			Background(accentColor).
			Padding(0, 3).
			MarginTop(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			MarginTop(1)

	historySelectedStyle = lipgloss.NewStyle().
				Foreground(accentColor).
				Bold(true)

	historyItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#8B949E"))
)

var (
	panelW         = 54
	histPanelW     = 54
	sideBySideMinW = 116

	cHeaderTitle = lipgloss.NewStyle().Bold(true).Foreground(accentColor)
	cHeaderDB    = lipgloss.NewStyle().Foreground(textColor).Bold(true)
	cHeaderSep   = lipgloss.NewStyle().Foreground(dimColor)

	driverOnStyle = lipgloss.NewStyle().
			Background(accentColor).
			Foreground(darkColor).
			Bold(true).
			Padding(0, 1)

	driverOffStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Padding(0, 1)

	driverHintStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)

	fieldLabelActive = lipgloss.NewStyle().
				Foreground(accentColor).
				Bold(true).
				Width(9)

	fieldLabelInactive = lipgloss.NewStyle().
				Foreground(mutedColor).
				Width(9)

	fieldArrow = lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	fieldGap   = lipgloss.NewStyle().Foreground(dimColor)

	btnConnect = lipgloss.NewStyle().
			Background(accentColor).
			Foreground(darkColor).
			Bold(true).
			Padding(0, 3)

	btnConnecting = lipgloss.NewStyle().
			Foreground(mutedColor).
			Padding(0, 3)

	errStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444"))

	cHelpStyle = lipgloss.NewStyle().Foreground(mutedColor)

	histTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(textColor)

	histNumStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true).
			Width(2)

	histNumMutedStyle = lipgloss.NewStyle().
				Foreground(mutedColor).
				Width(2)

	histTagPGStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#1C4E80")).
			Foreground(lipgloss.Color("#79C0FF")).
			Padding(0, 1)

	histTagMYStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#3B2300")).
			Foreground(lipgloss.Color("#F7A663")).
			Padding(0, 1)

	histNameActiveStyle = lipgloss.NewStyle().Foreground(textColor).Bold(true)
	histNameStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#8B949E"))

	histHelpStyle = lipgloss.NewStyle().Foreground(mutedColor)
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
	historyFocused  bool
	driver          db.Driver
	editingIndex    int
}

func NewConnectModel() ConnectModel {
	inputs := make([]textinput.Model, fieldCount)

	for i := range inputs {
		t := textinput.New()
		t.CharLimit = 64
		t.PromptStyle = lipgloss.NewStyle()
		t.TextStyle = lipgloss.NewStyle().Foreground(textColor)
		t.Prompt = ""

		switch i {
		case fieldName:
			t.Placeholder = "My Production DB  (optional)"
			t.Focus()
		case fieldDriver:
			t.Placeholder = "postgres"
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
		focused:         fieldName,
		history:         history,
		selectedHistory: -1,
		historyFocused:  len(history) > 0,
		driver:          db.DriverPostgres,
		editingIndex:    -1,
	}
}

func (m ConnectModel) Init() tea.Cmd { return textinput.Blink }

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
	m.inputs[fieldName].SetValue(cfg.Name)
	m.inputs[fieldDriver].SetValue(string(m.driver))
	m.inputs[fieldHost].SetValue(cfg.Host)
	m.inputs[fieldPort].SetValue(cfg.Port)
	m.inputs[fieldUser].SetValue(cfg.User)
	m.inputs[fieldPassword].SetValue(cfg.Password)
	m.inputs[fieldDBName].SetValue(cfg.DBName)
	m.historyFocused = false
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
		if m.historyFocused {
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
					return m, m.connectToHistory(m.selectedHistory)
				}
				return m, nil
			case tea.KeyTab, tea.KeyEsc:
				m.historyFocused = false
				return m, nil
			}
			if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'd' {
				if m.selectedHistory >= 0 && m.selectedHistory < len(m.history) {
					db.DeleteConnection(m.history[m.selectedHistory])
					m.history = db.LoadHistory()
					if len(m.history) == 0 {
						m.historyFocused = false
						m.selectedHistory = -1
					} else if m.selectedHistory >= len(m.history) {
						m.selectedHistory = len(m.history) - 1
					}
				}
				return m, nil
			}
			if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 {
				r := msg.Runes[0]
				if r >= '1' && r <= '9' {
					idx := int(r - '1')
					if idx < len(m.history) {
						return m, m.connectToHistory(idx)
					}
				}
			}
			if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'e' {
				if m.selectedHistory >= 0 && m.selectedHistory < len(m.history) {
					cfg := m.history[m.selectedHistory]
					m.editingIndex = m.selectedHistory
					m.driver = cfg.Driver
					if m.driver == "" {
						m.driver = db.DriverPostgres
					}
					m.applyDriverDefaults()
					m.inputs[fieldName].SetValue(cfg.Name)
					m.inputs[fieldDriver].SetValue(string(m.driver))
					m.inputs[fieldHost].SetValue(cfg.Host)
					m.inputs[fieldPort].SetValue(cfg.Port)
					m.inputs[fieldUser].SetValue(cfg.User)
					m.inputs[fieldPassword].SetValue(cfg.Password)
					m.inputs[fieldDBName].SetValue(cfg.DBName)
					m.historyFocused = false
					m.focused = fieldName
					for i := range m.inputs {
						m.inputs[i].Blur()
					}
					m.inputs[fieldName].Focus()
				}
				return m, nil
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
			if m.focused == fieldName {
				m.inputs[m.focused].Blur()
				m.focused = fieldDriver
				m.inputs[m.focused].Focus()
				return m, nil
			}
			if len(m.history) > 0 {
				m.historyFocused = true
				m.editingIndex = -1
				if m.selectedHistory < 0 {
					m.selectedHistory = 0
				}
			}
			return m, nil
		case tea.KeyEnter:
			if m.connecting {
				return m, nil
			}
			m.connecting = true
			m.err = nil
			cfg := db.Config{
				Name:     m.inputs[fieldName].Value(),
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

	if m.focused != fieldDriver {
		var cmd tea.Cmd
		m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m ConnectModel) labelWidth() int { return 10 }

func (m ConnectModel) View() string {
	w, h := m.width, m.height
	if w == 0 {
		w = 80
	}
	if h == 0 {
		h = 24
	}
	
	if len(m.history) > 0 && w >= sideBySideMinW {
		form := m.renderForm()
		hist := m.renderHistory()
		panels := lipgloss.JoinHorizontal(lipgloss.Top, form, "  ", hist)
		return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, panels)
	}
	
	if m.historyFocused && len(m.history) > 0 {
		return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, m.renderHistory())
	}
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, m.renderForm())
}

func (m ConnectModel) renderForm() string {
	icon, dbName := "🐘", "PostgreSQL"
	if m.driver == db.DriverMySQL {
		icon, dbName = "🐬", "MySQL"
	}

	sep := lipgloss.NewStyle().Foreground(dimColor).Render(strings.Repeat("─", panelW-4))

	titleText := icon + "  otto"
	if m.editingIndex >= 0 {
		titleText = icon + "  otto  ✎ Edit"
	}
	leftH := cHeaderTitle.Render(titleText)
	rightH := cHeaderDB.Render(dbName)
	gap := (panelW - 4) - lipgloss.Width(leftH) - lipgloss.Width(rightH)
	if gap < 1 {
		gap = 1
	}
	header := leftH + strings.Repeat(" ", gap) + rightH

	labels := []string{"Name", "Driver", "Host", "Port", "User", "Password", "Database"}
	var rows []string

	for i, inp := range m.inputs {
		active := i == m.focused
		label := labels[i]

		lbl := fieldLabelInactive.Render(label)
		if active {
			lbl = fieldLabelActive.Render(label)
		}

		var val string
		if i == fieldDriver {
			pgS, myS := driverOffStyle, driverOffStyle
			if m.driver == db.DriverPostgres {
				pgS = driverOnStyle
			} else {
				myS = driverOnStyle
			}
			hint := driverHintStyle.Render("Tab to switch")
			val = pgS.Render("postgres") +
				fieldGap.Render("  ·  ") +
				myS.Render("mysql") +
				"  " + hint
		} else {
			val = inp.View()
		}

		var row string
		if active {
			row = fieldArrow.Render("▸") + " " + lbl + "  " + val
		} else {
			row = "  " + lbl + "  " + val
		}

		rows = append(rows, row)
		if i == fieldName {
			rows = append(rows, sep)
		}
	}
	fields := strings.Join(rows, "\n")

	var status string
	switch {
	case m.err != nil:
		msg := m.err.Error()
		if len(msg) > panelW-4 {
			msg = msg[:panelW-4] + "…"
		}
		status = errStyle.Render("✕  " + msg)
	case m.connecting:
		status = btnConnecting.Render("⟳  Connecting…")
	default:
		btnText := "  Connect  "
		if m.editingIndex >= 0 {
			btnText = " Save & Connect "
		}
		btn := btnConnect.Render(btnText)
		bw := lipgloss.Width(btn)
		pad := (panelW - bw) / 2
		if pad < 0 {
			pad = 0
		}
		status = strings.Repeat(" ", pad) + btn
	}

	hint := cHelpStyle.Render("↑↓ navigate · Enter connect · Ctrl+C quit")

	inner := lipgloss.JoinVertical(lipgloss.Left,
		header,
		sep,
		fields,
		"",
		status,
		"",
		hint,
	)

	borderColor := accentColor
	if m.historyFocused {
		borderColor = dimColor
	}
	pStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 2)
	return pStyle.Width(panelW).Render(inner)
}

func (m ConnectModel) renderHistory() string {
	title := histTitleStyle.Render("Recent Connections")
	sep := lipgloss.NewStyle().Foreground(dimColor).Render(strings.Repeat("─", histPanelW-4))

	max := len(m.history)
	if max > 9 {
		max = 9
	}

	var rows []string
	for i := 0; i < max; i++ {
		cfg := m.history[i]
		active := i == m.selectedHistory

		var num string
		if active {
			num = histNumStyle.Render(fmt.Sprintf("%d", i+1))
		} else {
			num = histNumMutedStyle.Render(fmt.Sprintf("%d", i+1))
		}

		var tag string
		if cfg.Driver == db.DriverMySQL {
			tag = histTagMYStyle.Render("my")
		} else {
			tag = histTagPGStyle.Render("pg")
		}

		name := db.DisplayName(cfg)
		if len(name) > histPanelW-14 {
			name = name[:histPanelW-17] + "…"
		}
		var nameStr string
		if active {
			nameStr = histNameActiveStyle.Render(name)
		} else {
			nameStr = histNameStyle.Render(name)
		}

		if active {
			rows = append(rows, fieldArrow.Render("▸")+" "+num+"  "+tag+"  "+nameStr)
		} else {
			rows = append(rows, "  "+num+"  "+tag+"  "+nameStr)
		}
	}

	list := strings.Join(rows, "\n")
	hint := histHelpStyle.Render("↑↓ select · 1-9/Enter · d delete · Esc · e Edit")

	inner := lipgloss.JoinVertical(lipgloss.Left,
		title,
		sep,
		"",
		list,
		"",
		hint,
	)

	borderColor := accentColor
	if !m.historyFocused {
		borderColor = dimColor
	}
	hStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 2)
	return hStyle.Width(histPanelW).Render(inner)
}
