package weather

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestCacheFreshHit(t *testing.T) {
	hits := 0
	srv := serve(t, 200, fixture(t, "forecast_rain.json"), &hits)
	cache := &Cache{Dir: t.TempDir()}
	c := testClient(srv.URL, cache)
	ctx := context.Background()
	if _, err := c.Fetch(ctx, 35.69, 139.69, UnitsMetric); err != nil {
		t.Fatal(err)
	}
	r, err := c.Fetch(ctx, 35.69, 139.69, UnitsMetric)
	if err != nil {
		t.Fatal(err)
	}
	if hits != 1 {
		t.Errorf("hits = %d, want 1 (second fetch should be cached)", hits)
	}
	if r.Stale || r.Current.Code != 61 {
		t.Errorf("bad cached report %+v", r)
	}
	// Different units are a different cache key.
	if _, err := c.Fetch(ctx, 35.69, 139.69, UnitsImperial); err != nil {
		t.Fatal(err)
	}
	if hits != 2 {
		t.Errorf("hits = %d, want 2", hits)
	}
	// NoCacheRead bypasses the cache.
	c.NoCacheRead = true
	if _, err := c.Fetch(ctx, 35.69, 139.69, UnitsMetric); err != nil {
		t.Fatal(err)
	}
	if hits != 3 {
		t.Errorf("hits = %d, want 3", hits)
	}
}

func TestCacheExpiredMiss(t *testing.T) {
	hits := 0
	srv := serve(t, 200, fixture(t, "forecast_rain.json"), &hits)
	cache := &Cache{Dir: t.TempDir()}
	cache.Save(1, 2, UnitsMetric, fixture(t, "forecast_rain.json"), time.Now().Add(-11*time.Minute))
	r, err := testClient(srv.URL, cache).Fetch(context.Background(), 1, 2, UnitsMetric)
	if err != nil {
		t.Fatal(err)
	}
	if hits != 1 || r.Stale {
		t.Errorf("hits=%d stale=%v; want network fetch", hits, r.Stale)
	}
}

func TestCacheStaleFallback(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	srv.Close() // unreachable
	cache := &Cache{Dir: t.TempDir()}
	saved := time.Now().Add(-2 * time.Hour)
	cache.Save(1, 2, UnitsMetric, fixture(t, "forecast_rain.json"), saved)
	r, err := testClient(srv.URL, cache).Fetch(context.Background(), 1, 2, UnitsMetric)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Stale || r.FetchedAt.Sub(saved).Abs() > time.Second {
		t.Errorf("stale=%v fetchedAt=%v", r.Stale, r.FetchedAt)
	}

	// Older than 24h is not used.
	cache.Save(1, 2, UnitsMetric, fixture(t, "forecast_rain.json"), time.Now().Add(-25*time.Hour))
	if _, err := testClient(srv.URL, cache).Fetch(context.Background(), 1, 2, UnitsMetric); err == nil {
		t.Error("want error for too-old cache")
	}
}

func TestCacheCorruptIgnored(t *testing.T) {
	hits := 0
	srv := serve(t, 200, fixture(t, "forecast_rain.json"), &hits)
	cache := &Cache{Dir: t.TempDir()}
	if err := os.WriteFile(cache.path(1, 2, UnitsMetric), []byte("{garbage"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, ok := cache.Load(1, 2, UnitsMetric); ok {
		t.Error("corrupt entry should not load")
	}
	if _, err := testClient(srv.URL, cache).Fetch(context.Background(), 1, 2, UnitsMetric); err != nil {
		t.Fatal(err)
	}
	if hits != 1 {
		t.Errorf("hits = %d", hits)
	}
}

func TestNilCacheSafe(t *testing.T) {
	var c *Cache
	c.Save(1, 2, UnitsMetric, []byte("{}"), time.Now())
	if _, _, ok := c.Load(1, 2, UnitsMetric); ok {
		t.Error("nil cache returned data")
	}
}

func TestCacheUnwritableDir(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "file")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	c := &Cache{Dir: f.Name()} // a file, not a dir: every write fails
	c.Save(1, 2, UnitsMetric, []byte("{}"), time.Now())
}
