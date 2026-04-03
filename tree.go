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
	Children []Node
}

type Model struct {
	KeyMap KeyMap
	Styles Styles

	nodes  []Node
	cursor uint

	width  uint
	height uint

	childPrefix string
	// Highlight the full line from end-to-end, or just the contents
	highlightFullLine bool

	Help     help.Model
	showHelp bool

	AdditionalShortHelpKeys func() []key.Binding
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

	if options != nil {
		showHelp = options.ShowHelp

		if options.HelpKey != "" {
			keyMap.ShowFullHelp.SetKeys(options.HelpKey)
			keyMap.CloseFullHelp.SetKeys(options.HelpKey)
		} else {
			keyMap.ShowFullHelp.SetKeys("?")
			keyMap.CloseFullHelp.SetKeys("?")
		}

		if options.ChildPrefix != "" {
			childPrefix = string(options.ChildPrefix)
		}

		fullHighlight = options.HighlightFullLine
	}

	return Model{
		KeyMap:            keyMap,
		Styles:            defaultStyles(options.HighlightColor),
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
}

func (m Model) NumberOfNodes() uint {
	count := uint(0)

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
		m.cursor = 0
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

	if len(nodes) == 0 {
		return "No data"
	}
	return lipgloss.JoinVertical(lipgloss.Top, sections...)
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
		if m.highlightFullLine {
			str += style.Render(fmt.Sprintf("%-*s\t\t%-*s", VALUE_MAX_WIDTH, node.Value, DESC_MAX_WIDTH, node.Desc))
		} else {
			valLen := len(node.Value)
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
