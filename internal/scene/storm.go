package scene

import (
	"math"
	"math/rand"

	"github.com/YOUR_USERNAME/skycast/internal/weather"
)

type boltCell struct {
	x, y int
	g    rune
}

// storm is heavy rain under a dark, roiling band with lightning; the hail
// variant adds bouncing hailstones.
type storm struct {
	p      Params
	rng    *rand.Rand
	rain   *rain
	bolt   []boltCell
	strike int  // frames into the current strike, 0 = none
	double bool // re-flash on frame 3
	hail   []Particle
	w, h   int
}

func newStorm(p Params, rng *rand.Rand) *storm {
	rp := p
	rp.Condition = weather.Rain
	rp.Intensity = weather.Heavy
	if math.Abs(rp.Wind) < 0.3 {
		rp.Wind = math.Copysign(0.3, rp.Wind+1e-9)
	}
	r := newRain(rp, rng)
	r.bandSpeed = 0.2
	r.bandRows = func(h int) int { return clampI(h/4, 3, 5) }
	return &storm{p: p, rng: rng, rain: r}
}

func (s *storm) Name() string {
	if s.p.Condition == weather.Hail {
		return "hail"
	}
	return "storm"
}

func (s *storm) hailTarget() int {
	if s.p.Condition != weather.Hail {
		return 0
	}
	area := s.w * max(0, s.h-1-int(s.rain.top()))
	return density(area, byIntensity(s.p.Intensity, 0.008, 0.012, 0.02))
}

func (s *storm) Resize(w, h int) {
	s.w, s.h = max(w, 1), max(h, 1)
	s.rain.Resize(w, h)
	s.rain.band.dense, s.rain.band.thin = ColCloudDark, ColCloudDark
	s.bolt, s.strike = nil, 0
	s.hail = s.hail[:0]
	for i, n := 0, s.hailTarget(); i < n; i++ {
		p := s.newHail()
		p.Y = between(s.rng, s.rain.top(), float64(s.rain.ground()))
		s.hail = append(s.hail, p)
	}
}

func (s *storm) newHail() Particle {
	p := spawnTop(s.rng, s.w, s.rain.top())
	p.VY = between(s.rng, 1.6, 2.0)
	p.VX = s.rain.p.Wind * 0.5
	p.Glyph, p.Color = 'o', ColIce
	if !s.p.ASCIIOnly && s.rng.Intn(3) == 0 {
		p.Glyph = '•'
	}
	return p
}

func (s *storm) Step() {
	s.rain.Step()
	s.stepHail()
	if s.strike > 0 {
		s.strike++
		if s.strike > 4 {
			s.strike, s.bolt = 0, nil
		}
	} else if s.rng.Intn(60) == 0 {
		s.startStrike()
	}
	s.rain.band.flash = s.strike == 1 || (s.double && s.strike == 3)
}

func (s *storm) stepHail() {
	ground := float64(s.rain.ground()) - 1
	for i := range s.hail {
		p := &s.hail[i]
		p.X = wrapX(p.X+p.VX, s.w)
		p.Y += p.VY
		if p.Age > 0 { // bounced: gravity pulls it back down
			p.VY += 0.3
		}
		if p.Y >= ground {
			if p.Age == 0 {
				p.Age = 1
				p.Y = ground
				p.VY = -0.6 * p.VY
				p.VX += between(s.rng, -0.6, 0.6)
				continue
			}
			s.hail[i] = s.newHail()
		}
	}
}

func (s *storm) startStrike() {
	top := s.rain.band.rows - 1
	ground := s.rain.ground()
	if ground-top < 3 {
		return
	}
	s.strike, s.double, s.bolt = 1, s.rng.Float64() < 0.3, s.bolt[:0]
	x := s.w/6 + s.rng.Intn(max(1, s.w*2/3))
	for y := top + 1; y < ground; y++ {
		dx := s.rng.Intn(3) - 1
		x += dx
		s.bolt = append(s.bolt, boltCell{x: x, y: y, g: stepGlyph(dx)})
	}
	if len(s.bolt) > 3 && s.rng.Float64() < 0.2 {
		// A short side branch heading away diagonally.
		start := s.bolt[1+s.rng.Intn(len(s.bolt)-2)]
		bdx := 1
		if s.rng.Intn(2) == 0 {
			bdx = -1
		}
		bx := start.x
		for by, n := start.y+1, 2+s.rng.Intn(3); n > 0 && by < ground; n, by = n-1, by+1 {
			bx += bdx
			s.bolt = append(s.bolt, boltCell{x: bx, y: by, g: stepGlyph(bdx)})
		}
	}
}

// stepGlyph returns the glyph for a bolt segment moving dx per row downward.
func stepGlyph(dx int) rune {
	switch dx {
	case -1:
		return '/'
	case 1:
		return '\\'
	}
	return '|'
}

func (s *storm) Draw(c *Canvas) {
	s.rain.Draw(c)
	for _, p := range s.hail {
		y := ix(p.Y)
		if y >= int(s.rain.top()) && y < s.rain.ground() {
			c.Set(int(p.X), y, p.Glyph, p.Color)
		}
	}
	if s.rain.band.flash {
		s.flashBand(c)
	}
	for _, b := range s.bolt {
		if s.strike == 4 {
			c.Set(b.x, b.y, b.g, ColStarDim)
		} else {
			c.SetBold(b.x, b.y, b.g, ColBolt)
		}
	}
}

// flashBand lights up the gaps in the cloud band during a flash.
func (s *storm) flashBand(c *Canvas) {
	g := s.rain.band.ramp[1]
	for y := 0; y < s.rain.band.rows; y++ {
		for x := 0; x < c.W; x++ {
			if c.IsEmpty(x, y) {
				c.SetBold(x, y, g, ColBolt)
			}
		}
	}
}

func (s *storm) particles() []Particle {
	return append(append([]Particle(nil), s.rain.drops...), s.hail...)
}

func (s *storm) targetCount() int { return s.rain.target() + s.hailTarget() }
