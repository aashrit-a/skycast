package location

import (
	"context"
	"errors"
)

// DefaultIPURL is the ip-api.com endpoint (HTTP only on the free tier).
const DefaultIPURL = "http://ip-api.com/json/?fields=status,message,country,countryCode,regionName,city,lat,lon,timezone"

// ErrIPLookup is returned when the location can't be detected from the IP.
var ErrIPLookup = errors.New(`couldn't detect your location; try: skycast "new york"`)

type ipResponse struct {
	Status      string  `json:"status"`
	Message     string  `json:"message"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string  `json:"timezone"`
}

// FromIP detects a coarse location from the caller's public IP address.
func (r *Resolver) FromIP(ctx context.Context) (Location, error) {
	var resp ipResponse
	if err := r.getJSON(ctx, r.IPURL, &resp); err != nil {
		if ctx.Err() != nil {
			return Location{}, ctx.Err()
		}
		return Location{}, ErrIPLookup
	}
	if resp.Status != "success" {
		return Location{}, ErrIPLookup
	}
	return Location{
		Name:        resp.City,
		Region:      resp.RegionName,
		Country:     resp.Country,
		CountryCode: resp.CountryCode,
		Timezone:    resp.Timezone,
		Lat:         resp.Lat,
		Lon:         resp.Lon,
	}, nil
}
