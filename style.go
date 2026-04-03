package tree

import (
	"os"

	"charm.land/lipgloss/v2"
)

const (
	BOTTOM_LEFT_CURVED   = " ╰──"
	BOTTOM_LEFT_STRAIGHT = " └──"

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

func defaultStyles() Styles {
	hasDarkBg := lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
	lightDark := lipgloss.LightDark(hasDarkBg)

	return Styles{
		Shapes:     lipgloss.NewStyle().Margin(0).Foreground(PURPLE),
		Selected:   lipgloss.NewStyle().Margin(0).Background(PURPLE),
		Unselected: lipgloss.NewStyle().Margin(0).Foreground(lightDark(BLACK, WHITE)),
		Help:       lipgloss.NewStyle().Margin(0).Foreground(lightDark(BLACK, WHITE)),
	}
}
