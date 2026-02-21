package ui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"otto/db"
)

const pageSize = 50

type dataLoadedMsg struct {
	result *db.QueryResult
}

type dataErrMsg struct {
	err error
}

type GoBackMsg struct{}

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
			if idx := strings.IndexAny(val, "\n\r"); idx >= 0 {
				val = val[:idx]
			}
			if len([]rune(val)) > m.colWidths[i] {
				m.colWidths[i] = len([]rune(val))
			}
			if m.colWidths[i] > 30 {
				m.colWidths[i] = 30
			}
		}
	}
}

func (m TableModel) Update(msg tea.Msg) (TableModel, tea.Cmd) {
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
		case "esc", "q":
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
	var buf strings.Builder
	buf.Grow(len(s))
	for _, r := range s {
		if r < 0x20 || r == 0x7F {
			buf.WriteByte(' ')
		} else {
			buf.WriteRune(r)
		}
	}
	s = buf.String()
	runes := []rune(s)
	if len(runes) > w {
		return string(runes[:w])
	}
	return s + strings.Repeat(" ", w-len(runes))
}

func clipLine(s string, maxW int) string {
	runes := []rune(s)
	if len(runes) > maxW {
		return string(runes[:maxW])
	}
	return s
}

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

var (
	tblHeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF6F61"))
	tblRowStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#E6EDF3"))
	tblSelStyle    = lipgloss.NewStyle().Background(lipgloss.Color("#FF6F61")).Foreground(lipgloss.Color("#000000"))
	tblBorderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#30363D"))
)

func (m TableModel) ViewPanel(w, h int) string {
	if m.err != nil {
		errSty := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
		return errSty.Render(fmt.Sprintf(" Error: %v", m.err))
	}
	if m.result == nil {
		return " Loading..."
	}

	visibleRows := h - 4
	if visibleRows < 1 {
		visibleRows = 1
	}

	var b strings.Builder

	rowCount := len(m.result.Rows)
	startRow := m.offset + 1
	if rowCount == 0 {
		startRow = 0
	}
	title := fmt.Sprintf(" %s.%s  (%d – %d)", m.schema, m.tableName, startRow, m.offset+rowCount)
	b.WriteString(tblHeaderStyle.Render(title) + "\n\n")

	var headerCells, separators []string
	for i, col := range m.result.Columns {
		cw := m.colWidths[i]
		headerCells = append(headerCells, padRight(col, cw))
		separators = append(separators, strings.Repeat("─", cw))
	}

	headerLine := " │ " + strings.Join(headerCells, " │ ") + " │"
	sepLine := " ├─" + strings.Join(separators, "─┼─") + "─┤"

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
		line := " │ " + strings.Join(cells, " │ ") + " │"
		line = clipLine(truncateLine(line, m.scrollX, w), w)
		if i == m.cursor {
			b.WriteString(tblSelStyle.Render(line) + "\n")
		} else {
			b.WriteString(tblRowStyle.Render(line) + "\n")
		}
	}

	return b.String()
}

func (m TableModel) View() string {
	return m.ViewPanel(m.width, m.height)
}
