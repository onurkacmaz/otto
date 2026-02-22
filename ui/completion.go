package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
)

var sqlKeywords = []string{
	"SELECT", "FROM", "WHERE", "JOIN", "INNER", "LEFT", "RIGHT", "FULL",
	"OUTER", "CROSS", "ON", "GROUP", "ORDER", "BY", "HAVING", "LIMIT",
	"OFFSET", "INSERT", "INTO", "VALUES", "UPDATE", "SET", "DELETE",
	"CREATE", "TABLE", "DROP", "ALTER", "ADD", "COLUMN", "INDEX",
	"PRIMARY", "KEY", "FOREIGN", "REFERENCES", "UNIQUE", "NOT", "NULL",
	"AND", "OR", "IN", "IS", "LIKE", "ILIKE", "BETWEEN", "EXISTS", "AS",
	"DISTINCT", "COUNT", "SUM", "AVG", "MAX", "MIN",
	"CASE", "WHEN", "THEN", "ELSE", "END",
	"WITH", "UNION", "ALL", "INTERSECT", "EXCEPT",
	"RETURNING", "COALESCE", "NULLIF", "CAST", "OVER", "PARTITION",
	"WINDOW", "ROW_NUMBER", "RANK", "DENSE_RANK",
	"TRUE", "FALSE",
}

var sqlKeywordSet = func() map[string]bool {
	m := make(map[string]bool, len(sqlKeywords))
	for _, kw := range sqlKeywords {
		m[strings.ToLower(kw)] = true
	}
	return m
}()

type wordCtx struct {
	word  string
	table string
}

type suggKind int

const (
	kindKeyword suggKind = iota
	kindTable
	kindColumn
)

type sugg struct {
	text string
	kind suggKind
}

type completionModel struct {
	items       []sugg
	selected    int
	active      bool
	prefix      string
	lowercaseKw bool
}

func fuzzyMatchCI(pattern string, targets []string) []string {
	if len(targets) == 0 || pattern == "" {
		return nil
	}
	upper := make([]string, len(targets))
	for i, t := range targets {
		upper[i] = strings.ToUpper(t)
	}
	matches := fuzzy.Find(strings.ToUpper(pattern), upper)
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		out = append(out, targets[m.Index])
	}
	return out
}

func (c *completionModel) refresh(wctx wordCtx, tableNames []string, colsByTable map[string][]string, lowercaseKw bool) {
	c.lowercaseKw = lowercaseKw

	if wctx.table != "" {
		c.prefix = wctx.table + "." + wctx.word
		tableCols := colsByTable[strings.ToLower(wctx.table)]
		if len(tableCols) == 0 {
			c.active = false
			c.items = nil
			return
		}
		var items []sugg
		if wctx.word == "" {
			for _, col := range tableCols {
				items = append(items, sugg{text: wctx.table + "." + col, kind: kindColumn})
			}
		} else {
			for _, col := range fuzzyMatchCI(wctx.word, tableCols) {
				items = append(items, sugg{text: wctx.table + "." + col, kind: kindColumn})
			}
		}
		c.items = items
		c.active = len(items) > 0
		if c.selected >= len(items) {
			c.selected = 0
		}
		return
	}

	c.prefix = wctx.word
	if len([]rune(wctx.word)) < 1 {
		c.active = false
		c.items = nil
		return
	}

	kwCandidates := make([]string, len(sqlKeywords))
	for i, kw := range sqlKeywords {
		if lowercaseKw {
			kwCandidates[i] = strings.ToLower(kw)
		} else {
			kwCandidates[i] = kw
		}
	}

	colSeen := map[string]bool{}
	var allCols []string
	for _, cols := range colsByTable {
		for _, col := range cols {
			if !colSeen[col] {
				colSeen[col] = true
				allCols = append(allCols, col)
			}
		}
	}

	var items []sugg
	for _, m := range fuzzy.Find(strings.ToUpper(wctx.word), sqlKeywords) {
		text := m.Str
		if lowercaseKw {
			text = strings.ToLower(text)
		}
		items = append(items, sugg{text: text, kind: kindKeyword})
	}
	for _, t := range fuzzyMatchCI(wctx.word, tableNames) {
		items = append(items, sugg{text: t, kind: kindTable})
	}
	for _, col := range fuzzyMatchCI(wctx.word, allCols) {
		items = append(items, sugg{text: col, kind: kindColumn})
	}

	c.items = items
	c.active = len(items) > 0
	if c.selected >= len(items) {
		c.selected = 0
	}
}

func (c *completionModel) next() {
	if len(c.items) > 0 {
		c.selected = (c.selected + 1) % len(c.items)
	}
}

func (c *completionModel) prev() {
	if len(c.items) > 0 {
		c.selected = (c.selected - 1 + len(c.items)) % len(c.items)
	}
}

func (c *completionModel) current() string {
	if !c.active || len(c.items) == 0 {
		return ""
	}
	return c.items[c.selected].text
}

func (c *completionModel) dismiss() {
	c.active = false
	c.items = nil
	c.selected = 0
}

const (
	popupInnerW   = 28
	popupMaxItems = 6
)

var (
	popupBg = lipgloss.Color("#1C2128")

	popupBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#444C56")).
				Background(popupBg)

	popupSelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#316DCA")).
			Bold(true).
			PaddingLeft(1)

	popupKwStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6CB6FF")).
			Background(popupBg).
			PaddingLeft(1)

	popupTableStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#57AB5A")).
			Background(popupBg).
			PaddingLeft(1)

	popupColStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DCBDFB")).
			Background(popupBg).
			PaddingLeft(1)

	popupHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#444C56")).
			Background(popupBg).
			PaddingLeft(1)
)

func (c *completionModel) renderPopup() string {
	if !c.active || len(c.items) == 0 {
		return ""
	}
	var lines []string
	for i, item := range c.items {
		if i >= popupMaxItems {
			break
		}
		text := item.text
		if runes := []rune(text); len(runes) > popupInnerW-2 {
			text = string(runes[:popupInnerW-2])
		}
		var line string
		switch {
		case i == c.selected:
			line = popupSelStyle.Width(popupInnerW).Render(text)
		case item.kind == kindKeyword:
			line = popupKwStyle.Width(popupInnerW).Render(text)
		case item.kind == kindTable:
			line = popupTableStyle.Width(popupInnerW).Render(text)
		default:
			line = popupColStyle.Width(popupInnerW).Render(text)
		}
		lines = append(lines, line)
	}

	var hint string
	kwMode := "KW:UPPER"
	if c.lowercaseKw {
		kwMode = "KW:lower"
	}
	if len(c.items) > popupMaxItems {
		hint = fmt.Sprintf("+%d  ↑↓ Tab Esc  Ctrl+T %s", len(c.items)-popupMaxItems, kwMode)
	} else {
		hint = fmt.Sprintf("↑↓ Tab Esc  Ctrl+T %s", kwMode)
	}
	lines = append(lines, popupHintStyle.Width(popupInnerW).Render(hint))

	return popupBorderStyle.Render(strings.Join(lines, "\n"))
}

func overlayAtRow(base, overlay string, startRow int) string {
	baseLines := strings.Split(base, "\n")
	overlayLines := strings.Split(overlay, "\n")
	result := make([]string, len(baseLines))
	copy(result, baseLines)
	for i, ol := range overlayLines {
		r := startRow + i
		if r < 0 || r >= len(result) {
			continue
		}
		result[r] = ol
	}
	return strings.Join(result, "\n")
}
