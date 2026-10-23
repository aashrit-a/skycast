// Command skycast shows the current weather as an animated ASCII scene.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/aashrit-a/skycast/internal/config"
	"github.com/aashrit-a/skycast/internal/location"
	"github.com/aashrit-a/skycast/internal/scene"
	"github.com/aashrit-a/skycast/internal/ui"
	"github.com/aashrit-a/skycast/internal/weather"
)

// version is set at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	cfg, err := config.Parse(args, os.Getenv, stderr)
	if errors.Is(err, config.ErrHelp) {
		fmt.Fprint(stdout, config.Usage)
		return 0
	}
	var ue *config.UsageError
	if errors.As(err, &ue) {
		fmt.Fprintf(stderr, "skycast: %v\nRun 'skycast --help' for usage.\n", err)
		return 2
	}
	if err != nil {
		fmt.Fprintf(stderr, "skycast: %v\n", err)
		return 1
	}
	if cfg.Version {
		fmt.Fprintf(stdout, "skycast %s\n", version)
		return 0
	}
	if cfg.DebugFetch {
		return debugFetch(cfg, stdout, stderr)
	}
	loader := ui.NewLoader(cfg, "skycast/"+version)
	if cfg.Once {
		// Piped output gets no escape codes.
		if !isTerminal(os.Stdout) {
			cfg.NoColor = true
		}
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		if err := ui.RunOnce(ctx, cfg, loader, scene.NewPalette(cfg.NoColor), stdout); err != nil {
			fmt.Fprintf(stderr, "skycast: %v\n", err)
			return 1
		}
		return 0
	}

	m := ui.New(ui.Options{
		Config:  cfg,
		Palette: scene.NewPalette(cfg.NoColor),
		Loader:  loader,
	})
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(stderr, "skycast: %v\n", err)
		return 1
	}
	return 0
}

// debugFetch prints the parsed report as plain text (hidden --debug-fetch).
func debugFetch(cfg config.Config, stdout, stderr io.Writer) int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	ua := "skycast/" + version
	loc, err := location.NewResolver(ua).Resolve(ctx, cfg.City, cfg.Lat, cfg.Lon, cfg.HasCoords)
	if err != nil {
		fmt.Fprintf(stderr, "skycast: %v\n", err)
		return 1
	}
	units, ok := weather.ParseUnits(cfg.Units)
	if !ok {
		units = weather.DefaultUnits(loc.CountryCode)
	}
	client := weather.NewClient(ua, weather.NewCache())
	client.NoCacheRead = cfg.NoCache
	r, err := client.Fetch(ctx, loc.Lat, loc.Lon, units)
	if err != nil {
		fmt.Fprintf(stderr, "skycast: %v\n", err)
		return 1
	}
	r.Location = loc.Display()
	cond, in, label := weather.Classify(r.Current.Code, r.Current.IsDay)
	c := r.Current
	fmt.Fprintf(stdout, "location:  %s (%.4f, %.4f)\n", r.Location, loc.Lat, loc.Lon)
	fmt.Fprintf(stdout, "fetched:   %s (stale=%v)\n", r.FetchedAt.Format(time.RFC3339), r.Stale)
	fmt.Fprintf(stdout, "condition: %s / %s / %s (code %d, day=%v)\n", label, cond, in, c.Code, c.IsDay)
	fmt.Fprintf(stdout, "temp:      %.1f%s (feels %.1f)\n", c.Temp, weather.TempSuffix(units), c.FeelsLike)
	fmt.Fprintf(stdout, "humidity:  %d%%  precip: %.2f\n", c.Humidity, c.Precip)
	fmt.Fprintf(stdout, "wind:      %.1f %s %s (drift %.2f)\n", c.WindSpeed, weather.WindSuffix(units), weather.Compass(c.WindDirDeg), weather.WindDrift(c.WindSpeed, c.WindDirDeg, units))
	for _, d := range r.Daily {
		_, _, l := weather.Classify(d.Code, true)
		fmt.Fprintf(stdout, "  %s  %5.1f/%5.1f  %3d%%  rise %s set %s  %s\n", d.Date.Format("Mon 01-02"), d.High, d.Low, d.PrecipChance, d.Sunrise.Format("15:04"), d.Sunset.Format("15:04"), l)
	}
	return 0
}

// isTerminal reports whether f is a character device (a TTY), using only
// the standard library.
func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}
