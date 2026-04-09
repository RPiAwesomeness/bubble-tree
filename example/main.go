package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/term"
	tree "github.com/rpiawesomeness/bubble-tree"
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
	w, h, err := term.GetSize(os.Stdout.Fd())
	if err != nil {
		w = WIDTH
		h = HEIGHT
	}

	top, right, bottom, left := styleDoc.GetPadding()
	w = w - left - right
	h = h - top - bottom - 1

	nodes := []tree.Node{
		{
			Value: "history | grep docker",
			Desc:  "Search through the command history for any references to \"docker\"",
			Children: []tree.Node{
				{
					Value:    "history",
					Desc:     "Shows history of all commands in the terminal",
					Children: nil, // default value is nil, this is just for demonstration
				},
				{
					Value: "|",
					Desc:  "Used to combine two or more commands",
				},
				{
					Value: "grep",
					Desc:  "Short for 'global regular expression print' - used in searching & matching text files",
					Children: []tree.Node{
						{Value: "g", Desc: "7th letter of the English alphabet"},
						{Value: "r", Desc: "18th letter of the English alphabet"},
						{Value: "e", Desc: "5th letter of the English alphabet"},
						{Value: "p", Desc: "16th letter of the English alphabet"},
					},
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
			Tags:  []string{"command"},
			Children: []tree.Node{
				{
					Value: "echo",
					Desc:  "display a line of text",
				},
				{
					Value: "\"Success\"",
					Tags:  []string{"child", "nested"},
					Children: []tree.Node{
						{Value: "\"", Desc: "Begin quote"},
						{Value: "Success"},
						{Value: "\"", Desc: "End quote"},
					},
				},
			},
		},
	}

	return model{
		tree: tree.New(
			nodes,
			w, h,
			&tree.TreeOptions{
				ChildPrefix:       "==>",         // tree.Sharp and tree.Smooth are options provided, but this can be any string
				HighlightFullLine: true,          // Set to false to only highlight characters in .Value/.Desc
				HighlightColor:    lipgloss.Cyan, // Any color.Color value, defaults to a nice purple
				HelpKey:           "f1",          // Change keybind for showing help, default is "?". Requires ShowHelp to be set to true.
				ShowHelp:          true,          // Set to true to show help text, default is false
			}),
	}
}

type model struct {
	tree tree.Model
	msg  string
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "/":
			// A bit contrived, since we know the default state, but it serves the demo
			activeNodeVal := m.tree.ActiveNode().Value
			newChildren := []tree.Node{
				{Value: "SUCCESSfully found", Desc: "Active node value: " + activeNodeVal},
			}
			cmp := func(nodes *[]tree.Node) (node *tree.Node, continueSearch bool) {
				for i := range *nodes {
					nodeVal := (*nodes)[i].Value

					if !strings.Contains(nodeVal, "Success") {
						continue
					}

					isLeaf := nodeVal == "Success"
					return &(*nodes)[i], !isLeaf
				}

				return nil, false
			}
			if err := m.tree.SetChildrenFunc(newChildren, cmp); err != nil {
				m.msg = "ERROR: " + err.Error()
			}

		case "+":
			newChildren := []tree.Node{
				{Value: "Test0", Desc: "First new value"},
				{Value: "Test1", Desc: "Second new value"},
				{Value: "Test2", Desc: "Third new value"},
			}

			// Add child nodes at current depth
			path := m.tree.ActivePath()
			if len(path) == 1 {
				// Must always pass the root value, but can pass just that, path is variadic and optional
				newChildren[0].Desc = "Set via SetChildren without path"
				m.tree.SetChildren(newChildren, path[0].Value)
				break
			}

			pathValues := make([]string, len(path))
			for i, node := range path {
				pathValues[i] = node.Value
			}

			newChildren[0].Desc = "Set via SetChildren"
			if err := m.tree.SetChildren(newChildren, pathValues[0], pathValues[1:]...); err != nil {
				m.msg = "ERROR: " + err.Error()
			}

		case "enter":
			if m.msg == "" {
				// On the first time, update height to account for new string below
				m.tree.SetHeight(m.tree.Height() - 2)
			}

			var details strings.Builder
			details.WriteString("\nActive node selected: ")
			for i, node := range m.tree.ActivePath() {
				if i > 0 {
					details.WriteString(" > ")
				}
				details.WriteString(node.Value)
			}

			if active := m.tree.ActiveNode(); active != nil {
				fmt.Fprintf(&details, "\n(desc: %q, tags: %q)", active.Desc, strings.Join(active.Tags, ","))
			}

			m.msg = details.String()
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.tree, cmd = m.tree.Update(msg)
	return m, cmd
}

func (m model) View() tea.View {
	return tea.NewView(
		styleDoc.Render(
			lipgloss.JoinVertical(lipgloss.Left,
				m.tree.View(),
				lipgloss.NewStyle().Faint(true).Foreground(lipgloss.White).Render("+ replace children"),
				m.msg,
			),
		),
	)
}
