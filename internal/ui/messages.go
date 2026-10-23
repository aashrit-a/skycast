// Package ui is the Bubble Tea application: state, layout and rendering.
package ui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/aashrit-a/skycast/internal/weather"
)

// refreshInterval is how often the weather is re-fetched automatically.
const refreshInterval = 15 * time.Minute

// tickMsg advances the animation by one frame.
type tickMsg time.Time

// weatherMsg carries a freshly loaded report.
type weatherMsg struct{ report *weather.Report }

// errMsg carries a failed load.
type errMsg struct{ err error }

// refreshMsg triggers the periodic re-fetch.
type refreshMsg struct{}

func tick(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func refreshTimer() tea.Cmd {
	return tea.Tick(refreshInterval, func(time.Time) tea.Msg { return refreshMsg{} })
}

func fetchCmd(ctx context.Context, load Loader, req LoadRequest) tea.Cmd {
	return func() tea.Msg {
		r, err := load(ctx, req)
		if err != nil {
			return errMsg{err: err}
		}
		return weatherMsg{report: r}
	}
}
