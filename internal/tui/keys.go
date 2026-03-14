package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Quit       key.Binding
	Back       key.Binding
	Up         key.Binding
	Down       key.Binding
	Enter      key.Binding
	Tab        key.Binding
	Help       key.Binding
	Delete     key.Binding
	Trash      key.Binding
	Export     key.Binding
	Copy       key.Binding
	Select     key.Binding
	BulkDelete key.Binding
}

var keys = keyMap{
	Quit:       key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Back:       key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
	Up:         key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:       key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	Enter:      key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	Tab:        key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "switch pane")),
	Help:       key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Delete:     key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
	Trash:      key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "move to trash")),
	Export:     key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "export")),
	Copy:       key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "copy path")),
	Select:     key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "select")),
	BulkDelete: key.NewBinding(key.WithKeys("D"), key.WithHelp("D", "bulk delete")),
}
