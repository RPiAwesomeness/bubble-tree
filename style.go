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

	fgColor := lightDark(BLACK, WHITE)
	if highlightColor == nil {
		highlightColor = PURPLE
		fgColor = BLACK // Always black on purple for contrast
	}

	// Contrast should be handled better, but this is good enough for now
	return Styles{
		Shapes:     lipgloss.NewStyle().Margin(0).Foreground(highlightColor),
		Selected:   lipgloss.NewStyle().Margin(0).Background(highlightColor).Foreground(fgColor),
		Unselected: lipgloss.NewStyle().Margin(0).Foreground(lightDark(BLACK, WHITE)),
		Help:       lipgloss.NewStyle().Margin(0).Foreground(lightDark(BLACK, WHITE)),
	}
}
