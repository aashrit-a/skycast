// Package ui is the Bubble Tea application: state, layout and rendering.
package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// tickMsg advances the animation by one frame.
type tickMsg time.Time

func tick(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg { return tickMsg(t) })
}
