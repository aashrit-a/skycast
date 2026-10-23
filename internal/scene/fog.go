package scene

import (
	"math/rand"

	"github.com/YOUR_USERNAME/skycast/internal/weather"
)

type fogBand struct {
	y0, y1 int     // rows covered (inclusive, exclusive)
	offset float64 // horizontal scroll
	speed  float64
	color  ColorID
}

type sparkle struct{ x, y, age int }

// fog draws drifting horizontal bands of noise-shaded mist.
type fog struct {
	p        Params
	rng      *rand.Rand
	w, h     int
	n        *noise
	bands    []fogBand
	t        float64
	sparkles []sparkle
	ramp     []rune
	floor    float64 // density below which a cell is clear
}

func newFog(p Params, rng *rand.Rand) *fog {
	f := &fog{p: p, rng: rng, n: newNoise(rng), floor: 0.25}
	if p.ASCIIOnly {
		f.ramp = []rune{' ', '.', ':', '-', '='}
	} else {
		f.ramp = []rune{' ', '░', '░', '▒', '▒'}
	}
	if p.Intensity == weather.Heavy {
		// Rime fog: thicker, with the densest glyph available.
		f.floor = 0.15
		if !p.ASCIIOnly {
			f.ramp = append(f.ramp, '▓')
		}
	}
	return f
}

func (f *fog) Name() string { return "fog" }

func (f *fog) heavy() bool { return f.p.Intensity == weather.Heavy }

func (f *fog) Resize(w, h int) {
	f.w, f.h = max(w, 1), max(h, 1)
	n := clampI(f.h/4, 4, 6)
	f.bands = f.bands[:0]
	colors := []ColorID{ColFog3, ColFog3, ColFog2, ColFog2, ColFog1, ColFog1}
	for i := 0; i < n; i++ {
		y0, y1 := i*f.h/n, (i+1)*f.h/n
		f.bands = append(f.bands, fogBand{
			y0: y0, y1: max(y1, y0+1),
			offset: f.rng.Float64() * 200,
			// Front (lower) bands drift faster: parallax.
			speed: 0.05 + 0.15*float64(i)/float64(max(1, n-1)),
			color: colors[i*len(colors)/n],
		})
	}
	f.sparkles = f.sparkles[:0]
}

func (f *fog) Step() {
	f.t += 0.015
	dir := 1.0
	if f.p.Wind < 0 {
		dir = -1
	}
	for i := range f.bands {
		f.bands[i].offset -= f.bands[i].speed * dir
	}
	if f.heavy() {
		kept := f.sparkles[:0]
		for _, s := range f.sparkles {
			if s.age++; s.age < 6 {
				kept = append(kept, s)
			}
		}
		f.sparkles = kept
		if f.rng.Intn(4) == 0 {
			f.sparkles = append(f.sparkles, sparkle{x: f.rng.Intn(f.w), y: f.rng.Intn(f.h)})
		}
	}
}

func (f *fog) Draw(c *Canvas) {
	last := float64(len(f.ramp) - 1)
	for bi, b := range f.bands {
		for y := b.y0; y < b.y1; y++ {
			for x := 0; x < f.w; x++ {
				// Each band is its own strip of noise; t makes it breathe.
				d := f.n.fbm((float64(x)+b.offset)*0.07, float64(bi)*3.7+float64(y-b.y0)*0.4+f.t)
				i := int((d - f.floor) * last * 1.7)
				if i <= 0 {
					continue
				}
				c.Set(x, y, f.ramp[min(i, len(f.ramp)-1)], b.color)
			}
		}
	}
	for _, s := range f.sparkles {
		c.Set(s.x, s.y, '*', ColIce)
	}
}
