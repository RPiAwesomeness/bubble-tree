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
	Tags     []string
	Children []Node
}

type Model struct {
	KeyMap KeyMap
	Styles Styles

	nodes    []Node
	numNodes uint
	cursor   uint

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
	HighlightFullLine bool
}

const (
	VALUE_MAX_WIDTH uint = 10
	DESC_MAX_WIDTH  uint = 20
	DEFAULT_WIDTH   uint = 80
	DEFAULT_HEIGHT  uint = 24
)

func New(nodes []Node, width, height int, options *TreeOptions) Model {
	w := DEFAULT_WIDTH
	if width > 0 {
		w = uint(width)
	}

	h := DEFAULT_HEIGHT
	if height > 0 {
		h = uint(height)
	}

	showHelp := false
	keyMap := DefaultKeyMap()
	childPrefix := string(Smooth)
	fullHighlight := false

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
	}

	return Model{
		KeyMap:            keyMap,
		Styles:            defaultStyles(),
		childPrefix:       childPrefix,
		highlightFullLine: fullHighlight,

		nodes:    nodes,
		numNodes: numberOfNodes(nodes),

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

// Used to reset the tree to a completely new state, resetting cursor to root
func (m *Model) SetNodes(nodes []Node) {
	m.nodes = nodes
	m.numNodes = numberOfNodes(nodes)
	m.cursor = 0
}

// Update nodes in the tree along with total count, setting cursor to value provided.
//
// NOTE: If the cursor value is greater than or equal to the number of nodes, the cursor will be
// set to the last valid index
func (m *Model) UpdateNodes(nodes []Node, cursor uint) {
	m.nodes = nodes
	m.numNodes = numberOfNodes(nodes)
	if cursor >= m.numNodes {
		m.cursor = m.numNodes - 1
	} else {
		m.cursor = cursor
	}
}

// Updates child nodes at path specified and numNodes to match new node count.
//
// Searches the tree for the values provided for path, setting the Children only
// if matching nodes are found for every path value.
func (m *Model) SetChildren(childNodes []Node, rootNodeValue string, path ...string) error {
	idx := -1
	for i, n := range m.nodes {
		if n.Value == rootNodeValue {
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Errorf("root node not found for value %q", rootNodeValue)
	}

	// Start at found root node
	n := &m.nodes[idx]

	for _, searchVal := range path {
		// Iterate over all nodes at the current depth
		for i, child := range n.Children {
			// Comparing value against the current path string
			if child.Value != searchVal {
				continue
			}

			// This node matches, point n at it and stop early
			n = &n.Children[i]
			break
		}

		if n == nil {
			return fmt.Errorf("branch with %q value not found", searchVal)
		}
	}

	if n == nil {
		return fmt.Errorf("no matching node to set children on")
	}

	n.Children = childNodes
	m.numNodes = numberOfNodes(m.nodes)

	return nil
}

// Wrapper around SetChildren, starting search from first node.
//
// Helper method, intended for usage with trees with a single root.
func (m *Model) SetChildrenFromRoot(childNodes []Node, path ...string) error {
	if len(m.nodes) == 0 {
		return fmt.Errorf("No nodes")
	}
	return m.SetChildren(childNodes, m.nodes[0].Value, path...)
}

// Wrapper around SetChildren, starting search from node specified by index.
func (m *Model) SetChildrenStartingAt(childNodes []Node, index uint, path ...string) error {
	numRootNodes := uint(len(m.nodes))
	if numRootNodes == 0 {
		return fmt.Errorf("No nodes")
	}
	if index >= numRootNodes {
		return fmt.Errorf("Index %d invalid for %d root nodes", index, numRootNodes)
	}
	return m.SetChildren(childNodes, m.nodes[index].Value, path...)
}

func (m *Model) SetChildrenFunc(childNodes []Node, cmp func(nodes *[]Node) (node *Node, continueSearch bool)) error {
	nodes := &m.nodes
	for {
		result, continueSearch := cmp(nodes)
		if result == nil {
			return fmt.Errorf("Unable to find matching value")
		}
		if continueSearch {
			nodes = &result.Children
			continue
		}

		// Found match & not continuing, set children
		result.Children = childNodes
		m.numNodes = numberOfNodes(m.nodes)
		break
	}

	return nil
}

func (m Model) ActiveNode() *Node {
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

func (m Model) ActivePath() []Node {
	index := 0

	var findPath func([]Node) []Node
	findPath = func(nodes []Node) []Node {
		for i, node := range nodes {
			// Check if this is the active node
			if index == int(m.cursor) {
				return []Node{nodes[i]}
			}
			index++

			// Check children if active node not found yet
			if len(node.Children) == 0 {
				continue
			}

			if path := findPath(node.Children); path != nil {
				// Prepend current node to the path found in children
				return append([]Node{nodes[i]}, path...)
			}
		}
		return nil
	}

	return findPath(m.nodes)
}

func (m Model) NumberOfNodes() int {
	return int(m.numNodes)
}

func (m Model) Width() int {
	return int(m.width)
}

func (m Model) Height() int {
	return int(m.height)
}

func (m *Model) SetSize(width, height int) {
	if width >= 0 {
		m.width = uint(width)
	}
	if height >= 0 {
		m.height = uint(width)
	}
}

func (m *Model) SetWidth(width int) {
	if width < 0 {
		return
	}
	m.width = uint(width)
}

func (m *Model) SetHeight(height int) {
	if height < 0 {
		return
	}
	m.height = uint(height)
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
	numNodes := m.numNodes
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
	if len(m.nodes) == 0 {
		return "No data"
	}

	renderedTree, _ := m.renderTree(m.nodes, 0, 0)

	if !m.showHelp {
		return renderedTree
	}

	help := m.helpView()
	availableHeight := int(m.height) - lipgloss.Height(help)
	return lipgloss.NewStyle().
		Background(m.Styles.Background).
		Render(
			lipgloss.JoinVertical(lipgloss.Left,
				lipgloss.NewStyle().
					Height(availableHeight).
					Background(m.Styles.Background).
					Render(renderedTree),
				help,
			),
		)
}

type row struct {
	indent string
	value  string
	desc   string
	idx    uint
}

func (m Model) renderTree(nodes []Node, indent uint, count uint) (string, uint) {
	var rows []row
	maxColWidth := 0

	// First pass: collect data and find the widest column 1 (visible width)
	var collect func([]Node, uint, uint) uint
	collect = func(nodes []Node, indent uint, count uint) uint {
		finalCount := count
		for _, node := range nodes {
			idx := finalCount
			finalCount++

			indentStr := ""
			if indent > 0 {
				indentStr = fmt.Sprintf("%s%s",
					strings.Repeat(" ", int((indent-1)*2)),
					m.Styles.Shapes.Render(m.childPrefix),
				)
			}

			// Use lipgloss.Width for proper Unicode/ANSI handling
			width := lipgloss.Width(indentStr) + lipgloss.Width(node.Value) + 1
			if width > maxColWidth {
				maxColWidth = width
			}

			rows = append(rows, row{indentStr, node.Value, node.Desc, idx})

			if len(node.Children) > 0 {
				finalCount = collect(node.Children, indent+1, finalCount)
			}
		}
		return finalCount
	}

	finalCount := collect(nodes, indent, count)

	// Second pass: render with manual padding (plain spaces between styled parts)
	var b strings.Builder
	for _, r := range rows {
		style := m.Styles.Unselected
		if m.cursor == r.idx {
			style = m.Styles.Selected
		}

		// Calculate how many spaces needed to align to maxColWidth + 1 (for gap)
		currentWidth := lipgloss.Width(r.indent) + lipgloss.Width(r.value)
		padding := strings.Repeat(" ", maxColWidth-currentWidth+2) // 2 include spaces

		if m.highlightFullLine {
			// Style everything after the tree indent
			content := r.value + padding + r.desc
			lineContent := fmt.Sprintf("%s%s", r.indent, style.Render(content))
			fmt.Fprintf(&b, "%s%s",
				lineContent,
				style.Render(strings.Repeat(" ", int(m.width)-lipgloss.Width(lineContent))),
			)
		} else {
			// Style value and desc separately, keep padding plain
			fmt.Fprintf(&b, "%s%s%s%s",
				r.indent,
				style.Render(r.value),
				padding, // Plain spaces - never highlighted
				style.Render(r.desc),
			)
		}
		fmt.Fprintln(&b)
	}

	return b.String(), finalCount
}

func numberOfNodes(nodes []Node) uint {
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

	return uint(countNodes(nodes))
}
