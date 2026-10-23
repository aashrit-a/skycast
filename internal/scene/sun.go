package scene

import (
	"math"
	"math/rand"

	"github.com/aashrit-a/skycast/internal/weather"
)

// sun is the clear-day scene: a pulsing sun with slowly rotating rays.
type sun struct {
	p      Params
	rng    *rand.Rand
	w, h   int
	frame  int
	angle  float64 // ray rotation, radians
	wisps  []Particle
	cx, cy float64
	r      float64
}

func newSun(p Params, rng *rand.Rand) *sun { return &sun{p: p, rng: rng} }

func (s *sun) Name() string { return "clear" }

func (s *sun) Resize(w, h int) {
	s.w, s.h = max(w, 1), max(h, 1)
	s.r = clampF(math.Floor(math.Min(float64(s.w)/2, float64(s.h))/4), 2, 6)
	s.cx, s.cy = float64(s.w)*0.35, float64(s.h)*0.45
	s.wisps = s.wisps[:0]
	if s.p.Intensity == weather.Light && s.h >= 8 {
		n := clampI(s.w/30, 1, 3)
		for i := 0; i < n; i++ {
			s.wisps = append(s.wisps, Particle{
				X:  s.rng.Float64() * float64(s.w),
				Y:  float64(1 + s.rng.Intn(max(1, s.h/4))),
				VX: between(s.rng, 0.02, 0.05),
				// Age holds the wisp length.
				Age: 3 + s.rng.Intn(5),
			})
		}
	}
}

func (s *sun) Step() {
	s.frame++
	s.angle += 4 * math.Pi / 180
	for i := range s.wisps {
		s.wisps[i].X = wrapX(s.wisps[i].X+s.wisps[i].VX, s.w+10)
	}
}

func (s *sun) Draw(c *Canvas) {
	for _, wp := range s.wisps {
		for i := 0; i < wp.Age; i++ {
			c.Set(int(wp.X)-5+i, int(wp.Y), '~', ColCloudLight)
		}
	}
	s.drawRays(c)
	s.drawBody(c)
}

func (s *sun) drawRays(c *Canvas) {
	breath := math.Round(math.Sin(float64(s.frame) * 2 * math.Pi / 40))
	length := s.r + 3 + breath // r+2 .. r+4
	for i := 0; i < 8; i++ {
		a := s.angle + float64(i)*math.Pi/4
		g := rayGlyph(a)
		for d := s.r + 1.2; d <= length; d += 0.5 {
			x := ix(s.cx + math.Cos(a)*d*2)
			y := ix(s.cy - math.Sin(a)*d)
			col := ColSun
			if d > length-1 {
				col = ColSunDim
			}
			c.Set(x, y, g, col)
		}
	}
}

// rayGlyph picks the character that best matches a ray's direction.
func rayGlyph(a float64) rune {
	deg := math.Mod(a*180/math.Pi, 180)
	if deg < 0 {
		deg += 180
	}
	switch {
	case deg < 22.5 || deg >= 157.5:
		return '-'
	case deg < 67.5:
		return '/'
	case deg < 112.5:
		return '|'
	default:
		return '\\'
	}
}

func (s *sun) drawBody(c *Canvas) {
	pulse := math.Sin(float64(s.frame) * 2 * math.Pi / 60)
	body, edge := ColSun, ColSunGlow
	if pulse < 0 {
		body, edge = ColSunGlow, ColSun
	}
	fill, rim := '█', '▓'
	if s.p.ASCIIOnly {
		fill, rim = '#', 'O'
	}
	r := s.r
	for y := int(s.cy - r - 1); y <= int(s.cy+r+1); y++ {
		for x := int(s.cx - 2*r - 2); x <= int(s.cx+2*r+2); x++ {
			dx, dy := (float64(x)-s.cx)/2, float64(y)-s.cy
			d := math.Hypot(dx, dy)
			switch {
			case d <= r-0.8:
				c.Set(x, y, fill, body)
			case d <= r+0.2:
				c.Set(x, y, rim, edge)
			}
		}
	}
}
