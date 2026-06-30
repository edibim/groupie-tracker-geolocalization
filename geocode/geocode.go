// Package geocode converts concert location strings into geographic
// coordinates using the OpenStreetMap Nominatim service, with caching.
package geocode

import "strings"

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
