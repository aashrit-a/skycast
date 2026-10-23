package scene

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"testing"

	"github.com/aashrit-a/skycast/internal/weather"
)

// TestPreview prints a frame when SKYCAST_PREVIEW is set (dev aid).
func TestPreview(t *testing.T) {
	cond := os.Getenv("SKYCAST_PREVIEW")
	if cond == "" {
		t.Skip()
	}
	c, _ := strconv.Atoi(cond)
	in, _ := strconv.Atoi(os.Getenv("SKYCAST_IN"))
	wind, _ := strconv.ParseFloat(os.Getenv("SKYCAST_WIND"), 64)
	w, _ := strconv.Atoi(os.Getenv("SKYCAST_W"))
	h, _ := strconv.Atoi(os.Getenv("SKYCAST_H"))
	steps, _ := strconv.Atoi(os.Getenv("SKYCAST_STEPS"))
	s := New(Params{Condition: weather.Condition(c), Intensity: weather.Intensity(in), Wind: wind, IsDay: os.Getenv("SKYCAST_NIGHT") == "", ASCIIOnly: os.Getenv("SKYCAST_ASCII") != ""}, rand.New(rand.NewSource(1)))
	s.Resize(w, h)
	cv := NewCanvas(w, h)
	for i := 0; i < steps; i++ {
		s.Step()
	}
	for f := 0; f < 2; f++ {
		cv.Clear()
		s.Draw(cv)
		fmt.Println("+" + string(make([]byte, 0)) + "frame")
		fmt.Println(cv.Render(Palette{}))
		for i := 0; i < 3; i++ {
			s.Step()
		}
	}
}
