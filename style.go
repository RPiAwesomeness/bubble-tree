package tree

import (
	"image/color"
	"os"

	"charm.land/lipgloss/v2"
)

const (
	WHITE = lipgloss.White
	BLACK = lipgloss.Black
)

var (
	PURPLE = lipgloss.Color("#bd93f9")
)

type Styles struct {
	Shapes     lipgloss.Style
	Selected   lipgloss.Style
	Unselected lipgloss.Style
	Help       lipgloss.Style
}

func defaultStyles(highlightColor color.Color) Styles {
	hasDarkBg := lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
	lightDark := lipgloss.LightDark(hasDarkBg)

	if highlightColor == nil {
		highlightColor = PURPLE
	}

	return Styles{
		Shapes:     lipgloss.NewStyle().Margin(0).Foreground(highlightColor),
		Selected:   lipgloss.NewStyle().Margin(0).Background(highlightColor).Foreground(BLACK),
		Unselected: lipgloss.NewStyle().Margin(0).Foreground(lightDark(BLACK, WHITE)),
		Help:       lipgloss.NewStyle().Margin(0).Foreground(lightDark(BLACK, WHITE)),
	}
}
