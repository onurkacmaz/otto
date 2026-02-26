package ui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"otto/db"
)

const sidebarW = 26

type contentPane int

const (
	paneWelcome contentPane = iota
	paneTable
	paneEditor
)

type panelFocus int

const (
	focusSidebar panelFocus = iota
	focusContent
)

type GoBackToConnectMsg struct{}

type MainModel struct {
	db      db.DB
	cfg     db.Config
	sidebar SidebarModel
	table   TableModel
	editor  EditorModel
	content contentPane
	focus   panelFocus
	width   int
	height  int
}

func NewMainModel(d db.DB, cfg db.Config, width, height int) MainModel {
	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 24
	}
	_, ch := contentDims(width, height)
	return MainModel{
		db:      d,
		cfg:     cfg,
		sidebar: NewSidebarModel(d, cfg, sidebarW, ch),
		content: paneWelcome,
		focus:   focusSidebar,
		width:   width,
		height:  height,
	}
}

func contentDims(totalW, totalH int) (int, int) {
	w := totalW - sidebarW - 1
	h := totalH - 4
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return w, h
}

func (m MainModel) dims() (int, int) {
	return contentDims(m.width, m.height)
}

func (m MainModel) Init() tea.Cmd {
	return m.sidebar.Init()
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		cw, ch := m.dims()
		m.sidebar.width = sidebarW
		m.sidebar.height = ch
		m.table.width = cw
		m.table.height = ch
		m.editor.width = cw
		m.editor.height = ch
		if m.content == paneEditor {
			m.editor.textarea.SetWidth(cw - 2)
		}
		return m, nil

	case sidebarTablesMsg:
		var cmd tea.Cmd
		m.sidebar, cmd = m.sidebar.Update(msg)
		return m, cmd

	case GoBackMsg:
		m.focus = focusSidebar
		m.sidebar.focused = true
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.focus == focusSidebar {
				if m.sidebar.searching {
					var cmd tea.Cmd
					m.sidebar, cmd = m.sidebar.Update(msg)
					return m, cmd
				}
				if m.db != nil {
					m.db.Close(context.Background())
				}
				return m, func() tea.Msg { return GoBackToConnectMsg{} }
			}
		case "tab":
			if m.sidebar.searching {
				break
			}
			if m.focus == focusContent && m.content == paneEditor && m.editor.CompletionActive() {
				var cmd tea.Cmd
				m.editor, cmd = m.editor.Update(msg)
				return m, cmd
			}
			if m.focus == focusSidebar {
				if m.content != paneWelcome {
					m.focus = focusContent
					m.sidebar.focused = false
				}
			} else {
				m.focus = focusSidebar
				m.sidebar.focused = true
			}
			return m, nil
		case "enter":
			if m.focus == focusSidebar {
				m.sidebar.searching = false
				t := m.sidebar.SelectedTable()
				if t != nil {
					cw, ch := m.dims()
					m.table = NewTableModel(m.db, t.Schema, t.Name, cw, ch)
					m.content = paneTable
					m.focus = focusContent
					m.sidebar.focused = false
					return m, m.table.Init()
				}
				return m, nil
			}
		case "s":
			if m.focus == focusSidebar && !m.sidebar.searching {
				cw, ch := m.dims()
				m.editor = NewEditorModel(m.db, cw, ch)
				m.content = paneEditor
				m.focus = focusContent
				m.sidebar.focused = false
				return m, m.editor.Init()
			}
		}

		if m.focus == focusSidebar {
			m.sidebar.focused = true
			var cmd tea.Cmd
			m.sidebar, cmd = m.sidebar.Update(msg)
			return m, cmd
		}

		switch m.content {
		case paneTable:
			var cmd tea.Cmd
			m.table, cmd = m.table.Update(msg)
			return m, cmd
		case paneEditor:
			var cmd tea.Cmd
			m.editor, cmd = m.editor.Update(msg)
			return m, cmd
		}

	default:
		var cmds []tea.Cmd

		var sCmd tea.Cmd
		m.sidebar, sCmd = m.sidebar.Update(msg)
		if sCmd != nil {
			cmds = append(cmds, sCmd)
		}

		if m.content == paneTable {
			var tCmd tea.Cmd
			m.table, tCmd = m.table.Update(msg)
			if tCmd != nil {
				cmds = append(cmds, tCmd)
			}
		}

		if m.content == paneEditor {
			var eCmd tea.Cmd
			m.editor, eCmd = m.editor.Update(msg)
			if eCmd != nil {
				cmds = append(cmds, eCmd)
			}
		}

		if len(cmds) > 0 {
			return m, tea.Batch(cmds...)
		}
	}

	return m, nil
}

var (
	layoutAccent  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF6F61"))
	layoutMuted   = lipgloss.NewStyle().Foreground(lipgloss.Color("#8B949E"))
	layoutDivider = lipgloss.NewStyle().Foreground(lipgloss.Color("#30363D"))
	layoutFooter  = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	layoutSepNorm = lipgloss.NewStyle().Foreground(lipgloss.Color("#30363D"))
	layoutSepFoc  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6F61"))
)

func (m MainModel) renderHeader() string {
	driverIcon := "🐘"
	if m.cfg.Driver == db.DriverMySQL {
		driverIcon = "🐬"
	}
	dbName := m.cfg.DBName
	if dbName == "" {
		dbName = m.cfg.Host
	}
	left := layoutAccent.Render(" otto") +
		layoutMuted.Render("  ●  "+driverIcon+" "+dbName+" @ "+m.cfg.Host)
	right := layoutMuted.Render("[s] SQL  [Tab] Switch  [Esc] Disconnect  ")

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}

func (m MainModel) renderFooter() string {
	var hints string
	if m.focus == focusSidebar {
		if m.sidebar.searching {
			hints = "type to filter  ·  ↑↓ navigate  ·  Enter open  ·  Esc clear search"
		} else {
			hints = "↑↓ navigate  ·  Enter open  ·  / search  ·  s SQL  ·  Tab switch  ·  Esc disconnect"
		}
	} else {
		switch m.content {
		case paneTable:
			hints = "↑↓ rows  ·  ←→ scroll  ·  a/d column  ·  o sort  ·  u clear  ·  n/p page  ·  r refresh  ·  Tab sidebar  ·  Esc close"
		case paneEditor:
			if m.editor.mode == modeEditing {
				hints = "Ctrl+E run  ·  Ctrl+R editor↔results  ·  Tab sidebar  ·  Esc sidebar"
			} else {
				hints = "↑↓ rows  ·  ←→ scroll  ·  Ctrl+R editor↔results  ·  Tab sidebar  ·  Esc sidebar"
			}
		}
	}
	return layoutFooter.Render(" " + hints)
}

func (m MainModel) renderWelcome(w, h int) string {
	lines := make([]string, h)
	if h > 2 {
		msg := "Select a table from the sidebar  ·  [s] Open SQL editor"
		lines[h/2] = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#333344")).
			Width(w).
			Align(lipgloss.Center).
			Render(msg)
	}
	return strings.Join(lines, "\n")
}

func (m MainModel) View() string {
	w := m.width
	h := m.height
	if w == 0 {
		w = 80
	}
	if h == 0 {
		h = 24
	}

	header := m.renderHeader()
	footer := m.renderFooter()
	divider := layoutDivider.Render(strings.Repeat("─", w))

	cw, ch := m.dims()

	sb := m.sidebar
	sb.width = sidebarW
	sb.height = ch
	sidebarView := sb.View()

	sepSty := layoutSepNorm
	if m.focus == focusSidebar {
		sepSty = layoutSepFoc
	}
	sepLines := make([]string, ch)
	for i := range sepLines {
		sepLines[i] = sepSty.Render("│")
	}
	sep := strings.Join(sepLines, "\n")

	var content string
	switch m.content {
	case paneWelcome:
		content = m.renderWelcome(cw, ch)
	case paneTable:
		content = m.table.ViewPanel(cw, ch)
	case paneEditor:
		content = m.editor.ViewPanel(cw, ch)
	}

	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, sep, content)

	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s",
		header, divider, body, divider, footer)
}
