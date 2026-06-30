// Package geocode converts concert location strings into geographic
// coordinates using the OpenStreetMap Nominatim service, with caching.
package geocode

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// nominatimURL is the geocoding endpoint. It is a var (not const) so tests can
// point it at a fake server.
var nominatimURL = "https://nominatim.openstreetmap.org/search"

// userAgent identifies this app to Nominatim, as required by its usage policy.
const userAgent = "groupie-tracker-geolocalization/1.0"

// nominatimResult mirrors the fields we need from a Nominatim search response.
type nominatimResult struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

// Geocode resolves a normalized query (e.g. "Mainz, Germany") to coordinates
// using the Nominatim service.
func Geocode(query string) (lat, lng float64, err error) {
	endpoint := nominatimURL + "?" + url.Values{
		"q":      {query},
		"format": {"json"},
		"limit":  {"1"},
	}.Encode()

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, 0, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("nominatim: unexpected status %d", resp.StatusCode)
	}

	var results []nominatimResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return 0, 0, err
	}
	if len(results) == 0 {
		return 0, 0, fmt.Errorf("no results for %q", query)
	}

	lat, err = strconv.ParseFloat(results[0].Lat, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parse lat: %w", err)
	}
	lng, err = strconv.ParseFloat(results[0].Lon, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parse lng: %w", err)
	}
	return lat, lng, nil
}

// Normalize converts a raw API location string (e.g. "north_carolina-usa")
// into a human-readable query (e.g. "North Carolina, USA"). The result is both
// shown to users and sent to the geocoding service.
func Normalize(raw string) string {
	// Underscores separate words within the city or country.
	s := strings.ReplaceAll(raw, "_", " ")

	// City and country are separated by the last "-".
	i := strings.LastIndex(s, "-")
	if i == -1 {
		return titleCase(s)
	}

	city := titleCase(s[:i])
	country := s[i+1:]
	return city + ", " + formatCountry(country)
}

// titleCase upper-cases the first letter of each word, lower-casing the rest.
func titleCase(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		words[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
	}
	return strings.Join(words, " ")
}

// formatCountry upper-cases short country codes (e.g. "usa" -> "USA") and
// title-cases longer names (e.g. "germany" -> "Germany").
func formatCountry(c string) string {
	c = strings.TrimSpace(c)
	if len(c) <= 3 {
		return strings.ToUpper(c)
	}
	return titleCase(c)
}
