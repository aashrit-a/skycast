package location

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// DefaultGeocodeURL is the Open-Meteo geocoding endpoint.
const DefaultGeocodeURL = "https://geocoding-api.open-meteo.com/v1/search"

// ErrLocationNotFound is returned (wrapped) when geocoding finds nothing.
var ErrLocationNotFound = errors.New("location not found")

type notFoundError struct{ name string }

func (e *notFoundError) Error() string {
	return fmt.Sprintf("could not find a place called %q", e.name)
}
func (e *notFoundError) Unwrap() error { return ErrLocationNotFound }

type geoResult struct {
	Name        string  `json:"name"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Admin1      string  `json:"admin1"`
	Timezone    string  `json:"timezone"`
}

type geoResponse struct {
	Results []geoResult `json:"results"`
}

// Geocode looks up a city by name. A qualifier after a comma ("san jose, ca",
// "paris, us") is matched against the region, country and country code.
func (r *Resolver) Geocode(ctx context.Context, name string) (Location, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Location{}, &notFoundError{name: name}
	}
	// NOTE: Open-Meteo matches the whole query string as a place name, so
	// "san jose, ca" finds nothing. We search for the part before the first
	// comma and use the rest to choose among the top results.
	place, qualifier, _ := strings.Cut(name, ",")
	place = strings.TrimSpace(place)
	qualifier = strings.ToLower(strings.TrimSpace(qualifier))
	count := "1"
	if qualifier != "" {
		count = "10"
	}

	v := url.Values{}
	v.Set("name", place)
	v.Set("count", count)
	v.Set("language", "en")
	v.Set("format", "json")
	var resp geoResponse
	if err := r.getJSON(ctx, r.GeocodeURL+"?"+v.Encode(), &resp); err != nil {
		if ctx.Err() != nil {
			return Location{}, ctx.Err()
		}
		return Location{}, fmt.Errorf("can't reach the geocoding service — check your connection (%v)", err)
	}
	if len(resp.Results) == 0 {
		return Location{}, &notFoundError{name: name}
	}
	best := resp.Results[0]
	if qualifier != "" {
		found := false
		for _, res := range resp.Results {
			if matchesQualifier(res, qualifier) {
				best, found = res, true
				break
			}
		}
		if !found {
			return Location{}, &notFoundError{name: name}
		}
	}
	return Location{
		Name:        best.Name,
		Region:      best.Admin1,
		Country:     best.Country,
		CountryCode: strings.ToUpper(best.CountryCode),
		Timezone:    best.Timezone,
		Lat:         best.Latitude,
		Lon:         best.Longitude,
	}, nil
}

func matchesQualifier(r geoResult, q string) bool {
	if strings.EqualFold(r.CountryCode, q) {
		return true
	}
	for _, s := range []string{r.Admin1, r.Country} {
		if s != "" && strings.HasPrefix(strings.ToLower(s), q) {
			return true
		}
	}
	return false
}
