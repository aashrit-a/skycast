package scene

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/YOUR_USERNAME/skycast/internal/weather"
)

func TestCanvasBasics(t *testing.T) {
	c := NewCanvas(5, 2)
	c.Set(-1, 0, 'x', ColRain)
	c.Set(5, 0, 'x', ColRain)
	c.Set(0, 2, 'x', ColRain)
	c.DrawString(3, 1, "abcd", ColSun)
	c.Set(0, 0, 'z', ColNone)
	if got := c.Render(Palette{}); got != "z    \n   ab" {
		t.Errorf("Render = %q", got)
	}
	if c.Get(9, 9).R != ' ' || c.Get(3, 1).R != 'a' || c.Get(3, 1).Color != ColSun {
		t.Error("Get mismatch")
	}
	c.Resize(3, 3)
	if len(c.Cells) != 9 || c.Render(Palette{}) != "   \n   \n   " {
		t.Error("Resize did not clear")
	}
	if NewCanvas(-1, -4).Render(Palette{}) != "" {
		t.Error("negative canvas should render empty")
	}
}

func TestRenderRuns(t *testing.T) {
	p := testPalette()
	c := NewCanvas(6, 1)
	c.DrawString(0, 0, "aaa", ColRain)
	c.Set(3, 0, ' ', ColRain) // blank: default style
	c.DrawString(4, 0, "bb", ColSun)
	got := c.Render(p)
	want := "<9>aaa</9><0> </0><1>bb</1>"
	if got != want {
		t.Errorf("Render = %q, want %q", got, want)
	}
}

func TestNoColorPaletteIsPlain(t *testing.T) {
	p := NewPalette(true)
	c := NewCanvas(4, 1)
	c.DrawString(0, 0, "ab", ColRain)
	c.SetBold(2, 0, 'c', ColBolt)
	if got := c.Render(p); got != "abc " {
		t.Errorf("Render = %q", got)
	}
}

// testPalette marks each run with its color id so tests can see grouping.
func testPalette() Palette {
	var p Palette
	for id := 0; id < int(numColors); id++ {
		for b := 0; b < 2; b++ {
			tag := strings.Repeat("!", b) + itoa(id)
			p.prefix[id*2+b], p.suffix[id*2+b] = "<"+tag+">", "</"+tag+">"
		}
	}
	return p
}

func itoa(i int) string {
	if i < 10 {
		return string(rune('0' + i))
	}
	return itoa(i/10) + string(rune('0'+i%10))
}

func ansiPalette() Palette {
	var p Palette
	for i := range p.prefix {
		p.prefix[i] = "\x1b[38;2;95;175;255m"
		p.suffix[i] = "\x1b[0m"
	}
	return p
}

func BenchmarkRender(b *testing.B) {
	s := New(Params{Condition: weather.Rain, Intensity: weather.Heavy, Wind: 0.5}, rand.New(rand.NewSource(1)))
	s.Resize(80, 24)
	c := NewCanvas(80, 24)
	for i := 0; i < 40; i++ {
		s.Step()
	}
	s.Draw(c)
	p := ansiPalette()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = c.Render(p)
	}
}

func BenchmarkRainStep(b *testing.B) {
	s := New(Params{Condition: weather.Rain, Intensity: weather.Heavy, Wind: 0.5}, rand.New(rand.NewSource(1)))
	s.Resize(80, 24)
	c := NewCanvas(80, 24)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s.Step()
		c.Clear()
		s.Draw(c)
	}
}
