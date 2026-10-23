package ui

import (
	"context"
	"io"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/aashrit-a/skycast/internal/config"
	"github.com/aashrit-a/skycast/internal/scene"
	"github.com/aashrit-a/skycast/internal/weather"
)

// onceSteps is how far the scene is simulated before the single frame.
const onceSteps = 60

// RunOnce loads the weather (or the first demo scene), prints one frame
// with the panel beside it to out, and returns. Errors are returned for
// the caller to print.
func RunOnce(ctx context.Context, cfg config.Config, load Loader, palette scene.Palette, out io.Writer) error {
	now := time.Now()
	var r *weather.Report
	if cfg.Demo {
		u, _ := weather.ParseUnits(cfg.Units)
		r = demoReport(0, u, now)
	} else {
		u, ok := weather.ParseUnits(cfg.Units)
		var err error
		r, err = load(ctx, LoadRequest{Units: u, AutoUnits: !ok})
		if err != nil {
			return err
		}
	}
	rng := rand.New(rand.NewSource(now.UnixNano()))
	_, err := io.WriteString(out, RenderOnce(r, cfg, palette, rng, now)+"\n")
	return err
}

// RenderOnce renders a single frame for r. It is deterministic for a given
// rng and now, which the golden test relies on.
func RenderOnce(r *weather.Report, cfg config.Config, palette scene.Palette, rng *rand.Rand, now time.Time) string {
	m := Model{cfg: cfg, rng: rng, styles: newStyles(cfg.NoColor, cfg.ASCII)}
	s := scene.New(m.sceneParams(r), rng)
	s.Resize(cfg.Width, cfg.Height)
	for i := 0; i < onceSteps; i++ {
		s.Step()
	}
	c := scene.NewCanvas(cfg.Width, cfg.Height)
	s.Draw(c)
	panel := renderPanel(r, m.styles, panelWidth, max(cfg.Height, fullPanelHeight), now)
	out := lipgloss.JoinHorizontal(lipgloss.Center, c.Render(palette), strings.Repeat(" ", panelGap), panel)
	lines := strings.Split(out, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " ")
	}
	return strings.Join(lines, "\n")
}
