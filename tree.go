package tree

import (
	"fmt"
	"image/color"
	"math"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Node struct {
	Value    string
	Desc     string
	Tags     []string
	Children []Node
}

type Model struct {
	KeyMap KeyMap
	Styles Styles

	nodes  []Node
	cursor uint

	width  uint
	height uint

	// Prefix string for child nodes, defaults to Smooth TreeVariant
	childPrefix string
	// Highlight the full line from end-to-end, or just the contents
	highlightFullLine bool

	Help     help.Model
	showHelp bool
}

type TreeVariant string

const (
	Sharp  TreeVariant = " └──"
	Smooth TreeVariant = " ╰──"
)

type TreeOptions struct {
	HelpKey           string
	ChildPrefix       TreeVariant
	ShowHelp          bool
	HighlightColor    color.Color
	HighlightFullLine bool
}

const (
	VALUE_MAX_WIDTH uint = 10
	DESC_MAX_WIDTH  uint = 20
	DEFAULT_WIDTH   uint = 80
	DEFAULT_HEIGHT  uint = 24
)

func New(nodes []Node, width, height int, options *TreeOptions) Model {
	var w, h uint
	if width < 0 {
		w = 0
	} else {
		w = uint(width)
	}
	if height < 0 {
		h = 0
	} else {
		h = uint(height)
	}

	showHelp := false
	keyMap := DefaultKeyMap()
	childPrefix := string(Smooth)
	fullHighlight := false
	var highlightColor color.Color = nil

	if options != nil {
		showHelp = options.ShowHelp

		if options.HelpKey != "" {
			keyMap.ShowFullHelp.SetKeys(options.HelpKey)
			keyMap.ShowFullHelp.SetHelp(options.HelpKey, "help")
			keyMap.CloseFullHelp.SetKeys(options.HelpKey)
			keyMap.CloseFullHelp.SetHelp(options.HelpKey, "close help")
		}

		if options.ChildPrefix != "" {
			childPrefix = string(options.ChildPrefix)
		}

		fullHighlight = options.HighlightFullLine

		highlightColor = options.HighlightColor
	}

	return Model{
		KeyMap:            keyMap,
		Styles:            defaultStyles(highlightColor),
		childPrefix:       childPrefix,
		highlightFullLine: fullHighlight,

		nodes: nodes,

		width:  w,
		height: h,
		cursor: 0,

		showHelp: showHelp,
		Help:     help.New(),
	}
}

func (m Model) Nodes() []Node {
	return m.nodes
}

func (m *Model) SetNodes(nodes []Node) {
	m.nodes = nodes
	m.cursor = 0
}

func (m *Model) ActiveNode() *Node {
	index := 0

	var traverse func(*Node) *Node
	traverse = func(node *Node) *Node {
		// Check if current position matches cursor
		if index == int(m.cursor) {
			return node
		}
		index++

		// Search in children
		for i := range node.Children {
			if found := traverse(&node.Children[i]); found != nil {
				return found
			}
		}
		return nil
	}

	// Search through root nodes
	for i := range m.nodes {
		if found := traverse(&m.nodes[i]); found != nil {
			return found
		}
	}
	return nil
}

func (m Model) NumberOfNodes() uint {
	var countNodes func([]Node) int
	countNodes = func(nodes []Node) int {
		count := len(nodes)
		for _, node := range nodes {
			if node.Children != nil {
				count += countNodes(node.Children)
			}
		}
		return count
	}

	return uint(countNodes(m.nodes))
}

func (m Model) Width() uint {
	return m.width
}

func (m Model) Height() uint {
	return m.height
}

func (m *Model) SetSize(width, height uint) {
	m.width = width
	m.height = height
}

func (m *Model) SetWidth(newWidth uint) {
	m.width = newWidth
}

func (m *Model) SetHeight(newHeight uint) {
	m.height = newHeight
}

func (m Model) Cursor() uint {
	return m.cursor
}

func (m *Model) SetCursor(cursor uint) {
	m.cursor = cursor
}

func (m *Model) ResetCursor() {
	m.cursor = 0
}

func (m *Model) ShowHelp() bool {
	return m.showHelp
}

func (m *Model) NavUp() {
	if m.cursor <= 0 {
		m.ResetCursor()
		return
	}

	m.cursor--
}

func (m *Model) NavDown() {
	numNodes := m.NumberOfNodes()
	if m.cursor >= numNodes-1 {
		m.cursor = numNodes - 1
		return
	}

	m.cursor++
}

func (m *Model) NavParent() {
	if m.cursor == 0 {
		return
	}

	// Create a mapping of each node index to its parent index
	nodeToParent := make(map[uint]uint)

	var traverse func(nodes []Node, startIndex, parentIndex uint) uint
	traverse = func(nodes []Node, startIndex, parentIndex uint) uint {
		count := startIndex

		for _, node := range nodes {
			currentIndex := count
			count++ // Move past current node

			// Record parent of current node (except root nodes which have no parent)
			if parentIndex != ^uint(0) { // ^uint(0) is max uint value, used as "no parent"
				nodeToParent[currentIndex] = parentIndex
			}

			// Traverse children
			if len(node.Children) > 0 {
				count = traverse(node.Children, count, currentIndex)
			}
		}
		return count
	}

	// Start traversal with no parent (using max uint as sentinel)
	traverse(m.nodes, 0, ^uint(0))

	// Move cursor to parent if exists
	if parent, exists := nodeToParent[m.cursor]; exists {
		m.cursor = parent
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Up):
			m.NavUp()
		case key.Matches(msg, m.KeyMap.Down):
			m.NavDown()
		case key.Matches(msg, m.KeyMap.Parent):
			m.NavParent()
		case key.Matches(msg, m.KeyMap.ShowFullHelp):
			fallthrough
		case key.Matches(msg, m.KeyMap.CloseFullHelp):
			m.Help.ShowAll = !m.Help.ShowAll
		}
	}

	return m, nil
}

func (m Model) View() string {
	help := ""
	availableHeight := int(m.height)
	if m.showHelp {
		help = m.helpView()
		availableHeight -= lipgloss.Height(help)
	}

	renderedTree, _ := m.renderTree(m.nodes, 0, 0)
	sections := []string{
		lipgloss.NewStyle().Height(availableHeight).Render(renderedTree),
		help,
	}

	if len(m.nodes) == 0 {
		return "No data"
	}
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *Model) renderTree(nodes []Node, indent uint, count uint) (string, uint) {
	var b strings.Builder
	finalCount := count

	for _, node := range nodes {
		str := ""

		// If we aren't at the root, we add the arrow shape to the string
		if indent > 0 {
			indentSpaces := int((indent - 1) * 2)
			shape := fmt.Sprintf("%s%s ",
				strings.Repeat(" ", indentSpaces),
				m.Styles.Shapes.Render(m.childPrefix))
			str += shape
		}

		// Generate the correct index for the node
		idx := finalCount
		finalCount++

		// If we are at the cursor, add the selected style to the string
		style := m.Styles.Unselected
		if m.cursor == idx {
			style = m.Styles.Selected
		}

		// Format the string with fixed width for the value and description fields
		valLen := len(node.Value)

		if m.highlightFullLine {
			str += style.Render(fmt.Sprintf("%-*s\t\t%-*s", VALUE_MAX_WIDTH, node.Value, DESC_MAX_WIDTH, node.Desc))
		} else {
			valHighlightLen := int(math.Min(
				float64(valLen),
				math.Max(float64(VALUE_MAX_WIDTH), float64(valLen)),
			))
			fillerLen := int(VALUE_MAX_WIDTH) - valHighlightLen

			valueStr := fmt.Sprintf("%-*s", valHighlightLen, node.Value)
			str += fmt.Sprintf("%s%*s", style.Render(valueStr), fillerLen, "")

			if node.Desc != "" {
				descLen := len(node.Value)
				descHighlightLen := int(math.Min(
					float64(descLen),
					math.Max(float64(DESC_MAX_WIDTH), float64(descLen)),
				))
				fillerLen := int(DESC_MAX_WIDTH) - descHighlightLen
				descStr := fmt.Sprintf("%-*s", descHighlightLen, node.Desc)
				str += fmt.Sprintf("\t\t%s%*s", style.Render(descStr), fillerLen, "")
			}
		}

		str += "\n"
		b.WriteString(str)

		if node.Children != nil {
			childStr, childCount := m.renderTree(node.Children, indent+1, finalCount)
			finalCount += childCount - 1
			b.WriteString(childStr)
		}
	}

	return b.String(), finalCount
}
