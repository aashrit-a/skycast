package ui

import "testing"

func TestComputeLayout(t *testing.T) {
	tests := []struct {
		w, h           int
		panel, ambient bool
		want           Layout
	}{
		{200, 50, true, false, Layout{SideBySide, 168, 30, 30, 17}},
		{100, 30, true, false, Layout{SideBySide, 68, 29, 30, 17}},
		{72, 24, true, false, Layout{SideBySide, 40, 23, 30, 17}},
		{71, 24, true, false, Layout{Stacked, 71, 14, 71, 9}},
		{50, 30, true, false, Layout{Stacked, 50, 20, 50, 9}},
		{50, 15, true, false, Layout{Minimal, 0, 0, 0, 0}},
		{30, 10, true, false, Layout{Minimal, 0, 0, 0, 0}},
		{80, 12, true, false, Layout{SideBySide, 48, 11, 30, 11}},
		{80, 9, true, false, Layout{Minimal, 0, 0, 0, 0}},
		{200, 50, true, true, Layout{Ambient, 200, 50, 0, 0}},
		{30, 10, true, true, Layout{Ambient, 30, 10, 0, 0}},
		{9, 10, true, true, Layout{Minimal, 0, 0, 0, 0}},
		{100, 30, false, false, Layout{SceneOnly, 100, 29, 0, 0}},
		{30, 10, false, false, Layout{SceneOnly, 30, 9, 0, 0}},
		{30, 4, false, false, Layout{Minimal, 0, 0, 0, 0}},
		{0, 0, true, false, Layout{Minimal, 0, 0, 0, 0}},
	}
	for _, tt := range tests {
		got := ComputeLayout(tt.w, tt.h, tt.panel, tt.ambient)
		if got != tt.want {
			t.Errorf("ComputeLayout(%d,%d,panel=%v,ambient=%v) = %+v, want %+v", tt.w, tt.h, tt.panel, tt.ambient, got, tt.want)
		}
	}
}

func TestLayoutNeverDegenerate(t *testing.T) {
	for w := 0; w <= 220; w += 3 {
		for h := 0; h <= 60; h++ {
			for _, p := range []bool{true, false} {
				for _, a := range []bool{true, false} {
					l := ComputeLayout(w, h, p, a)
					if l.Mode == Minimal {
						continue
					}
					if l.SceneW < minSceneW || l.SceneH < minSceneH {
						t.Fatalf("%dx%d p=%v a=%v: scene too small %+v", w, h, p, a, l)
					}
					used := l.SceneH
					if l.Mode == Stacked {
						used += l.PanelH
					}
					if l.HasFooter() {
						used++
					}
					if used > h || l.SceneW > w {
						t.Fatalf("%dx%d: layout overflows %+v", w, h, l)
					}
				}
			}
		}
	}
}
