package scene

import "strings"

// Cell is one character position on the canvas.
type Cell struct {
	R     rune
	Color ColorID // index into the active palette; 0 = default/none
	Bold  bool
}

// Canvas is a fixed-size grid of cells that scenes draw into.
type Canvas struct {
	W, H  int
	Cells []Cell // row-major, len W*H
}

// NewCanvas returns a blank canvas. Negative sizes are treated as zero.
func NewCanvas(w, h int) *Canvas {
	w, h = max(w, 0), max(h, 0)
	c := &Canvas{W: w, H: h, Cells: make([]Cell, w*h)}
	c.Clear()
	return c
}

// Resize changes the canvas size (reusing memory when possible) and clears it.
func (c *Canvas) Resize(w, h int) {
	w, h = max(w, 0), max(h, 0)
	c.W, c.H = w, h
	if cap(c.Cells) >= w*h {
		c.Cells = c.Cells[:w*h]
	} else {
		c.Cells = make([]Cell, w*h)
	}
	c.Clear()
}

// Clear blanks every cell.
func (c *Canvas) Clear() {
	for i := range c.Cells {
		c.Cells[i] = Cell{R: ' '}
	}
}

// In reports whether (x, y) is on the canvas.
func (c *Canvas) In(x, y int) bool {
	return x >= 0 && y >= 0 && x < c.W && y < c.H
}

// Set writes a rune; out-of-bounds coordinates are silently ignored.
func (c *Canvas) Set(x, y int, r rune, col ColorID) {
	if c.In(x, y) {
		c.Cells[y*c.W+x] = Cell{R: r, Color: col}
	}
}

// SetBold writes a bold rune; out-of-bounds coordinates are ignored.
func (c *Canvas) SetBold(x, y int, r rune, col ColorID) {
	if c.In(x, y) {
		c.Cells[y*c.W+x] = Cell{R: r, Color: col, Bold: true}
	}
}

// Get returns the cell at (x, y), or a blank cell if out of bounds.
func (c *Canvas) Get(x, y int) Cell {
	if c.In(x, y) {
		return c.Cells[y*c.W+x]
	}
	return Cell{R: ' '}
}

// IsEmpty reports whether (x, y) is on the canvas and blank.
func (c *Canvas) IsEmpty(x, y int) bool {
	return c.In(x, y) && c.Cells[y*c.W+x].R == ' '
}

// DrawString writes s starting at (x, y), one rune per cell.
func (c *Canvas) DrawString(x, y int, s string, col ColorID) {
	for _, r := range s {
		c.Set(x, y, r, col)
		x++
	}
}

// Render returns the canvas as H lines of exactly W columns joined by "\n".
// Consecutive cells with the same style are emitted as a single run.
func (c *Canvas) Render(p Palette) string {
	var b strings.Builder
	b.Grow(c.W*c.H*2 + c.H*24)
	for y := 0; y < c.H; y++ {
		if y > 0 {
			b.WriteByte('\n')
		}
		row := c.Cells[y*c.W : (y+1)*c.W]
		x := 0
		for x < len(row) {
			col, bold := styleOf(row[x])
			end := x + 1
			for end < len(row) {
				c2, b2 := styleOf(row[end])
				if c2 != col || b2 != bold {
					break
				}
				end++
			}
			pre, suf := p.codes(col, bold)
			b.WriteString(pre)
			for _, cell := range row[x:end] {
				r := cell.R
				if r == 0 {
					r = ' '
				}
				b.WriteRune(r)
			}
			b.WriteString(suf)
			x = end
		}
	}
	return b.String()
}

// styleOf normalizes blanks to the default style so runs of spaces merge
// regardless of the color they were written with.
func styleOf(c Cell) (ColorID, bool) {
	if c.R == ' ' || c.R == 0 {
		return ColNone, false
	}
	return c.Color, c.Bold
}
