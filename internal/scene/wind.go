package scene

import (
	"math"
	"math/rand"
)

// windOverlayThreshold is the |wind| above which streaks are drawn.
const windOverlayThreshold = 0.45

type streak struct {
	x, y, v float64
	n       int
	glyph   rune
}

// windOverlay decorates a scene with horizontal wind streaks.
type windOverlay struct {
	inner   Scene
	wind    float64
	rng     *rand.Rand
	w, h    int
	streaks []streak
}

// WithWind wraps s with a wind-streak overlay drifting in the wind direction.
func WithWind(s Scene, wind float64, rng *rand.Rand) Scene {
	return &windOverlay{inner: s, wind: clampF(wind, -1, 1), rng: rng}
}

func (o *windOverlay) base() Scene  { return o.inner }
func (o *windOverlay) Name() string { return o.inner.Name() }

func (o *windOverlay) Resize(w, h int) {
	o.inner.Resize(w, h)
	o.w, o.h = max(w, 1), max(h, 1)
	n := max(1, int(math.Round(float64(o.h)*math.Abs(o.wind)*0.5)))
	o.streaks = o.streaks[:0]
	for i := 0; i < n; i++ {
		s := o.newStreak()
		s.x = between(o.rng, -float64(o.w), float64(o.w)) // scatter on start
		o.streaks = append(o.streaks, s)
	}
}

func (o *windOverlay) newStreak() streak {
	s := streak{
		y:     float64(o.rng.Intn(o.h)),
		v:     between(o.rng, 1.5, 3) * math.Copysign(1, o.wind),
		n:     3 + o.rng.Intn(5),
		glyph: '~',
	}
	if o.rng.Intn(2) == 0 {
		s.glyph = '-'
	}
	// Spawn off-screen on the upwind side, staggered so they don't arrive together.
	gap := between(o.rng, 0, float64(o.w))
	if o.wind > 0 {
		s.x = -float64(s.n) - gap
	} else {
		s.x = float64(o.w) + gap
	}
	return s
}

func (o *windOverlay) Step() {
	o.inner.Step()
	for i := range o.streaks {
		s := &o.streaks[i]
		s.x += s.v
		if (s.v > 0 && s.x > float64(o.w)) || (s.v < 0 && s.x+float64(s.n) < 0) {
			*s = o.newStreak()
		}
	}
}

func (o *windOverlay) Draw(c *Canvas) {
	o.inner.Draw(c)
	for _, s := range o.streaks {
		x0, y := int(math.Floor(s.x)), int(s.y)
		for i := 0; i < s.n; i++ {
			// Streaks pass behind anything already drawn.
			if c.IsEmpty(x0+i, y) {
				c.Set(x0+i, y, s.glyph, ColWind)
			}
		}
	}
}
