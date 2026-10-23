// Package location turns a city name, coordinates, or the caller's IP
// address into a place with coordinates.
package location

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// Location is a resolved place.
type Location struct {
	Name        string
	Region      string // state / province
	Country     string
	CountryCode string // ISO 3166-1 alpha-2, may be empty
	Timezone    string
	Lat, Lon    float64
}

// Display returns a short human label such as "Tokyo, Japan".
func (l Location) Display() string {
	switch {
	case l.Name == "":
		return l.Country
	case l.Country == "" || l.Country == l.Name:
		return l.Name
	}
	return l.Name + ", " + l.Country
}

// Resolver looks up locations. Base URLs are fields so tests can point
// them at httptest servers.
type Resolver struct {
	GeocodeURL string
	IPURL      string
	HTTP       *http.Client
	UserAgent  string
}

// NewResolver returns a Resolver using the public Open-Meteo and ip-api endpoints.
func NewResolver(userAgent string) *Resolver {
	return &Resolver{
		GeocodeURL: DefaultGeocodeURL,
		IPURL:      DefaultIPURL,
		HTTP:       &http.Client{Timeout: 10 * time.Second},
		UserAgent:  userAgent,
	}
}

// Resolve picks a strategy: explicit coordinates, then city geocoding,
// then IP lookup.
func (r *Resolver) Resolve(ctx context.Context, city string, lat, lon float64, hasCoords bool) (Location, error) {
	if hasCoords {
		// NOTE: we don't reverse-geocode coordinates (no free keyless
		// endpoint on Open-Meteo); the label is just the coordinates and
		// the country code is unknown, so units default to metric.
		return Location{
			Name: strconv.FormatFloat(lat, 'f', 2, 64) + ", " + strconv.FormatFloat(lon, 'f', 2, 64),
			Lat:  lat, Lon: lon,
		}, nil
	}
	if city != "" {
		return r.Geocode(ctx, city)
	}
	return r.FromIP(ctx)
}

func (r *Resolver) getJSON(ctx context.Context, u string, v any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	if r.UserAgent != "" {
		req.Header.Set("User-Agent", r.UserAgent)
	}
	hc := r.HTTP
	if hc == nil {
		hc = &http.Client{Timeout: 10 * time.Second}
	}
	resp, err := hc.Do(req)
	if err != nil {
		return &netError{err: err}
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

// netError marks a transport failure (as opposed to a bad response).
type netError struct{ err error }

func (e *netError) Error() string { return e.err.Error() }
func (e *netError) Unwrap() error { return e.err }
