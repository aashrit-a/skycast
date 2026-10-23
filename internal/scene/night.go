package scene

import (
	"math"
	"math/rand"
)

type star struct {
	x, y   int
	glyph  rune
	phase  float64
	period float64
}

type shootingStar struct {
	x, y   float64
	age    int
	active bool
}

// night is the clear-night scene: a crescent moon and twinkling stars.
type night struct {
	p      Params
	rng    *rand.Rand
	w, h   int
	frame  int
	stars  []star
	shoot  shootingStar
	cx, cy float64
	r      float64
}

func newNight(p Params, rng *rand.Rand) *night { return &night{p: p, rng: rng} }

func (n *night) Name() string { return "night" }

func (n *night) Resize(w, h int) {
	n.w, n.h = max(w, 1), max(h, 1)
	n.r = clampF(math.Floor(math.Min(float64(n.w)/2, float64(n.h))/3), 2, 6)
	n.cx, n.cy = float64(n.w)*0.68, float64(n.h)*0.38
	glyphs := []rune{'·', '·', '✦', '✧', '·'}
	if n.p.ASCIIOnly {
		glyphs = []rune{'.', '.', '*', '+', '.'}
	}
	n.stars = n.stars[:0]
	count := n.w * n.h / 30
	for tries := 0; len(n.stars) < count && tries < count*10; tries++ {
		x, y := n.rng.Intn(n.w), n.rng.Intn(n.h)
		dx, dy := (float64(x)-n.cx)/2, float64(y)-n.cy
		if math.Hypot(dx, dy) < n.r+1.5 {
			continue
		}
		n.stars = append(n.stars, star{
			x: x, y: y,
			glyph:  glyphs[n.rng.Intn(len(glyphs))],
			phase:  n.rng.Float64() * 2 * math.Pi,
			period: between(n.rng, 30, 120),
		})
	}
	n.shoot = shootingStar{}
}

func (n *night) Step() {
	n.frame++
	if n.shoot.active {
		n.shoot.x += 2
		n.shoot.y++
		n.shoot.age++
		if n.shoot.age > 10 {
			n.shoot.active = false
		}
	} else if n.rng.Intn(400) == 0 && n.w > 20 {
		n.shoot = shootingStar{
			x:      float64(n.rng.Intn(n.w * 2 / 3)),
			y:      float64(n.rng.Intn(max(1, n.h/3))),
			active: true,
		}
	}
}

func (n *night) Draw(c *Canvas) {
	for _, s := range n.stars {
		b := math.Sin(s.phase + float64(n.frame)*2*math.Pi/s.period)
		switch {
		case b > 0.3:
			c.Set(s.x, s.y, s.glyph, ColStar)
		case b > -0.4:
			c.Set(s.x, s.y, s.glyph, ColStarDim)
		}
	}
	if n.shoot.active {
		x, y := int(n.shoot.x), int(n.shoot.y)
		head := '✦'
		if n.p.ASCIIOnly {
			head = '*'
		}
		c.Set(x, y, head, ColStar)
		for i := 1; i <= 3; i++ {
			col := ColStar
			if i > 1 {
				col = ColStarDim
			}
			c.Set(x-2*i, y-i, '\\', col)
		}
	}
	n.drawMoon(c)
}

func (n *night) drawMoon(c *Canvas) {
	fill, rim := '█', '▓'
	if n.p.ASCIIOnly {
		fill, rim = '@', '('
	}
	r := n.r
	// The crescent is the moon disc minus a disc shifted right and up.
	ox, oy := n.cx+r*0.55*2, n.cy-r*0.1
	for y := int(n.cy - r - 1); y <= int(n.cy+r+1); y++ {
		for x := int(n.cx - 2*r - 2); x <= int(n.cx+2*r+2); x++ {
			d := math.Hypot((float64(x)-n.cx)/2, float64(y)-n.cy)
			cut := math.Hypot((float64(x)-ox)/2, float64(y)-oy)
			if d > r+0.2 || cut < r {
				continue
			}
			g := fill
			if d > r-0.8 || cut < r+0.9 {
				g = rim
			}
			c.Set(x, y, g, ColMoon)
		}
	}
}
