package scene

import (
	"math"
	"math/rand"

	"github.com/YOUR_USERNAME/skycast/internal/weather"
)

// Cloud sprites. Leading/trailing spaces on each row are transparent;
// spaces between the outline are opaque so clouds occlude what's behind.
var cloudSprites = [][]string{
	{
		"   .--.    ",
		".-(    ).  ",
		"(___.__)__)",
	},
	{
		"     .-~~~-.   ",
		" .-~~       ~-.",
		"(_____________)",
	},
	{
		"      .--~~--.      ",
		"  .--(        )--.  ",
		" (                ) ",
		"  `--.________.--'  ",
	},
	{
		" .-~-. ",
		"(_____)",
	},
}

var maxSpriteW = func() int {
	m := 0
	for _, s := range cloudSprites {
		m = max(m, len(s[0]))
	}
	return m
}()

type cloud struct {
	x      float64 // position along the layer's ring
	y      int
	sprite int
	speed  float64 // multiplier on the layer speed
}

// cloudLayer is a ring of clouds that scrolls horizontally and wraps
// seamlessly: a cloud leaving one edge re-enters at the other.
type cloudLayer struct {
	clouds []cloud
	speed  float64 // cells per frame, before wind
	color  ColorID
	ring   float64 // ring length >= w + maxSpriteW
}

// newCloudLayer spreads clouds evenly across width w with spacing roughly
// `spacing` cells, at rows yMin..yMax (inclusive), using the given sprites.
func newCloudLayer(rng *rand.Rand, w, yMin, yMax int, spacing, speed float64, color ColorID, sprites []int) *cloudLayer {
	ring := float64(w + maxSpriteW)
	n := max(1, int(math.Round(ring/spacing)))
	l := &cloudLayer{speed: speed, color: color, ring: ring}
	step := ring / float64(n)
	for i := 0; i < n; i++ {
		y := yMin
		if yMax > yMin {
			y += rng.Intn(yMax - yMin + 1)
		}
		l.clouds = append(l.clouds, cloud{
			x:      float64(i)*step + rng.Float64()*step*0.5,
			y:      y,
			sprite: sprites[rng.Intn(len(sprites))],
			speed:  1,
		})
	}
	return l
}

// step scrolls the layer; wind is the signed drift (calm = right).
func (l *cloudLayer) step(wind float64) {
	v := l.speed * (1 + 3*math.Abs(wind))
	if wind < 0 {
		v = -v
	}
	for i := range l.clouds {
		l.clouds[i].x = math.Mod(l.clouds[i].x+v*l.clouds[i].speed+l.ring, l.ring)
	}
}

func (l *cloudLayer) draw(c *Canvas) {
	for _, cl := range l.clouds {
		drawSprite(c, int(math.Floor(cl.x))-maxSpriteW, cl.y, cloudSprites[cl.sprite], l.color)
	}
}

// drawSprite paints a sprite; interior spaces are opaque, edges transparent.
func drawSprite(c *Canvas, x, y int, sprite []string, col ColorID) {
	for dy, row := range sprite {
		runes := []rune(row)
		first, last := -1, -1
		for i, r := range runes {
			if r != ' ' {
				if first < 0 {
					first = i
				}
				last = i
			}
		}
		for i := first; i >= 0 && i <= last; i++ {
			c.Set(x+i, y+dy, runes[i], col)
		}
	}
}

// clouds is the partly-cloudy scene (sun or moon behind a few light
// clouds) and the overcast scene (three parallax layers).
type clouds struct {
	p        Params
	rng      *rand.Rand
	w, h     int
	backdrop Scene      // sun/moon for partly cloudy; nil when overcast
	ceiling  *cloudBand // overcast back layer
	layers   []*cloudLayer
}

func newClouds(p Params, rng *rand.Rand) *clouds {
	c := &clouds{p: p, rng: rng}
	if p.Condition == weather.PartlyCloudy {
		if p.IsDay {
			c.backdrop = newSun(Params{ASCIIOnly: p.ASCIIOnly, Intensity: weather.Moderate}, rng)
		} else {
			c.backdrop = newNight(Params{ASCIIOnly: p.ASCIIOnly}, rng)
		}
	}
	return c
}

func (c *clouds) Name() string {
	if c.backdrop != nil {
		return "cloudy"
	}
	return "overcast"
}

func (c *clouds) Resize(w, h int) {
	c.w, c.h = max(w, 1), max(h, 1)
	if c.backdrop != nil {
		c.backdrop.Resize(w, h)
		// 2-3 light clouds, each drifting at its own pace.
		n := 2 + c.rng.Intn(2)
		spacing := float64(c.w+maxSpriteW) / float64(n)
		l := newCloudLayer(c.rng, c.w, 0, max(0, c.h*55/100-3), spacing, 0.1, ColCloudLight, []int{0, 1, 2, 3})
		for i := range l.clouds {
			l.clouds[i].speed = between(c.rng, 0.5, 1.5) // 0.05-0.15 cells/frame
		}
		c.layers = []*cloudLayer{l}
		return
	}
	// Overcast: a dark, slow ceiling at the back, then mid and front sprite
	// layers moving progressively faster (parallax), filling ~60%.
	top := c.h * 60 / 100
	c.ceiling = newCloudBand(c.rng, max(1, top*45/100), ColCloudDark, ColStarDim, 0.03, c.p.ASCIIOnly)
	c.layers = []*cloudLayer{
		newCloudLayer(c.rng, c.w, top*15/100, max(0, top*55/100-2), 26, 0.06, ColCloudDark, []int{1, 2}),
		newCloudLayer(c.rng, c.w, top*40/100, max(0, top-3), 34, 0.1, ColCloudLight, []int{0, 1, 2, 3}),
	}
}

func (c *clouds) Step() {
	if c.backdrop != nil {
		c.backdrop.Step()
	}
	if c.ceiling != nil {
		c.ceiling.step(c.p.Wind)
	}
	for _, l := range c.layers {
		l.step(c.p.Wind)
	}
}

func (c *clouds) Draw(cv *Canvas) {
	if c.backdrop != nil {
		c.backdrop.Draw(cv)
	}
	if c.ceiling != nil {
		c.ceiling.draw(cv)
	}
	for _, l := range c.layers {
		l.draw(cv)
	}
}

// cloudBand is a dense, noise-shaded ceiling of cloud across the top rows,
// used by the rain, snow and storm scenes. Density falls off toward the
// bottom edge so its underside looks puffy; the noise scrolls with wind.
type cloudBand struct {
	rows        int
	n           *noise
	offset      float64
	t           float64
	speed       float64
	dense, thin ColorID
	ramp        []rune
	flash       bool // storm: paint the band bright
}

var (
	bandRampUnicode = []rune{' ', '░', '▒', '▓', '█'}
	bandRampASCII   = []rune{' ', '.', ':', '%', '#'}
)

func newCloudBand(rng *rand.Rand, rows int, dense, thin ColorID, speed float64, ascii bool) *cloudBand {
	b := &cloudBand{rows: rows, n: newNoise(rng), speed: speed, dense: dense, thin: thin, ramp: bandRampUnicode}
	if ascii {
		b.ramp = bandRampASCII
	}
	b.offset = rng.Float64() * 100
	return b
}

func (b *cloudBand) step(wind float64) {
	v := b.speed * (1 + 3*math.Abs(wind))
	if wind < 0 {
		v = -v
	}
	b.offset -= v
	b.t += 0.01 // slow churn so the shape evolves, not just slides
}

// density returns cloud density 0..1 at a band cell.
func (b *cloudBand) density(x, y int) float64 {
	fy := (float64(y) + 0.5) / float64(max(b.rows, 1))
	nv := b.n.fbm((float64(x)+b.offset)*0.13, float64(y)*0.35+b.t)
	return clampF(1.25-1.35*fy+(nv-0.5)*1.6, 0, 1)
}

func (b *cloudBand) draw(c *Canvas) {
	last := len(b.ramp) - 1
	for y := 0; y < b.rows; y++ {
		for x := 0; x < c.W; x++ {
			d := b.density(x, y)
			i := int(d*float64(last) + 0.5)
			if i <= 0 {
				continue
			}
			col := b.thin
			if i >= last-1 {
				col = b.dense
			}
			if b.flash {
				c.SetBold(x, y, b.ramp[i], ColBolt)
				continue
			}
			c.Set(x, y, b.ramp[i], col)
		}
	}
}
