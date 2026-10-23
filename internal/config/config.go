// Package config parses and validates skycast's command-line flags.
package config

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/pflag"
)

// Default values for flags.
const (
	DefaultFPS    = 12
	MinFPS        = 4
	MaxFPS        = 30
	DefaultWidth  = 40
	DefaultHeight = 12
)

// SceneNames lists the values accepted by --scene, in display order.
var SceneNames = []string{
	"clear", "night", "cloudy", "overcast", "fog", "drizzle",
	"rain", "freezing", "snow", "storm", "hail",
}

// IntensityNames lists the values accepted by --intensity.
var IntensityNames = []string{"light", "moderate", "heavy"}

// ErrHelp is returned by Parse when -h/--help was requested.
var ErrHelp = errors.New("help requested")

// UsageError is a validation error that should exit with status 2.
type UsageError struct{ Msg string }

func (e *UsageError) Error() string { return e.Msg }

// Config holds every option skycast can be started with.
type Config struct {
	City       string  // positional args joined with spaces
	Lat, Lon   float64 // valid only when HasCoords
	HasCoords  bool
	Units      string // "", "metric" or "imperial"; "" = based on country
	Once       bool
	Width      int
	Height     int
	Demo       bool
	Ambient    bool
	Scene      string // "" or one of SceneNames
	Intensity  string // "" or one of IntensityNames
	FPS        int
	NoColor    bool
	ASCII      bool
	NoCache    bool
	Version    bool
	DebugFetch bool // hidden: print the parsed report and exit
}

// Usage is the help text printed for -h/--help.
const Usage = `skycast - weather in your terminal, animated.

Usage:
  skycast [city] [flags]

Examples:
  skycast                       # auto-detect location
  skycast tokyo
  skycast "san jose, ca" -u imperial
  skycast --lat 48.85 --lon 2.35
  skycast --once
  skycast --demo
  skycast --scene snow          # force a scene, real weather in panel

Flags:
  -u, --units string      metric | imperial (default: based on country)
      --lat float         latitude (requires --lon)
      --lon float         longitude (requires --lat)
      --once              print a single frame and exit
      --width int         scene width for --once (default 40)
      --height int        scene height for --once (default 12)
      --demo              cycle through all scenes with fake data (no network)
      --ambient           start in fullscreen ambient mode
      --scene string      force a scene: clear|night|cloudy|overcast|fog|drizzle|rain|freezing|snow|storm|hail
      --intensity string  light|moderate|heavy (only with --scene)
      --fps int           animation frames per second (default 12, range 4-30)
      --no-color          disable colors (also honors NO_COLOR env var)
      --ascii             ASCII-only glyphs (no Unicode/emoji)
      --no-cache          don't read cached weather
  -v, --version           print version and exit
  -h, --help              show help
`

// Parse parses args (without the program name). getenv is used to read
// NO_COLOR; pass os.Getenv. Non-fatal warnings (e.g. clamped fps) are
// written to warn. It returns ErrHelp when help was requested and a
// *UsageError for invalid input.
func Parse(args []string, getenv func(string) string, warn io.Writer) (Config, error) {
	var c Config
	fs := pflag.NewFlagSet("skycast", pflag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {}

	var help bool
	fs.StringVarP(&c.Units, "units", "u", "", "")
	fs.Float64Var(&c.Lat, "lat", 0, "")
	fs.Float64Var(&c.Lon, "lon", 0, "")
	fs.BoolVar(&c.Once, "once", false, "")
	fs.IntVar(&c.Width, "width", DefaultWidth, "")
	fs.IntVar(&c.Height, "height", DefaultHeight, "")
	fs.BoolVar(&c.Demo, "demo", false, "")
	fs.BoolVar(&c.Ambient, "ambient", false, "")
	fs.StringVar(&c.Scene, "scene", "", "")
	fs.StringVar(&c.Intensity, "intensity", "", "")
	fs.IntVar(&c.FPS, "fps", DefaultFPS, "")
	fs.BoolVar(&c.NoColor, "no-color", false, "")
	fs.BoolVar(&c.ASCII, "ascii", false, "")
	fs.BoolVar(&c.NoCache, "no-cache", false, "")
	fs.BoolVarP(&c.Version, "version", "v", false, "")
	fs.BoolVarP(&help, "help", "h", false, "")
	fs.BoolVar(&c.DebugFetch, "debug-fetch", false, "")
	_ = fs.MarkHidden("debug-fetch")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			return c, ErrHelp
		}
		return c, &UsageError{Msg: err.Error()}
	}
	if help {
		return c, ErrHelp
	}
	if c.Version {
		return c, nil
	}

	c.City = strings.TrimSpace(strings.Join(fs.Args(), " "))
	c.Units = strings.ToLower(strings.TrimSpace(c.Units))
	c.Scene = strings.ToLower(strings.TrimSpace(c.Scene))
	c.Intensity = strings.ToLower(strings.TrimSpace(c.Intensity))
	if getenv != nil && getenv("NO_COLOR") != "" {
		c.NoColor = true
	}

	latSet, lonSet := fs.Changed("lat"), fs.Changed("lon")
	switch {
	case latSet && !lonSet:
		return c, &UsageError{Msg: "--lat requires --lon"}
	case lonSet && !latSet:
		return c, &UsageError{Msg: "--lon requires --lat"}
	case latSet && lonSet:
		if c.Lat < -90 || c.Lat > 90 {
			return c, &UsageError{Msg: fmt.Sprintf("--lat must be between -90 and 90, got %g", c.Lat)}
		}
		if c.Lon < -180 || c.Lon > 180 {
			return c, &UsageError{Msg: fmt.Sprintf("--lon must be between -180 and 180, got %g", c.Lon)}
		}
		c.HasCoords = true
	}

	switch c.Units {
	case "", "metric", "imperial":
	default:
		return c, &UsageError{Msg: fmt.Sprintf("unknown units %q (want metric or imperial)", c.Units)}
	}
	if c.Scene != "" && !contains(SceneNames, c.Scene) {
		return c, &UsageError{Msg: fmt.Sprintf("unknown scene %q (want one of: %s)", c.Scene, strings.Join(SceneNames, ", "))}
	}
	if c.Intensity != "" {
		if c.Scene == "" {
			return c, &UsageError{Msg: "--intensity requires --scene"}
		}
		if !contains(IntensityNames, c.Intensity) {
			return c, &UsageError{Msg: fmt.Sprintf("unknown intensity %q (want light, moderate or heavy)", c.Intensity)}
		}
	}
	if c.FPS < MinFPS || c.FPS > MaxFPS {
		clamped := clamp(c.FPS, MinFPS, MaxFPS)
		if warn != nil {
			fmt.Fprintf(warn, "skycast: --fps %d out of range %d-%d, using %d\n", c.FPS, MinFPS, MaxFPS, clamped)
		}
		c.FPS = clamped
	}
	if c.Width < 10 || c.Height < 4 {
		return c, &UsageError{Msg: "--width must be at least 10 and --height at least 4"}
	}
	return c, nil
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
