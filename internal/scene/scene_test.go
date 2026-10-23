package scene

import (
	"math"
	"math/rand"
	"testing"

	"github.com/YOUR_USERNAME/skycast/internal/weather"
)

// allParams covers every scene variant the factory can build.
func allParams() []Params {
	var out []Params
	conds := []weather.Condition{
		weather.ClearDay, weather.ClearNight, weather.PartlyCloudy, weather.Overcast,
		weather.Fog, weather.Drizzle, weather.Rain, weather.FreezingRain,
		weather.Snow, weather.Thunderstorm, weather.Hail,
	}
	for _, c := range conds {
		for _, in := range []weather.Intensity{weather.Light, weather.Moderate, weather.Heavy} {
			for _, wind := range []float64{0, -0.8, 0.6} {
				for _, ascii := range []bool{false, true} {
					out = append(out, Params{Condition: c, Intensity: in, Wind: wind, IsDay: c != weather.ClearNight, ASCIIOnly: ascii})
				}
			}
		}
	}
	out = append(out, Params{Condition: weather.Snow, SnowGrains: true})
	out = append(out, Params{Condition: weather.PartlyCloudy, IsDay: false})
	return out
}

func TestScenesRunWithoutPanic(t *testing.T) {
	for _, p := range allParams() {
		s := New(p, rand.New(rand.NewSource(1)))
		s.Resize(40, 15)
		c := NewCanvas(40, 15)
		for i := 0; i < 200; i++ {
			s.Step()
			if i%20 == 0 {
				c.Clear()
				s.Draw(c)
			}
		}
		c.Clear()
		s.Draw(c)
		if s.Name() == "" {
			t.Errorf("%+v: empty name", p)
		}
		if out := c.Render(Palette{}); len([]rune(out)) != 40*15+14 {
			t.Errorf("%s: rendered %d runes", s.Name(), len([]rune(out)))
		}
	}
}

func TestUsableBeforeResize(t *testing.T) {
	for _, p := range allParams() {
		s := New(p, rand.New(rand.NewSource(1)))
		s.Step()
		s.Draw(NewCanvas(20, 8))
	}
}

func TestResizeExtremes(t *testing.T) {
	sizes := [][2]int{{120, 40}, {10, 4}, {1, 1}, {0, 0}, {24, 10}, {120, 40}, {3, 2}, {80, 24}}
	for _, p := range allParams() {
		s := New(p, rand.New(rand.NewSource(2)))
		for _, sz := range sizes {
			s.Resize(sz[0], sz[1])
			c := NewCanvas(sz[0], sz[1])
			for i := 0; i < 30; i++ {
				s.Step()
			}
			s.Draw(c)
		}
	}
}

// particleScene is implemented by scenes that expose particles to tests.
type particleScene interface {
	particles() []Particle
	targetCount() int
}

func TestParticleDensityAndBounds(t *testing.T) {
	const w, h = 40, 15
	for _, p := range allParams() {
		s := New(p, rand.New(rand.NewSource(3)))
		s.Resize(w, h)
		for i := 0; i < 200; i++ {
			s.Step()
		}
		ps, ok := unwrap(s).(particleScene)
		if !ok {
			continue
		}
		got, want := len(ps.particles()), ps.targetCount()
		if math.Abs(float64(got-want)) > 0.2*float64(want)+1 {
			t.Errorf("%s %+v: %d particles, want ~%d", s.Name(), p, got, want)
		}
		for _, pt := range ps.particles() {
			if pt.X < 0 || pt.X >= w || pt.Y >= h || math.IsNaN(pt.X) || math.IsNaN(pt.Y) {
				t.Errorf("%s: particle out of bounds: %+v", s.Name(), pt)
				break
			}
		}
	}
}

func TestDeterministic(t *testing.T) {
	render := func() string {
		s := New(Params{Condition: weather.Rain, Intensity: weather.Moderate, Wind: 0.3}, rand.New(rand.NewSource(7)))
		s.Resize(30, 10)
		for i := 0; i < 50; i++ {
			s.Step()
		}
		c := NewCanvas(30, 10)
		s.Draw(c)
		return c.Render(Palette{})
	}
	if render() != render() {
		t.Error("same seed produced different frames")
	}
}

func TestPaletteRendersMarkerFree(t *testing.T) {
	p := NewPalette(false)
	for i := range p.prefix {
		if containsNUL(p.prefix[i]) || containsNUL(p.suffix[i]) {
			t.Fatal("marker leaked into palette")
		}
	}
}

func containsNUL(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == 0 {
			return true
		}
	}
	return false
}

func TestByNameCoversConfig(t *testing.T) {
	names := []string{"clear", "night", "cloudy", "overcast", "fog", "drizzle", "rain", "freezing", "snow", "storm", "hail"}
	for _, n := range names {
		cond, _, _, ok := ByName(n)
		if !ok {
			t.Errorf("ByName(%q) not ok", n)
			continue
		}
		if got := New(Params{Condition: cond}, rand.New(rand.NewSource(1))).Name(); got != n {
			t.Errorf("scene for %q is named %q", n, got)
		}
	}
	if _, _, _, ok := ByName("tornado"); ok {
		t.Error("unknown name accepted")
	}
}
