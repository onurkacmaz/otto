package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"otto/db"
)

type appState int

const (
	stateConnect appState = iota
	stateMain
)

type App struct {
	state   appState
	connect ConnectModel
	main    MainModel
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
		a.main = NewMainModel(msg.DB, msg.Cfg, a.width, a.height)
		a.state = stateMain
		return a, a.main.Init()
	case GoBackToConnectMsg:
		a.connect = NewConnectModel()
		a.connect.width = a.width
		a.connect.height = a.height
		a.state = stateConnect
		return a, a.connect.Init()
	}

	switch a.state {
	case stateConnect:
		updated, cmd := a.connect.Update(msg)
		a.connect = updated.(ConnectModel)
		return a, cmd
	case stateMain:
		updated, cmd := a.main.Update(msg)
		a.main = updated.(MainModel)
		return a, cmd
	}

	return a, nil
}

func (a App) View() string {
	switch a.state {
	case stateConnect:
		return a.connect.View()
	case stateMain:
		return a.main.View()
	}
	return ""
}
