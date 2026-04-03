package main

import (
	"os"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	tree "github.com/rpiawesomeness/bubble-tree"
	"golang.org/x/term"
)

var (
	styleDoc = lipgloss.NewStyle().Padding(1)
)

const (
	WIDTH  = 80
	HEIGHT = 24
)

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		os.Exit(1)
	}
}

func initialModel() model {
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		w = WIDTH
		h = HEIGHT
	}

	top, right, bottom, left := styleDoc.GetPadding()
	w = w - left - right
	h = h - top - bottom

	nodes := []tree.Node{
		{
			Value: "history | grep docker",
			Desc: "Used in a Unix-like operating system to search through the " +
				"command history for any entries that contain the word 'docker.'",
			Children: []tree.Node{
				{
					Value:    "history",
					Desc:     "Shows the history of all commands in the terminal",
					Children: nil, // default value is nil, this is just for demonstration
				},
				{
					Value: "|",
					Desc:  "Used to combine two or more commands",
				},
				{
					Value: "grep",
					Desc:  "Short for 'global regular expression print'; A command used in searching and matching text files contained in the regular expressions.",
				},
				{
					Value: "docker",
					Desc:  "Used to interact with Docker",
				},
			},
		},
		{
			Value: "echo \"Success\"",
			Desc:  "A simple success string, printed to the terminal",
			Children: []tree.Node{
				{
					Value: "echo",
					Desc:  "display a line of text",
				},
				{Value: "echo"}, // No Desc field will not render the highligh to the right
			},
		},
	}

	return model{
		tree: tree.New(
			nodes,
			w, h,
			&tree.TreeOptions{
				ChildPrefix:       tree.Sharp,    // .Sharp and .Smooth are default values provided, but this can be any string
				HighlightFullLine: true,          // Set to false to only highlight characters in .Value/.Desc
				HighlightColor:    lipgloss.Cyan, // Any color.Color value, defaults to a nice purple
				HelpKey:           "f1",          // Change keybind for showing help, default is "?". Requires ShowHelp to be set to true.
				ShowHelp:          true,          // Set to true to show help text, default is false
			}),
	}
}

type model struct {
	tree tree.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.tree, cmd = m.tree.Update(msg)
	return m, cmd
}

func (m model) View() tea.View {
	return tea.NewView(styleDoc.Render(m.tree.View()))
}
