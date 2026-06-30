package handlers

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"groupie-tracker-geolocalization/geocode"
)

// geocoder resolves a normalized location query to coordinates. It is a var so
// tests can replace it without hitting the network. By default it uses the
// on-disk geocode cache (which falls back to Nominatim on a miss).
var geocoder = geocode.New("data/geocache.json").Coordinates

// concertCoordinate is one mapped concert location in the JSON response.
type concertCoordinate struct {
	Address string   `json:"address"`
	Lat     float64  `json:"lat"`
	Lng     float64  `json:"lng"`
	Dates   []string `json:"dates"`
}

// coordinatesHandler returns an artist's concert locations as geocoded points,
// sorted by the earliest concert date of each location.
func coordinatesHandler(w http.ResponseWriter, r *http.Request) {
	if !checkMethod(w, r) {
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		render404(w)
		return
	}

	var datesLocations map[string][]string
	for _, rel := range cache.Relations.Index {
		if rel.ID == id {
			datesLocations = rel.DatesLocations
			break
		}
	}
	if datesLocations == nil {
		render404(w)
		return
	}

	coords := make([]concertCoordinate, 0, len(datesLocations))
	for rawLocation, dates := range datesLocations {
		address := geocode.Normalize(rawLocation)
		point, err := geocoder(address)
		if err != nil {
			continue // skip locations we cannot geocode (full handling in M6)
		}
		coords = append(coords, concertCoordinate{
			Address: address,
			Lat:     point.Lat,
			Lng:     point.Lng,
			Dates:   dates,
		})
	}

	sort.Slice(coords, func(i, j int) bool {
		ei, ej := earliestDate(coords[i].Dates), earliestDate(coords[j].Dates)
		if ei.Equal(ej) {
			return coords[i].Address < coords[j].Address
		}
		return ei.Before(ej)
	})

	writeJSON(w, coords)
}

// earliestDate parses concert dates (format "02-01-2006", possibly prefixed
// with "*") and returns the earliest. Unparseable dates are ignored.
func earliestDate(dates []string) time.Time {
	var earliest time.Time
	for _, d := range dates {
		d = strings.TrimPrefix(strings.TrimSpace(d), "*")
		parsed, err := time.Parse("02-01-2006", d)
		if err != nil {
			continue
		}
		if earliest.IsZero() || parsed.Before(earliest) {
			earliest = parsed
		}
	}
	return earliest
}
