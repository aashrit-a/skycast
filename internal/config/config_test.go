package config

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func noEnv(string) string { return "" }

func TestParseValid(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		check func(t *testing.T, c Config)
	}{
		{"defaults", nil, func(t *testing.T, c Config) {
			if c.FPS != DefaultFPS || c.Width != DefaultWidth || c.Height != DefaultHeight || c.City != "" || c.HasCoords {
				t.Errorf("unexpected defaults: %+v", c)
			}
		}},
		{"positional join", []string{"new", "york"}, func(t *testing.T, c Config) {
			if c.City != "new york" {
				t.Errorf("City = %q", c.City)
			}
		}},
		{"quoted city with units", []string{"san jose, ca", "-u", "Imperial"}, func(t *testing.T, c Config) {
			if c.City != "san jose, ca" || c.Units != "imperial" {
				t.Errorf("got %+v", c)
			}
		}},
		{"coords", []string{"--lat", "48.85", "--lon", "2.35"}, func(t *testing.T, c Config) {
			if !c.HasCoords || c.Lat != 48.85 || c.Lon != 2.35 {
				t.Errorf("got %+v", c)
			}
		}},
		{"scene and intensity", []string{"--scene", "snow", "--intensity", "heavy"}, func(t *testing.T, c Config) {
			if c.Scene != "snow" || c.Intensity != "heavy" {
				t.Errorf("got %+v", c)
			}
		}},
		{"modes", []string{"--once", "--demo", "--ambient", "--ascii", "--no-cache", "--width", "50", "--height", "14"}, func(t *testing.T, c Config) {
			if !c.Once || !c.Demo || !c.Ambient || !c.ASCII || !c.NoCache || c.Width != 50 || c.Height != 14 {
				t.Errorf("got %+v", c)
			}
		}},
		{"version", []string{"-v"}, func(t *testing.T, c Config) {
			if !c.Version {
				t.Error("Version not set")
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := Parse(tt.args, noEnv, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			tt.check(t, c)
		})
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"--lat", "1"}, "--lat requires --lon"},
		{[]string{"--lon", "1"}, "--lon requires --lat"},
		{[]string{"--lat", "100", "--lon", "1"}, "--lat must be"},
		{[]string{"-u", "kelvin"}, "unknown units"},
		{[]string{"--scene", "tornado"}, "unknown scene"},
		{[]string{"--intensity", "heavy"}, "--intensity requires --scene"},
		{[]string{"--scene", "rain", "--intensity", "extreme"}, "unknown intensity"},
		{[]string{"--bogus"}, "unknown flag"},
		{[]string{"--width", "3"}, "--width must be"},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			_, err := Parse(tt.args, noEnv, nil)
			var ue *UsageError
			if !errors.As(err, &ue) {
				t.Fatalf("want UsageError, got %v", err)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error %q does not contain %q", err, tt.want)
			}
		})
	}
}

func TestParseHelp(t *testing.T) {
	for _, a := range []string{"-h", "--help"} {
		if _, err := Parse([]string{a}, noEnv, nil); !errors.Is(err, ErrHelp) {
			t.Errorf("%s: want ErrHelp, got %v", a, err)
		}
	}
}

func TestFPSClamp(t *testing.T) {
	var warn bytes.Buffer
	c, err := Parse([]string{"--fps", "100"}, noEnv, &warn)
	if err != nil {
		t.Fatal(err)
	}
	if c.FPS != MaxFPS {
		t.Errorf("FPS = %d, want %d", c.FPS, MaxFPS)
	}
	if !strings.Contains(warn.String(), "out of range") {
		t.Errorf("missing warning, got %q", warn.String())
	}
	c, _ = Parse([]string{"--fps", "1"}, noEnv, nil)
	if c.FPS != MinFPS {
		t.Errorf("FPS = %d, want %d", c.FPS, MinFPS)
	}
}

func TestNoColorEnv(t *testing.T) {
	env := func(k string) string {
		if k == "NO_COLOR" {
			return "1"
		}
		return ""
	}
	c, err := Parse(nil, env, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !c.NoColor {
		t.Error("NO_COLOR env not honored")
	}
}
