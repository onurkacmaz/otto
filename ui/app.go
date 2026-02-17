package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"otto/db"
)

type appState int

const (
	stateConnect appState = iota
	stateBrowse
	stateTable
	stateEditor
)

type App struct {
	state   appState
	connect ConnectModel
	browse  BrowseModel
	table   TableModel
	editor  EditorModel
	width   int
	height  int
}

func NewApp() App {
	return App{
		state:   stateConnect,
		connect: NewConnectModel(),
	}
}

func (a App) Init() tea.Cmd {
	return a.connect.Init()
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return a, tea.Quit
		}
	case ConnectedMsg:
		db.SaveConnection(msg.Cfg)
		a.browse = NewBrowseModel(msg.DB, a.width, a.height)
		a.state = stateBrowse
		return a, a.browse.Init()
	case TableSelectedMsg:
		a.table = NewTableModel(msg.DB, msg.Schema, msg.Name, a.width, a.height)
		a.state = stateTable
		return a, a.table.Init()
	case OpenEditorMsg:
		a.editor = NewEditorModel(msg.DB, a.width, a.height)
		a.state = stateEditor
		return a, a.editor.Init()
	case GoBackMsg:
		a.state = stateBrowse
		return a, nil
	case GoBackToConnectMsg:
		a.connect = NewConnectModel()
		a.state = stateConnect
		return a, a.connect.Init()
	}

	switch a.state {
	case stateConnect:
		updated, cmd := a.connect.Update(msg)
		a.connect = updated.(ConnectModel)
		return a, cmd
	case stateBrowse:
		updated, cmd := a.browse.Update(msg)
		a.browse = updated.(BrowseModel)
		return a, cmd
	case stateTable:
		updated, cmd := a.table.Update(msg)
		a.table = updated.(TableModel)
		return a, cmd
	case stateEditor:
		updated, cmd := a.editor.Update(msg)
		a.editor = updated.(EditorModel)
		return a, cmd
	}

	return a, nil
}

func (a App) View() string {
	switch a.state {
	case stateConnect:
		return a.connect.View()
	case stateBrowse:
		return a.browse.View()
	case stateTable:
		return a.table.View()
	case stateEditor:
		return a.editor.View()
	}
	return ""
}
