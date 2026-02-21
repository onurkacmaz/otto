package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"otto/db"
)

type editorFocus int

const (
	modeEditing editorFocus = iota
	modeResults
)

type editorMode = editorFocus

type queryResultMsg struct {
	result  *db.QueryResult
	elapsed time.Duration
}

type queryErrMsg struct {
	err error
}

type EditorModel struct {
	db        db.DB
	textarea  textarea.Model
	mode      editorFocus
	result    *db.QueryResult
	elapsed   time.Duration
	err       error
	running   bool
	cursor    int
	scrollX   int
	width     int
	height    int
	colWidths []int
}

func NewEditorModel(d db.DB, width, height int) EditorModel {
	editorH := editorHeight(height)
	ta := textarea.New()
	ta.Placeholder = "SELECT * FROM ..."
	placeholderSty := lipgloss.NewStyle().Foreground(lipgloss.Color("#555577"))
	ta.FocusedStyle.Placeholder = placeholderSty
	ta.BlurredStyle.Placeholder = placeholderSty
	ta.SetWidth(width - 4)
	ta.SetHeight(editorH - 2)
	ta.Focus()
	ta.ShowLineNumbers = true

	return EditorModel{
		db:       d,
		textarea: ta,
		mode:     modeEditing,
		width:    width,
		height:   height,
	}
}

func editorHeight(totalH int) int {
	h := totalH * 2 / 5
	if h < 6 {
		h = 6
	}
	if h > 18 {
		h = 18
	}
	return h
}

func (m EditorModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m *EditorModel) calcColWidths() {
	if m.result == nil {
		return
	}
	m.colWidths = make([]int, len(m.result.Columns))
	for i, col := range m.result.Columns {
		m.colWidths[i] = len([]rune(col))
	}
	for _, row := range m.result.Rows {
		for i, val := range row {
			if idx := strings.IndexAny(val, "\n\r"); idx >= 0 {
				val = val[:idx]
			}
			if l := len([]rune(val)); l > m.colWidths[i] {
				m.colWidths[i] = l
			}
			if m.colWidths[i] > 30 {
				m.colWidths[i] = 30
			}
		}
	}
}

func (m EditorModel) execQuery() tea.Msg {
	query := strings.TrimSpace(m.textarea.Value())
	if query == "" {
		return queryErrMsg{err: fmt.Errorf("empty query")}
	}
	start := time.Now()
	result, err := m.db.ExecQuery(context.Background(), query)
	if err != nil {
		return queryErrMsg{err: err}
	}
	return queryResultMsg{result: result, elapsed: time.Since(start)}
}

func (m EditorModel) Update(msg tea.Msg) (EditorModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		edH := editorHeight(msg.Height)
		m.textarea.SetWidth(msg.Width - 4)
		m.textarea.SetHeight(edH - 2)

	case queryResultMsg:
		m.result = msg.result
		m.elapsed = msg.elapsed
		m.err = nil
		m.cursor = 0
		m.scrollX = 0
		m.running = false
		m.mode = modeResults
		m.calcColWidths()

	case queryErrMsg:
		m.err = msg.err
		m.result = nil
		m.running = false
		m.mode = modeResults

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+e":
			if !m.running {
				m.running = true
				return m, m.execQuery
			}
			return m, nil
		case "esc":
			return m, func() tea.Msg { return GoBackMsg{} }
		case "ctrl+r":
			if m.mode == modeEditing {
				m.mode = modeResults
				m.textarea.Blur()
			} else {
				m.mode = modeEditing
				m.textarea.Focus()
				return m, textarea.Blink
			}
			return m, nil
		}

		if m.mode == modeEditing {
			var cmd tea.Cmd
			m.textarea, cmd = m.textarea.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "j", "down":
			if m.result != nil && m.cursor < len(m.result.Rows)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "l", "right":
			m.scrollX += 5
		case "h", "left":
			if m.scrollX >= 5 {
				m.scrollX -= 5
			} else {
				m.scrollX = 0
			}
		}

	default:
		if m.mode == modeEditing {
			var cmd tea.Cmd
			m.textarea, cmd = m.textarea.Update(msg)
			return m, cmd
		}
	}
	return m, nil
}

var (
	edBorderActive = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF6F61"))

	edBorderInactive = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#30363D"))

	edLabelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF6F61"))

	edLabelInactiveStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#8B949E"))

	edStatusOk  = lipgloss.NewStyle().Foreground(lipgloss.Color("#3FB950"))
	edStatusErr = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555"))
	edStatusRun = lipgloss.NewStyle().Foreground(lipgloss.Color("#E3B341"))
	edHintStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#444455"))
)

func (m EditorModel) ViewPanel(w, h int) string {
	edH := editorHeight(h)
	innerW := w - 2

	m.textarea.SetWidth(innerW - 2)
	m.textarea.SetHeight(edH - 2)

	hint := edHintStyle.Render("Ctrl+E run  ·  Ctrl+R editor↔results  ·  Tab sidebar")
	hintW := lipgloss.Width(hint)
	label := " SQL Editor "
	gap := innerW - len(label) - hintW
	if gap < 0 {
		gap = 0
	}
	editorTitle := edLabelStyle.Render(label) +
		strings.Repeat(" ", gap) +
		hint

	editorInner := editorTitle + "\n" + m.textarea.View()

	edBorder := edBorderInactive
	if m.mode == modeEditing {
		edBorder = edBorderActive
	}
	editorBox := edBorder.
		Width(innerW).
		Height(edH - 2).
		Render(editorInner)

	var statusLine string
	switch {
	case m.running:
		statusLine = edStatusRun.Render(" ⟳  Running...")
	case m.err != nil:
		msg := m.err.Error()
		if len(msg) > w-6 {
			msg = msg[:w-6] + "…"
		}
		statusLine = edStatusErr.Render(" ✗  " + msg)
	case m.result != nil:
		statusLine = edStatusOk.Render(fmt.Sprintf(" ✓  %d rows  (%dms)",
			len(m.result.Rows), m.elapsed.Milliseconds()))
	default:
		statusLine = edHintStyle.Render(" ─  no results yet")
	}

	resultsH := h - edH - 1 - 2
	if resultsH < 3 {
		resultsH = 3
	}

	resBorder := edBorderInactive
	if m.mode == modeResults {
		resBorder = edBorderActive
	}

	resLabelSty := edLabelInactiveStyle
	if m.mode == modeResults {
		resLabelSty = edLabelStyle
	}
	resTitle := resLabelSty.Render(" Results ")

	resultsInner := resTitle + "\n" + m.renderResults(innerW-2, resultsH-2)

	resultsBox := resBorder.
		Width(innerW).
		Height(resultsH).
		Render(resultsInner)

	return editorBox + "\n" + statusLine + "\n" + resultsBox
}

func (m EditorModel) renderResults(w, h int) string {
	if m.err != nil {
		return edStatusErr.Render(" " + m.err.Error())
	}
	if m.result == nil {
		return edHintStyle.Render(" Run a query with Ctrl+E")
	}
	if len(m.result.Rows) == 0 {
		return edHintStyle.Render(" Query returned 0 rows")
	}

	visibleRows := h - 2
	if visibleRows < 1 {
		visibleRows = 1
	}

	var b strings.Builder

	var headerCells, separators []string
	for i, col := range m.result.Columns {
		cw := m.colWidths[i]
		headerCells = append(headerCells, padRight(col, cw))
		separators = append(separators, strings.Repeat("─", cw))
	}

	headerLine := "│ " + strings.Join(headerCells, " │ ") + " │"
	sepLine := "├─" + strings.Join(separators, "─┼─") + "─┤"

	b.WriteString(tblHeaderStyle.Render(clipLine(truncateLine(headerLine, m.scrollX, w), w)) + "\n")
	b.WriteString(tblBorderStyle.Render(clipLine(truncateLine(sepLine, m.scrollX, w), w)) + "\n")

	viewStart := 0
	if m.cursor >= visibleRows {
		viewStart = m.cursor - visibleRows + 1
	}
	endRow := viewStart + visibleRows
	if endRow > len(m.result.Rows) {
		endRow = len(m.result.Rows)
	}

	for i := viewStart; i < endRow; i++ {
		row := m.result.Rows[i]
		var cells []string
		for j, val := range row {
			cells = append(cells, padRight(val, m.colWidths[j]))
		}
		line := "│ " + strings.Join(cells, " │ ") + " │"
		line = clipLine(truncateLine(line, m.scrollX, w), w)
		if i == m.cursor {
			b.WriteString(tblSelStyle.Render(line) + "\n")
		} else {
			b.WriteString(tblRowStyle.Render(line) + "\n")
		}
	}

	return b.String()
}

func (m EditorModel) View() string {
	return m.ViewPanel(m.width, m.height)
}
