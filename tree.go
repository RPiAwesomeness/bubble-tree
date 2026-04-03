package tree

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Node struct {
	Value    string
	Desc     string
	Children []Node
}

type Model struct {
	KeyMap KeyMap
	Styles Styles

	width  int
	height int
	nodes  []Node
	cursor int

	Help     help.Model
	showHelp bool

	AdditionalShortHelpKeys func() []key.Binding
}

const (
	VALUE_WIDTH = 10
	DESC_WIDTH  = 20
)

func New(nodes []Node, width int, height int) Model {
	return Model{
		KeyMap: DefaultKeyMap(),
		Styles: defaultStyles(),

		width:  width,
		height: height,
		nodes:  nodes,

		showHelp: true,
		Help:     help.New(),
	}
}

func (m Model) Nodes() []Node {
	return m.nodes
}

func (m *Model) SetNodes(nodes []Node) {
	m.nodes = nodes
}

func (m *Model) NumberOfNodes() int {
	count := 0

	var countNodes func([]Node)
	countNodes = func(nodes []Node) {
		for _, node := range nodes {
			count++
			if node.Children != nil {
				countNodes(node.Children)
			}
		}
	}

	countNodes(m.nodes)

	return count
}

func (m Model) Width() int {
	return m.width
}

func (m Model) Height() int {
	return m.height
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *Model) SetWidth(newWidth int) {
	m.SetSize(newWidth, m.height)
}

func (m *Model) SetHeight(newHeight int) {
	m.SetSize(m.width, newHeight)
}

func (m Model) Cursor() int {
	return m.cursor
}

func (m *Model) SetCursor(cursor int) {
	m.cursor = cursor
}

func (m *Model) SetShowHelp() bool {
	return m.showHelp
}

func (m *Model) NavUp() {
	if m.cursor <= 0 {
		m.cursor = 0
		return
	}

	m.cursor--
}

func (m *Model) NavDown() {
	numNodes := m.NumberOfNodes()
	if m.cursor >= numNodes-1 {
		m.cursor = numNodes
		return
	}

	m.cursor++
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Up):
			m.NavUp()
		case key.Matches(msg, m.KeyMap.Down):
			m.NavDown()
		case key.Matches(msg, m.KeyMap.ShowFullHelp):
			fallthrough
		case key.Matches(msg, m.KeyMap.CloseFullHelp):
			m.Help.ShowAll = !m.Help.ShowAll
		}
	}

	return m, nil
}

func (m Model) View() string {
	nodes := m.Nodes()

	help := ""
	availableHeight := m.height
	if m.showHelp {
		help = m.helpView()
		availableHeight -= lipgloss.Height(help)
	}

	renderedTree, _ := m.renderTree(m.nodes, 0, 0)
	sections := []string{
		lipgloss.NewStyle().Height(availableHeight).Render(renderedTree),
		help,
	}

	if len(nodes) == 0 {
		return "No data"
	}
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *Model) renderTree(nodes []Node, indent int, count int) (string, int) {
	var b strings.Builder
	finalCount := count

	for _, node := range nodes {
		str := ""

		// If we aren't at the root, we add the arrow shape to the string
		if indent > 0 {
			shape := strings.Repeat(" ", (indent-1)*2) + m.Styles.Shapes.Render(BOTTOM_LEFT_CURVED) + " "
			str += shape
		}

		// Generate the correct index for the node
		idx := finalCount
		finalCount++

		// Format the string with fixed width for the value and description fields
		valueStr := fmt.Sprintf("%-*s", VALUE_WIDTH, node.Value)
		descStr := fmt.Sprintf("%-*s", DESC_WIDTH, node.Desc)

		// If we are at the cursor, we add the selected style to the string
		style := m.Styles.Unselected
		if m.cursor == idx {
			style = m.Styles.Selected
		}
		str += fmt.Sprintf("%s\t\t%s\n", style.Render(valueStr), style.Render(descStr))

		b.WriteString(str)

		if node.Children != nil {
			childStr, childCount := m.renderTree(node.Children, indent+1, finalCount)
			finalCount += childCount
			b.WriteString(childStr)
		}
	}

	return b.String(), finalCount
}
