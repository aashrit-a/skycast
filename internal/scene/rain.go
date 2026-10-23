package scene

import (
	"math"
	"math/rand"

	"github.com/aashrit-a/skycast/internal/weather"
)

type splash struct {
	x, age int
}

// rain covers drizzle, rain, showers and freezing rain.
type rain struct {
	p        Params
	rng      *rand.Rand
	w, h     int
	band     *cloudBand
	drops    []Particle
	splashes []splash
	frame    int
	// The storm scene reuses heavy rain with a deeper, faster band.
	bandSpeed float64
	bandRows  func(h int) int
}

func newRain(p Params, rng *rand.Rand) *rain {
	return &rain{p: p, rng: rng, bandSpeed: 0.08}
}

func (r *rain) Name() string {
	switch r.p.Condition {
	case weather.Drizzle:
		return "drizzle"
	case weather.FreezingRain:
		return "freezing"
	}
	return "rain"
}

func (r *rain) drizzle() bool { return r.p.Condition == weather.Drizzle }
func (r *rain) freezing() bool {
	return r.p.Condition == weather.FreezingRain
}

// top is the first row below the cloud band; ground is the last row.
func (r *rain) top() float64 { return float64(r.band.rows) }
func (r *rain) ground() int  { return r.h - 1 }

func (r *rain) target() int {
	area := r.w * max(0, r.ground()-int(r.top()))
	f := byIntensity(r.p.Intensity, 0.04, 0.07, 0.12)
	if r.drizzle() {
		f = 0.03
	}
	return density(area, f)
}

func (r *rain) Resize(w, h int) {
	r.w, r.h = max(w, 1), max(h, 1)
	rows := clampI(r.h/4, 2, 4)
	if r.bandRows != nil {
		rows = r.bandRows(r.h)
	}
	if r.h < 6 {
		rows = 1
	}
	r.band = newCloudBand(r.rng, rows, ColCloudDark, ColStarDim, r.bandSpeed, r.p.ASCIIOnly)
	r.splashes = r.splashes[:0]
	n := r.target()
	r.drops = r.drops[:0]
	for i := 0; i < n; i++ {
		d := r.newDrop()
		// Scatter initial drops through the whole column of air.
		d.Y = between(r.rng, r.top(), float64(r.ground()))
		r.drops = append(r.drops, d)
	}
}

func (r *rain) newDrop() Particle {
	d := spawnTop(r.rng, r.w, r.top())
	var lo, hi float64
	switch {
	case r.drizzle():
		lo, hi = 0.5, 0.8
	default:
		lo, hi = byIntensity(r.p.Intensity, 0.8, 1.0, 1.3), byIntensity(r.p.Intensity, 1.1, 1.4, 1.8)
	}
	d.VY = between(r.rng, lo, hi)
	d.VX = r.p.Wind * 1.2 * d.VY
	// Faster drops are "closer" and brighter: a cheap depth effect.
	d.Color = ColRainLight
	if d.VY > (lo+hi)/2 {
		d.Color = ColRain
	}
	if r.freezing() {
		d.Color = ColIce
	}
	d.Glyph = r.dropGlyph(d)
	return d
}

func (r *rain) dropGlyph(d Particle) rune {
	if r.drizzle() {
		if r.rng.Intn(2) == 0 {
			return '\''
		}
		return '.'
	}
	switch {
	case math.Abs(d.VX/d.VY) < 0.25:
		return '|'
	case d.VX > 0:
		return '\\'
	default:
		return '/'
	}
}

func (r *rain) Step() {
	r.frame++
	r.band.step(r.p.Wind)
	// Age splashes and drop the finished ones.
	kept := r.splashes[:0]
	for _, s := range r.splashes {
		s.age++
		if s.age < 3 {
			kept = append(kept, s)
		}
	}
	r.splashes = kept

	ground := float64(r.ground())
	for i := range r.drops {
		d := &r.drops[i]
		d.X += d.VX
		d.Y += d.VY
		d.Age++
		if respawnIfOut(d, r.w, ground) {
			if !r.drizzle() || r.rng.Intn(3) == 0 {
				r.splashes = append(r.splashes, splash{x: int(d.X)})
			}
			*d = r.newDrop()
		}
	}
}

func (r *rain) Draw(c *Canvas) {
	r.drawGround(c)
	for _, d := range r.drops {
		x, y := int(d.X), ix(d.Y)
		if y < int(r.top()) || y >= r.ground() {
			continue
		}
		// Heavy, fast drops leave a faint streak above them.
		if r.p.Intensity == weather.Heavy && !r.drizzle() && d.Color != ColRainLight {
			tx, ty := int(wrapX(d.X-d.VX*0.6, r.w)), ix(d.Y-d.VY*0.6)
			if ty >= int(r.top()) && ty != y && c.IsEmpty(tx, ty) {
				c.Set(tx, ty, d.Glyph, ColRainLight)
			}
		}
		c.Set(x, y, d.Glyph, d.Color)
	}
	r.drawSplashes(c)
	r.band.draw(c)
}

func (r *rain) drawGround(c *Canvas) {
	if r.h < 4 {
		return
	}
	g := r.ground()
	glyph := '▁'
	if r.p.ASCIIOnly {
		glyph = '_'
	}
	for x := 0; x < r.w; x++ {
		if r.freezing() && (x*7+r.frame/4)%5 == 0 {
			c.Set(x, g, '^', ColIce)
			continue
		}
		c.Set(x, g, glyph, ColGround)
	}
}

func (r *rain) drawSplashes(c *Canvas) {
	y := r.ground() - 1
	seq := []rune{'.', 'o', '.'}
	if !r.p.ASCIIOnly {
		seq = []rune{'·', '°', '·'}
	}
	for _, s := range r.splashes {
		g := seq[s.age]
		col := ColSplash
		if r.freezing() {
			g, col = '*', ColIce
		}
		c.Set(s.x, y, g, col)
	}
}

func (r *rain) particles() []Particle { return r.drops }
func (r *rain) targetCount() int      { return r.target() }
