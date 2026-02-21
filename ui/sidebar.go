package ui

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"otto/db"
)

type sidebarTablesMsg struct {
	tables []db.Table
}

type sidebarErrMsg struct {
	err error
}

type SidebarModel struct {
	db        db.DB
	cfg       db.Config
	tables    []db.Table
	filtered  []db.Table
	cursor    int
	width     int
	height    int
	focused   bool
	loading   bool
	searching bool
	query     string
}

func NewSidebarModel(d db.DB, cfg db.Config, width, height int) SidebarModel {
	return SidebarModel{
		db:      d,
		cfg:     cfg,
		width:   width,
		height:  height,
		loading: true,
		focused: true,
	}
}

func (m SidebarModel) loadTables() tea.Msg {
	tables, err := m.db.ListTables(context.Background())
	if err != nil {
		return sidebarErrMsg{err: err}
	}
	return sidebarTablesMsg{tables: tables}
}

func (m SidebarModel) Init() tea.Cmd {
	return m.loadTables
}

func (m *SidebarModel) applyFilter() {
	if m.query == "" {
		m.filtered = m.tables
		return
	}
	q := strings.ToLower(m.query)
	m.filtered = nil
	for _, t := range m.tables {
		if strings.Contains(strings.ToLower(t.Name), q) {
			m.filtered = append(m.filtered, t)
		}
	}
}

func (m SidebarModel) Update(msg tea.Msg) (SidebarModel, tea.Cmd) {
	switch msg := msg.(type) {
	case sidebarTablesMsg:
		m.tables = msg.tables
		m.filtered = msg.tables
		m.loading = false

	case sidebarErrMsg:
		m.loading = false

	case tea.KeyMsg:
		if !m.focused {
			return m, nil
		}

		if m.searching {
			switch msg.Type {
			case tea.KeyEsc:
				m.searching = false
				m.query = ""
				m.filtered = m.tables
				m.cursor = 0
			case tea.KeyBackspace:
				if len(m.query) > 0 {
					m.query = m.query[:len(m.query)-1]
					m.applyFilter()
					m.cursor = 0
				}
			case tea.KeyDown:
				if m.cursor < len(m.filtered)-1 {
					m.cursor++
				}
			case tea.KeyUp:
				if m.cursor > 0 {
					m.cursor--
				}
			case tea.KeyRunes:
				m.query += string(msg.Runes)
				m.applyFilter()
				m.cursor = 0
			}
			return m, nil
		}

		switch msg.String() {
		case "/":
			m.searching = true
			m.query = ""
			m.applyFilter()
			m.cursor = 0
		case "j", "down":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		}
	}
	return m, nil
}

func (m SidebarModel) SelectedTable() *db.Table {
	if len(m.filtered) == 0 || m.cursor >= len(m.filtered) {
		return nil
	}
	t := m.filtered[m.cursor]
	return &t
}

var (
	sidebarTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#8B949E"))

	sidebarItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#8B949E"))

	sidebarSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FF6F61"))

	sidebarFocusTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FF6F61"))

	sidebarSearchStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#E6EDF3"))

	sidebarSearchLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FF6F61"))

	sidebarNoMatchStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#555555")).
				Italic(true)
)

func (m SidebarModel) View() string {
	w := m.width
	if w < 6 {
		w = 6
	}
	h := m.height

	var lines []string

	titleSty := sidebarTitleStyle
	if m.focused {
		titleSty = sidebarFocusTitleStyle
	}
	lines = append(lines, titleSty.Render(" TABLES"))

	if m.searching {
		cursor := "█"
		if len(m.query) == 0 {
			cursor = "█"
		}
		searchLine := sidebarSearchLabelStyle.Render(" /") +
			sidebarSearchStyle.Render(m.query+cursor)
		lines = append(lines, searchLine)
	} else {
		hint := ""
		if m.focused && !m.loading {
			hint = " /"
		}
		lines = append(lines, strings.Repeat("─", w-len(hint))+
			lipgloss.NewStyle().Foreground(lipgloss.Color("#333344")).Render(hint))
	}

	list := m.filtered
	if m.loading {
		lines = append(lines, sidebarItemStyle.Render("  loading..."))
	} else if len(list) == 0 && m.query != "" {
		lines = append(lines, sidebarNoMatchStyle.Render("  no match"))
	} else if len(list) == 0 {
		lines = append(lines, sidebarItemStyle.Render("  no tables"))
	} else {
		contentH := h - 2
		if contentH < 1 {
			contentH = 1
		}
		start := 0
		if m.cursor >= contentH {
			start = m.cursor - contentH + 1
		}
		end := start + contentH
		if end > len(list) {
			end = len(list)
		}
		for i := start; i < end; i++ {
			t := list[i]
			name := t.Name
			if m.cfg.DBName == "" && t.Schema != "" {
				name = t.Schema + "." + t.Name
			}
			maxLen := w - 4
			if maxLen < 1 {
				maxLen = 1
			}
			runes := []rune(name)
			if len(runes) > maxLen {
				name = string(runes[:maxLen])
			}
			if i == m.cursor {
				lines = append(lines, sidebarSelectedStyle.Render(" ▶ "+name))
			} else {
				lines = append(lines, sidebarItemStyle.Render("   "+name))
			}
		}
	}

	for len(lines) < h {
		lines = append(lines, "")
	}
	if len(lines) > h {
		lines = lines[:h]
	}

	result := make([]string, len(lines))
	for i, l := range lines {
		vis := lipgloss.Width(l)
		if vis < w {
			l += strings.Repeat(" ", w-vis)
		}
		result[i] = l
	}

	return strings.Join(result, "\n")
}
