package ui

import (
	"context"
	"sync"
	"time"

	"github.com/aashrit-a/skycast/internal/config"
	"github.com/aashrit-a/skycast/internal/location"
	"github.com/aashrit-a/skycast/internal/weather"
)

// fetchTimeout bounds one full load (location + forecast).
const fetchTimeout = 30 * time.Second

// LoadRequest describes one weather load.
type LoadRequest struct {
	Units       weather.Units
	AutoUnits   bool // pick units from the location's country
	BypassCache bool
}

// Loader resolves the location and fetches its report. Report.Units holds
// the units actually used.
type Loader func(ctx context.Context, req LoadRequest) (*weather.Report, error)

// NewLoader returns the real network loader for cfg. The location is
// resolved once and reused for refreshes and unit changes.
func NewLoader(cfg config.Config, userAgent string) Loader {
	resolver := location.NewResolver(userAgent)
	client := weather.NewClient(userAgent, weather.NewCache())
	var (
		mu       sync.Mutex
		resolved *location.Location
	)
	return func(ctx context.Context, req LoadRequest) (*weather.Report, error) {
		ctx, cancel := context.WithTimeout(ctx, fetchTimeout)
		defer cancel()
		mu.Lock()
		loc := resolved
		mu.Unlock()
		if loc == nil {
			l, err := resolver.Resolve(ctx, cfg.City, cfg.Lat, cfg.Lon, cfg.HasCoords)
			if err != nil {
				return nil, err
			}
			mu.Lock()
			resolved = &l
			mu.Unlock()
			loc = &l
		}
		units := req.Units
		if req.AutoUnits {
			units = weather.DefaultUnits(loc.CountryCode)
		}
		// Copy the client so per-request options can't race between loads.
		c := *client
		c.NoCacheRead = cfg.NoCache || req.BypassCache
		r, err := c.Fetch(ctx, loc.Lat, loc.Lon, units)
		if err != nil {
			return nil, err
		}
		r.Location = loc.Display()
		return r, nil
	}
}

// placeLabel describes where we're fetching for, for the loading screen.
func placeLabel(cfg config.Config) string {
	switch {
	case cfg.HasCoords:
		return formatCoords(cfg.Lat, cfg.Lon)
	case cfg.City != "":
		return cfg.City
	}
	return "your location"
}
