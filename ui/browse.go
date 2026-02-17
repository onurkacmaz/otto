package ui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"db-console/db"
)

type tablesLoadedMsg struct {
	tables []db.Table
}

type tablesErrMsg struct {
	err error
}

type GoBackToConnectMsg struct{}

type TableSelectedMsg struct {
	DB     db.DB
	Schema string
	Name   string
}

type OpenEditorMsg struct {
	DB db.DB
}

type tableItem struct {
	schema string
	name   string
}

func (t tableItem) Title() string       { return t.schema + "." + t.name }
func (t tableItem) Description() string { return t.schema }
func (t tableItem) FilterValue() string { return t.name }

type BrowseModel struct {
	db     db.DB
	list   list.Model
	err    error
	width  int
	height int
}

func NewBrowseModel(d db.DB, width, height int) BrowseModel {
	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 24
	}

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.SetHeight(1)

	l := list.New([]list.Item{}, delegate, width, height-4)
	l.Title = "Tables"
	l.SetShowHelp(true)
	l.DisableQuitKeybindings()

	return BrowseModel{
		db:     d,
		list:   l,
		width:  width,
		height: height,
	}
}

func (m BrowseModel) loadTables() tea.Msg {
	tables, err := m.db.ListTables(context.Background())
	if err != nil {
		return tablesErrMsg{err: err}
	}
	return tablesLoadedMsg{tables: tables}
}

func (m BrowseModel) Init() tea.Cmd {
	return m.loadTables
}

func (m BrowseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height-4)
	case tablesLoadedMsg:
		items := make([]list.Item, len(msg.tables))
		for i, t := range msg.tables {
			items[i] = tableItem{schema: t.Schema, name: t.Name}
		}
		m.list.SetItems(items)
	case tablesErrMsg:
		m.err = msg.err
	case tea.KeyMsg:
		if msg.Type == tea.KeyEsc && m.list.FilterState() != list.Filtering {
			if m.db != nil {
				m.db.Close(context.Background())
			}
			return m, func() tea.Msg { return GoBackToConnectMsg{} }
		}
if msg.String() == "s" && m.list.FilterState() != list.Filtering {
			return m, func() tea.Msg { return OpenEditorMsg{DB: m.db} }
		}
		if msg.Type == tea.KeyEnter {
			if item, ok := m.list.SelectedItem().(tableItem); ok {
				return m, func() tea.Msg {
					return TableSelectedMsg{DB: m.db, Schema: item.schema, Name: item.name}
				}
			}
		}
		m.list, cmd = m.list.Update(msg)
	default:
		m.list, cmd = m.list.Update(msg)
	}
	return m, cmd
}

func (m BrowseModel) View() string {
	if m.err != nil {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
		return errStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	return m.list.View()
}
