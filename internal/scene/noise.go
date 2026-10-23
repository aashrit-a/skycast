package scene

import (
	"math"
	"math/rand"
)

// noise is a tiny seeded value-noise generator (no dependencies).
type noise struct {
	perm [256]uint8
	vals [256]float64
}

func newNoise(rng *rand.Rand) *noise {
	n := &noise{}
	for i := range n.perm {
		n.perm[i] = uint8(i)
		n.vals[i] = rng.Float64()
	}
	rng.Shuffle(len(n.perm), func(i, j int) { n.perm[i], n.perm[j] = n.perm[j], n.perm[i] })
	return n
}

func (n *noise) lattice(x, y int) float64 {
	return n.vals[n.perm[(int(n.perm[x&255])+y)&255]]
}

func smooth(t float64) float64 { return t * t * (3 - 2*t) }

func lerp(a, b, t float64) float64 { return a + (b-a)*t }

// at2 returns smooth 2D value noise in [0, 1).
func (n *noise) at2(x, y float64) float64 {
	x0, y0 := math.Floor(x), math.Floor(y)
	tx, ty := smooth(x-x0), smooth(y-y0)
	ix, iy := int(x0), int(y0)
	a := lerp(n.lattice(ix, iy), n.lattice(ix+1, iy), tx)
	b := lerp(n.lattice(ix, iy+1), n.lattice(ix+1, iy+1), tx)
	return lerp(a, b, ty)
}

// at returns smooth 1D value noise in [0, 1).
func (n *noise) at(x float64) float64 { return n.at2(x, 0) }

// fbm sums two octaves for a more natural, billowy look.
func (n *noise) fbm(x, y float64) float64 {
	return (n.at2(x, y)*2 + n.at2(x*2.3+17, y*2.3+5)) / 3
}
