package ui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"db-console/db"
)

const pageSize = 50

type dataLoadedMsg struct {
	result *db.QueryResult
}

type dataErrMsg struct {
	err error
}

type TableModel struct {
	db        db.DB
	schema    string
	tableName string
	result    *db.QueryResult
	err       error
	cursor    int
	offset    int
	scrollX   int
	width     int
	height    int
	colWidths []int
}

func NewTableModel(d db.DB, schema, name string, width, height int) TableModel {
	return TableModel{
		db:        d,
		schema:    schema,
		tableName: name,
		width:     width,
		height:    height,
	}
}

func (m TableModel) loadData() tea.Msg {
	result, err := m.db.FetchTableData(context.Background(), m.schema, m.tableName, pageSize, m.offset)
	if err != nil {
		return dataErrMsg{err: err}
	}
	return dataLoadedMsg{result: result}
}

func (m TableModel) Init() tea.Cmd {
	return m.loadData
}

func (m *TableModel) calcColWidths() {
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

type GoBackMsg struct{}

func (m TableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case dataLoadedMsg:
		if len(msg.result.Rows) == 0 && m.offset > 0 {
			m.offset -= pageSize
			return m, m.loadData
		}
		m.result = msg.result
		m.cursor = 0
		m.calcColWidths()
	case dataErrMsg:
		m.err = msg.err
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, func() tea.Msg { return GoBackMsg{} }
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
		case "n":
			m.offset += pageSize
			m.cursor = 0
			return m, m.loadData
		case "p":
			if m.offset >= pageSize {
				m.offset -= pageSize
				m.cursor = 0
				return m, m.loadData
			}
		case "r":
			return m, m.loadData
		}
	}
	return m, nil
}

func padRight(s string, w int) string {
	if len(s) > w {
		return s[:w]
	}
	return s + strings.Repeat(" ", w-len(s))
}

var (
	headerStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF6F61"))
	rowStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA"))
	selectedStyle  = lipgloss.NewStyle().Background(lipgloss.Color("#FF6F61")).Foreground(lipgloss.Color("#000"))
	borderStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	tableHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).MarginTop(1)
)

func truncateLine(s string, scrollX, maxWidth int) string {
	runes := []rune(s)
	if scrollX >= len(runes) {
		return ""
	}
	runes = runes[scrollX:]
	if len(runes) > maxWidth {
		runes = runes[:maxWidth]
	}
	return string(runes)
}

func (m TableModel) View() string {
	if m.err != nil {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
		return errStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	if m.result == nil {
		return "Loading..."
	}

	w := m.width
	if w == 0 {
		w = 80
	}
	visibleRows := m.height - 6
	if visibleRows < 1 {
		visibleRows = 10
	}

	var b strings.Builder

	rowCount := len(m.result.Rows)
	startRow := m.offset + 1
	if rowCount == 0 {
		startRow = 0
	}
	title := fmt.Sprintf(" %s.%s (%d-%d)", m.schema, m.tableName, startRow, m.offset+rowCount)
	b.WriteString(headerStyle.Render(title) + "\n\n")

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
	if endRow > len(m.result.Rows) {
		endRow = len(m.result.Rows)
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
	b.WriteString("\n" + tableHelpStyle.Render("q: back • j/k: navigate • h/l: scroll left/right • n/p: page • r: refresh"+scrollInfo))

	return b.String()
}
