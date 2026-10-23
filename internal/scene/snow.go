package scene

import (
	"math"
	"math/rand"
)

// snow covers snow, snow showers and snow grains, with flakes slowly piling
// up on the ground.
type snow struct {
	p      Params
	rng    *rand.Rand
	w, h   int
	band   *cloudBand
	flakes []Particle
	// heights holds accumulated snow per column in half-row levels.
	heights   []int
	maxLevels int
}

func newSnow(p Params, rng *rand.Rand) *snow { return &snow{p: p, rng: rng} }

func (s *snow) Name() string { return "snow" }

func (s *snow) top() float64 { return float64(s.band.rows) }

func (s *snow) target() int {
	area := s.w * max(0, s.h-1-int(s.top()))
	return density(area, byIntensity(s.p.Intensity, 0.025, 0.05, 0.09))
}

func (s *snow) Resize(w, h int) {
	s.w, s.h = max(w, 1), max(h, 1)
	rows := clampI(s.h/5, 1, 3)
	s.band = newCloudBand(s.rng, rows, ColCloudLight, ColSnowDim, 0.04, s.p.ASCIIOnly)
	s.heights = make([]int, s.w)
	// Up to 3 rows of snow, but never more than a fifth of the scene.
	s.maxLevels = clampI(s.h/5, 0, 3) * 2
	s.flakes = s.flakes[:0]
	for i, n := 0, s.target(); i < n; i++ {
		f := s.newFlake()
		f.Y = between(s.rng, s.top(), float64(s.h-1))
		s.flakes = append(s.flakes, f)
	}
}

func (s *snow) newFlake() Particle {
	f := spawnTop(s.rng, s.w, s.top())
	if s.p.SnowGrains {
		f.Glyph, f.Color, f.VY = '.', ColSnow, between(s.rng, 0.45, 0.55)
		return f
	}
	big, small := []rune{'❄', '❅'}, []rune{'·', '•'}
	if s.p.ASCIIOnly {
		big, small = []rune{'*', '+'}, []rune{'.', '.'}
	}
	// Bigger flakes are closer: brighter and slightly faster.
	if s.rng.Float64() < 0.45 {
		f.Glyph, f.Color, f.VY = big[s.rng.Intn(len(big))], ColSnow, between(s.rng, 0.25, 0.4)
	} else {
		f.Glyph, f.Color, f.VY = small[s.rng.Intn(len(small))], ColSnowDim, between(s.rng, 0.15, 0.28)
	}
	return f
}

// surface returns the row a flake in column x lands on.
func (s *snow) surface(x int) float64 {
	return float64(s.h-1) - float64(s.heights[x]/2)
}

func (s *snow) Step() {
	s.band.step(s.p.Wind)
	for i := range s.flakes {
		f := &s.flakes[i]
		f.Age++
		sway := math.Sin(f.Phase+float64(f.Age)*0.15) * 0.3
		if s.p.SnowGrains {
			sway *= 0.3
		}
		f.X += sway + s.p.Wind*0.6
		f.Y += f.VY
		f.X = wrapX(f.X, s.w)
		col := min(int(f.X), s.w-1)
		if f.Y >= s.surface(col) {
			if s.rng.Float64() < 0.2 {
				s.accumulate(col)
			}
			*f = s.newFlake()
		}
	}
}

// accumulate adds a level at column x, rolling downhill so neighbouring
// columns never differ by more than one level.
func (s *snow) accumulate(x int) {
	for i := 0; i < s.w; i++ {
		switch {
		case x > 0 && s.heights[x-1] < s.heights[x]:
			x--
		case x < s.w-1 && s.heights[x+1] < s.heights[x]:
			x++
		default:
			if s.heights[x] < s.maxLevels {
				s.heights[x]++
			}
			return
		}
	}
}

func (s *snow) Draw(c *Canvas) {
	s.drawGround(c)
	for _, f := range s.flakes {
		y := ix(f.Y)
		if y >= int(s.top()) {
			c.Set(int(f.X), y, f.Glyph, f.Color)
		}
	}
	s.band.draw(c)
}

func (s *snow) drawGround(c *Canvas) {
	if s.h < 4 {
		return
	}
	for x := 0; x < s.w; x++ {
		lv := s.heights[x]
		if lv == 0 {
			g := '▁'
			if s.p.ASCIIOnly {
				g = '_'
			}
			c.Set(x, s.h-1, g, ColSnowDim)
			continue
		}
		for row := 0; row*2 < lv; row++ {
			full := lv-row*2 >= 2
			var g rune
			switch {
			case s.p.ASCIIOnly && lv == 1:
				g = '_'
			case s.p.ASCIIOnly && full:
				g = '#'
			case s.p.ASCIIOnly:
				g = '='
			case full:
				g = '█'
			default:
				g = '▄'
			}
			c.Set(x, s.h-1-row, g, ColSnow)
		}
	}
}

func (s *snow) particles() []Particle { return s.flakes }
func (s *snow) targetCount() int      { return s.target() }
