package ui

import (
	"bytes"
	"context"
	"flag"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aashrit-a/skycast/internal/config"
	"github.com/aashrit-a/skycast/internal/scene"
)

var updateGolden = flag.Bool("update", false, "rewrite golden files")

func TestOnceGolden(t *testing.T) {
	cfg := config.Config{Width: 40, Height: 12, NoColor: true, ASCII: true}
	got := RenderOnce(sampleReport(0), cfg, scene.NewPalette(true), rand.New(rand.NewSource(1)), fixedNow)
	golden := filepath.Join("..", "..", "testdata", "once_rain.golden")
	if *updateGolden {
		if err := os.WriteFile(golden, []byte(got+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("%v (run with -update to create it)", err)
	}
	if got+"\n" != string(want) {
		t.Errorf("output differs from %s (run with -update if intended)\n--- got ---\n%s\n--- want ---\n%s", golden, got, want)
	}
	for i := 0; i < len(got); i++ {
		if got[i] > 127 {
			t.Fatalf("non-ASCII byte %#x at %d", got[i], i)
		}
	}
}

func TestRunOnceDemoNoNetwork(t *testing.T) {
	var out bytes.Buffer
	cfg := config.Config{Demo: true, Width: 40, Height: 12, NoColor: true}
	err := RunOnce(context.Background(), cfg, nil, scene.NewPalette(true), &out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Demo City") || strings.Contains(out.String(), "\x1b[") {
		t.Errorf("unexpected output:\n%s", out.String())
	}
}

func TestRunOnceError(t *testing.T) {
	fl := &fakeLoader{err: os.ErrDeadlineExceeded}
	cfg := config.Config{Width: 40, Height: 12}
	if err := RunOnce(context.Background(), cfg, fl.load, scene.NewPalette(true), &bytes.Buffer{}); err == nil {
		t.Error("want error")
	}
}
