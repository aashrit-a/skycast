package ui

import (
	"math/rand"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/YOUR_USERNAME/skycast/internal/config"
	"github.com/YOUR_USERNAME/skycast/internal/scene"
)

func TestDemoCoversEveryScene(t *testing.T) {
	seen := map[string]bool{}
	for i := range demoScenes {
		r := demoReport(i, 0, fixedNow)
		m := Model{cfg: config.Config{}}
		p := m.sceneParams(r)
		seen[scene.New(p, rand.New(rand.NewSource(1))).Name()] = true
		if len(r.Daily) != 4 || r.Location != "Demo City" {
			t.Errorf("demo %d: bad report", i)
		}
	}
	for _, n := range config.SceneNames {
		if !seen[n] {
			t.Errorf("demo never shows %q", n)
		}
	}
	if len(demoScenes) != 11 {
		t.Errorf("want 11 demo scenes, got %d", len(demoScenes))
	}
}

func TestDemoNavigation(t *testing.T) {
	now := fixedNow
	m := New(Options{Config: config.Config{Demo: true, FPS: 12}, Palette: scene.NewPalette(true), Now: func() time.Time { return now }})
	if m.Init() == nil || m.state != stateReady {
		t.Fatal("demo should start ready")
	}
	m, _ = update(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	if v := m.View(); !strings.Contains(v, "demo · 1/11 · Clear sky") || !strings.Contains(v, "Demo City") {
		t.Errorf("footer/panel missing:\n%s", v)
	}

	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyRight})
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	if m.demo.idx != 2 || m.scene.Name() != "cloudy" {
		t.Errorf("idx=%d scene=%s", m.demo.idx, m.scene.Name())
	}
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyLeft})
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	if m.demo.idx != 10 || m.scene.Name() != "hail" {
		t.Errorf("wrap: idx=%d scene=%s", m.demo.idx, m.scene.Name())
	}

	// Manual navigation holds auto-advance for 15s.
	now = now.Add(10 * time.Second)
	m, _ = update(t, m, tickMsg(now))
	if m.demo.idx != 10 {
		t.Error("advanced during manual hold")
	}
	now = now.Add(6 * time.Second)
	m, _ = update(t, m, tickMsg(now))
	if m.demo.idx != 0 {
		t.Errorf("should auto-advance after hold, idx=%d", m.demo.idx)
	}
	now = now.Add(5 * time.Second)
	m, _ = update(t, m, tickMsg(now))
	if m.demo.idx != 0 {
		t.Error("advanced too early")
	}
	now = now.Add(time.Second)
	m, _ = update(t, m, tickMsg(now))
	if m.demo.idx != 1 {
		t.Errorf("should advance every 6s, idx=%d", m.demo.idx)
	}

	m, _ = update(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("u")})
	if !strings.Contains(m.View(), "°F") {
		t.Error("u should switch demo to imperial")
	}
	if _, cmd := update(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")}); cmd != nil {
		t.Error("r must not fetch in demo mode")
	}
}

func TestAmbientFlag(t *testing.T) {
	m := New(Options{Config: config.Config{Demo: true, Ambient: true, FPS: 12}, Palette: scene.NewPalette(true)})
	m, _ = update(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	if m.layout.Mode != Ambient {
		t.Errorf("mode = %v", m.layout.Mode)
	}
	if v := m.View(); strings.Contains(v, "q quit") || strings.Count(v, "\n") != 23 {
		t.Error("ambient should be scene only, full height")
	}
}
