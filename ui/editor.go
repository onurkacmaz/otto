package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"db-console/db"
)

type editorMode int

const (
	modeEditing editorMode = iota
	modeResults
)

type queryResultMsg struct {
	result *db.QueryResult
}

type queryErrMsg struct {
	err error
}

type EditorModel struct {
	db        db.DB
	textarea  textarea.Model
	mode      editorMode
	result    *db.QueryResult
	err       error
	cursor    int
	scrollX   int
	width     int
	height    int
	colWidths []int
}

func NewEditorModel(d db.DB, width, height int) EditorModel {
	ta := textarea.New()
	ta.Placeholder = "SELECT * FROM ..."
	ta.SetWidth(width - 4)
	ta.SetHeight(8)
	ta.Focus()

	return EditorModel{
		db:       d,
		textarea: ta,
		mode:     modeEditing,
		width:    width,
		height:   height,
	}
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
		m.colWidths[i] = len(col)
	}
	for _, row := range m.result.Rows {
		for i, val := range row {
			if len(val) > m.colWidths[i] {
				m.colWidths[i] = len(val)
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
	result, err := m.db.ExecQuery(context.Background(), query)
	if err != nil {
		return queryErrMsg{err: err}
	}
	return queryResultMsg{result: result}
}

func (m EditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textarea.SetWidth(msg.Width - 4)
	case queryResultMsg:
		m.result = msg.result
		m.err = nil
		m.cursor = 0
		m.scrollX = 0
		m.mode = modeResults
		m.calcColWidths()
	case queryErrMsg:
		m.err = msg.err
		m.result = nil
		m.mode = modeResults
	case tea.KeyMsg:
		switch m.mode {
		case modeEditing:
			switch msg.String() {
			case "esc":
				return m, func() tea.Msg { return GoBackMsg{} }
			case "ctrl+e":
				return m, m.execQuery
			default:
				var cmd tea.Cmd
				m.textarea, cmd = m.textarea.Update(msg)
				return m, cmd
			}
		case modeResults:
			switch msg.String() {
			case "esc":
				return m, func() tea.Msg { return GoBackMsg{} }
			case "e":
				m.mode = modeEditing
				m.textarea.Focus()
				return m, textarea.Blink
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
	editorTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF6F61"))
	editorHelpStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).MarginTop(1)
	editorErrStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
)

func (m EditorModel) View() string {
	var b strings.Builder

	b.WriteString(editorTitleStyle.Render(" SQL Editor") + "\n\n")

	if m.mode == modeEditing {
		b.WriteString(m.textarea.View() + "\n")
		b.WriteString(editorHelpStyle.Render("ctrl+e: execute • esc: back to browse"))
		return b.String()
	}

	if m.err != nil {
		b.WriteString(editorErrStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n\n")
		b.WriteString(editorHelpStyle.Render("e: edit query • esc: back to browse"))
		return b.String()
	}

	if m.result == nil {
		b.WriteString("No results.\n")
		b.WriteString(editorHelpStyle.Render("e: edit query • esc: back to browse"))
		return b.String()
	}

	w := m.width
	if w == 0 {
		w = 80
	}
	visibleRows := m.height - 10
	if visibleRows < 1 {
		visibleRows = 10
	}

	rowCount := len(m.result.Rows)
	info := fmt.Sprintf(" Results: %d rows", rowCount)
	b.WriteString(headerStyle.Render(info) + "\n\n")

	var headerCells []string
	var separators []string
	for i, col := range m.result.Columns {
		cw := m.colWidths[i]
		headerCells = append(headerCells, padRight(col, cw))
		separators = append(separators, strings.Repeat("─", cw))
	}

	headerLine := " │ " + strings.Join(headerCells, " │ ") + " │"
	sepLine := " ├─" + strings.Join(separators, "─┼─") + "─┤"

	b.WriteString(headerStyle.Render(truncateLine(headerLine, m.scrollX, w)) + "\n")
	b.WriteString(borderStyle.Render(truncateLine(sepLine, m.scrollX, w)) + "\n")

	viewStart := 0
	if m.cursor >= visibleRows {
		viewStart = m.cursor - visibleRows + 1
	}
	endRow := viewStart + visibleRows
	if endRow > rowCount {
		endRow = rowCount
	}

	for i := viewStart; i < endRow; i++ {
		row := m.result.Rows[i]
		var cells []string
		for j, val := range row {
			cells = append(cells, padRight(val, m.colWidths[j]))
		}
		line := " │ " + strings.Join(cells, " │ ") + " │"
		line = truncateLine(line, m.scrollX, w)
		if i == m.cursor {
			b.WriteString(selectedStyle.Render(line) + "\n")
		} else {
			b.WriteString(rowStyle.Render(line) + "\n")
		}
	}

	scrollInfo := ""
	if m.scrollX > 0 {
		scrollInfo = fmt.Sprintf(" • scroll: %d", m.scrollX)
	}
	b.WriteString("\n" + editorHelpStyle.Render("e: edit query • j/k: navigate • h/l: scroll • esc: back to browse"+scrollInfo))

	return b.String()
}
