package weather

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Cache lifetimes.
const (
	CacheTTL      = 10 * time.Minute
	CacheMaxStale = 24 * time.Hour
)

// Cache stores raw forecast responses on disk. A nil *Cache is valid and
// caches nothing. All failures are swallowed: caching is best-effort.
type Cache struct {
	Dir string
}

// NewCache returns a cache under os.UserCacheDir()/skycast, or nil if the
// user cache directory can't be determined.
func NewCache() *Cache {
	dir, err := os.UserCacheDir()
	if err != nil || dir == "" {
		return nil
	}
	return &Cache{Dir: filepath.Join(dir, "skycast")}
}

type cacheEntry struct {
	SavedAt time.Time       `json:"saved_at"`
	Data    json.RawMessage `json:"data"`
}

func (c *Cache) path(lat, lon float64, u Units) string {
	return filepath.Join(c.Dir, fmt.Sprintf("%.2f_%.2f_%s.json", lat, lon, u))
}

// Load returns the cached response and when it was saved. ok is false if
// there is no usable entry (missing, corrupt, or unreadable).
func (c *Cache) Load(lat, lon float64, u Units) (data []byte, savedAt time.Time, ok bool) {
	if c == nil {
		return nil, time.Time{}, false
	}
	b, err := os.ReadFile(c.path(lat, lon, u))
	if err != nil {
		return nil, time.Time{}, false
	}
	var e cacheEntry
	if err := json.Unmarshal(b, &e); err != nil || e.SavedAt.IsZero() || len(e.Data) == 0 {
		return nil, time.Time{}, false
	}
	return e.Data, e.SavedAt, true
}

// Save writes a response to the cache, ignoring any error.
func (c *Cache) Save(lat, lon float64, u Units, data []byte, now time.Time) {
	if c == nil || !json.Valid(data) {
		return
	}
	b, err := json.Marshal(cacheEntry{SavedAt: now, Data: data})
	if err != nil {
		return
	}
	if err := os.MkdirAll(c.Dir, 0o755); err != nil {
		return
	}
	// Write to a temp file then rename so readers never see a partial file.
	tmp, err := os.CreateTemp(c.Dir, "tmp-*.json")
	if err != nil {
		return
	}
	_, werr := tmp.Write(b)
	cerr := tmp.Close()
	if werr != nil || cerr != nil {
		_ = os.Remove(tmp.Name())
		return
	}
	if err := os.Rename(tmp.Name(), c.path(lat, lon, u)); err != nil {
		_ = os.Remove(tmp.Name())
	}
}
