package ui

import (
	"context"
	"errors"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/aashrit-a/skycast/internal/config"
	"github.com/aashrit-a/skycast/internal/location"
	"github.com/aashrit-a/skycast/internal/scene"
	"github.com/aashrit-a/skycast/internal/weather"
)

var fixedNow = time.Date(2026, 10, 22, 14, 5, 0, 0, time.UTC)

func sampleReport(u weather.Units) *weather.Report {
	loc := time.FixedZone("", 9*3600)
	day := func(i, code int, hi, lo float64, pc int) weather.DailyForecast {
		d := time.Date(2026, 10, 22+i, 0, 0, 0, 0, loc)
		return weather.DailyForecast{
			Date: d, Code: code, High: hi, Low: lo, PrecipChance: pc,
			Sunrise: d.Add(5*time.Hour + 58*time.Minute), Sunset: d.Add(17*time.Hour + 2*time.Minute),
		}
	}
	return &weather.Report{
		Location: "Tokyo, Japan",
		Units:    u,
		Current: weather.Current{
			Temp: 18.2, FeelsLike: 16.4, Humidity: 82, WindSpeed: 12.3, WindDirDeg: 45,
			Precip: 0.4, IsDay: true, Code: 61,
		},
		Daily:     []weather.DailyForecast{day(0, 61, 19, 14, 85), day(1, 3, 21.3, 14, 40), day(2, 2, 19.4, 12.1, 10), day(3, 0, 23, 15.2, 0)},
		FetchedAt: fixedNow,
	}
}

type fakeLoader struct {
	calls []LoadRequest
	err   error
}

func (f *fakeLoader) load(_ context.Context, req LoadRequest) (*weather.Report, error) {
	f.calls = append(f.calls, req)
	if f.err != nil {
		return nil, f.err
	}
	u := req.Units
	if req.AutoUnits {
		u = weather.UnitsMetric
	}
	return sampleReport(u), nil
}

func newTestModel(cfg config.Config, fl *fakeLoader) Model {
	if cfg.FPS == 0 {
		cfg.FPS = 12
	}
	return New(Options{
		Config:  cfg,
		Palette: scene.NewPalette(true),
		Loader:  fl.load,
		Rand:    rand.New(rand.NewSource(1)),
		Now:     func() time.Time { return fixedNow },
	})
}

func update(t *testing.T, m Model, msg tea.Msg) (Model, tea.Cmd) {
	t.Helper()
	nm, cmd := m.Update(msg)
	return nm.(Model), cmd
}

// run executes a command and feeds back its message (one level deep;
// batches are expanded).
func run(t *testing.T, m Model, cmd tea.Cmd) Model {
	t.Helper()
	if cmd == nil {
		return m
	}
	msg := cmd()
	if b, ok := msg.(tea.BatchMsg); ok {
		for _, c := range b {
			if c == nil {
				continue
			}
			// Skip timers: they would block.
			done := make(chan tea.Msg, 1)
			go func(c tea.Cmd) { done <- c() }(c)
			select {
			case msg := <-done:
				if _, ok := msg.(weatherMsg); ok {
					m, _ = update(t, m, msg)
				} else if _, ok := msg.(errMsg); ok {
					m, _ = update(t, m, msg)
				}
			case <-time.After(200 * time.Millisecond):
			}
		}
		return m
	}
	m, _ = update(t, m, msg)
	return m
}

func checkFrame(t *testing.T, m Model) string {
	t.Helper()
	v := m.View()
	lines := strings.Split(v, "\n")
	if len(lines) != m.height {
		t.Fatalf("%dx%d: view has %d lines", m.width, m.height, len(lines))
	}
	for i, l := range lines {
		if w := lipgloss.Width(l); w > m.width {
			t.Fatalf("%dx%d: line %d is %d wide: %q", m.width, m.height, i, w, l)
		}
	}
	return v
}

func TestLoadingToReady(t *testing.T) {
	fl := &fakeLoader{}
	m := newTestModel(config.Config{City: "tokyo"}, fl)
	m, _ = update(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	if v := checkFrame(t, m); !strings.Contains(v, "fetching weather for tokyo") {
		t.Errorf("loading view: %q", v)
	}
	m = run(t, m, m.Init())
	if m.state != stateReady || m.scene == nil || m.scene.Name() != "rain" {
		t.Fatalf("state=%v scene=%v", m.state, m.scene)
	}
	if !fl.calls[0].AutoUnits {
		t.Error("first load should auto-pick units")
	}
	v := checkFrame(t, m)
	for _, want := range []string{"Tokyo, Japan", "Slight rain", "18°C", "feels 16°C", "82%", "12 km/h NE", "05:58", "17:02", "Fri", "Sat", "Sun", "21/14°", "40%", "q quit"} {
		if !strings.Contains(v, want) {
			t.Errorf("view missing %q", want)
		}
	}
}

func TestResizeThroughAllModes(t *testing.T) {
	m := newTestModel(config.Config{}, &fakeLoader{})
	m = run(t, m, m.Init())
	sizes := [][2]int{{200, 50}, {100, 30}, {72, 24}, {71, 24}, {50, 30}, {50, 15}, {30, 10}, {5, 2}, {1, 1}, {80, 9}, {120, 40}, {40, 22}}
	for _, panel := range []bool{true, false} {
		for _, amb := range []bool{false, true} {
			m.showPanel, m.ambient = panel, amb
			for _, sz := range sizes {
				m, _ = update(t, m, tea.WindowSizeMsg{Width: sz[0], Height: sz[1]})
				for i := 0; i < 3; i++ {
					m, _ = update(t, m, tickMsg(fixedNow))
				}
				checkFrame(t, m)
			}
		}
	}
	m.showPanel, m.ambient = true, false
	m, _ = update(t, m, tea.WindowSizeMsg{Width: 50, Height: 15})
	m.relayout()
	if v := m.View(); !strings.Contains(v, "Tokyo, Japan · 18°C · Slight rain · 12 km/h NE") {
		t.Errorf("minimal view: %q", v)
	}
}

func TestKeys(t *testing.T) {
	fl := &fakeLoader{}
	m := newTestModel(config.Config{}, fl)
	m, _ = update(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m = run(t, m, m.Init())

	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	if m.layout.Mode != SceneOnly {
		t.Errorf("p: mode %v", m.layout.Mode)
	}
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	if m.layout.Mode != Ambient || strings.Contains(m.View(), "q quit") {
		t.Errorf("a: mode %v", m.layout.Mode)
	}
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	if strings.Contains(m.View(), "r refresh") || !strings.Contains(m.View(), "? help") {
		t.Error("? should collapse the footer")
	}

	var cmd tea.Cmd
	m, cmd = update(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("u")})
	m = run(t, m, cmd)
	if last := fl.calls[len(fl.calls)-1]; last.Units != weather.UnitsImperial || last.AutoUnits {
		t.Errorf("u: request %+v", last)
	}
	if !strings.Contains(m.View(), "°F") {
		t.Error("u: view not imperial")
	}

	m, cmd = update(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	m = run(t, m, cmd)
	if last := fl.calls[len(fl.calls)-1]; !last.BypassCache || last.Units != weather.UnitsImperial {
		t.Errorf("r: request %+v", last)
	}

	for _, k := range []tea.KeyMsg{{Type: tea.KeyRunes, Runes: []rune("q")}, {Type: tea.KeyEsc}, {Type: tea.KeyCtrlC}} {
		_, cmd = update(t, m, k)
		if cmd == nil {
			t.Fatalf("%v: no command", k)
		}
		if _, ok := cmd().(tea.QuitMsg); !ok {
			t.Errorf("%v did not quit", k)
		}
	}
}

func TestQuitWhileLoadingCancels(t *testing.T) {
	m := newTestModel(config.Config{}, &fakeLoader{})
	_, cmd := update(t, m, tea.KeyMsg{Type: tea.KeyCtrlC})
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("ctrl+c while loading should quit")
	}
	if m.ctx.Err() == nil {
		t.Error("quitting should cancel in-flight requests")
	}
}

func TestErrorScreenAndRetry(t *testing.T) {
	fl := &fakeLoader{err: errors.New(`could not find a place called "atlantys"`)}
	m := newTestModel(config.Config{City: "atlantys"}, fl)
	m, _ = update(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m = run(t, m, m.Init())
	if m.state != stateError {
		t.Fatalf("state = %v", m.state)
	}
	v := checkFrame(t, m)
	if !strings.Contains(v, "atlantys") || !strings.Contains(v, "r retry") {
		t.Errorf("error view: %q", v)
	}
	fl.err = nil
	m, cmd := update(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	if m.state != stateLoading {
		t.Errorf("retry should show loading, got %v", m.state)
	}
	m = run(t, m, cmd)
	if m.state != stateReady {
		t.Errorf("after retry state = %v", m.state)
	}
}

func TestNotFoundHint(t *testing.T) {
	m := newTestModel(config.Config{}, &fakeLoader{err: location.ErrLocationNotFound})
	m, _ = update(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m = run(t, m, m.Init())
	if !strings.Contains(m.View(), "add a country") {
		t.Error("missing not-found hint")
	}
}

func TestRefreshFailureKeepsReport(t *testing.T) {
	fl := &fakeLoader{}
	m := newTestModel(config.Config{}, fl)
	m, _ = update(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m = run(t, m, m.Init())
	sc := m.scene
	fl.err = weather.ErrUnreachable
	_, cmd := update(t, m, refreshMsg{})
	m = run(t, m, cmd)
	if m.state != stateReady || !m.report.Stale {
		t.Fatalf("state=%v stale=%v", m.state, m.report.Stale)
	}
	if !strings.Contains(m.View(), "offline · updated just now") {
		t.Error("missing offline line")
	}
	if m.scene != sc {
		t.Error("scene should keep running when the weather didn't change")
	}
}

func TestForcedScene(t *testing.T) {
	m := newTestModel(config.Config{Scene: "snow", Intensity: "heavy"}, &fakeLoader{})
	m, _ = update(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m = run(t, m, m.Init())
	if m.scene.Name() != "snow" || m.params.Intensity != weather.Heavy {
		t.Errorf("scene = %s %+v", m.scene.Name(), m.params)
	}
	if !strings.Contains(m.View(), "Slight rain") {
		t.Error("panel should still show the real weather")
	}
}

func TestMain(m *testing.M) {
	// Keep the user's real cache out of tests.
	dir, _ := os.MkdirTemp("", "skycast-test")
	_ = os.Setenv("XDG_CACHE_HOME", filepath.Join(dir, "cache"))
	code := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}
