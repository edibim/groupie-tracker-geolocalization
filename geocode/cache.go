package geocode

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Coords is a geographic coordinate pair.
type Coords struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// Cache stores geocoding results in memory and on disk, keyed by the
// normalized query (e.g. "Mainz, Germany"). It is safe for concurrent use.
type Cache struct {
	mu      sync.Mutex
	entries map[string]Coords
	path    string
	geocode func(query string) (float64, float64, error)
}

// New creates a Cache backed by the file at path, loading any existing entries.
func New(path string) *Cache {
	c := &Cache{
		entries: make(map[string]Coords),
		path:    path,
		geocode: Geocode,
	}
	c.load()
	return c
}

// Coordinates returns the coordinates for query. It serves cached results when
// available and otherwise calls the geocoder, persisting the new result.
func (c *Cache) Coordinates(query string) (Coords, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if co, ok := c.entries[query]; ok {
		return co, nil
	}

	lat, lng, err := c.geocode(query)
	if err != nil {
		return Coords{}, err
	}

	co := Coords{Lat: lat, Lng: lng}
	c.entries[query] = co
	c.save()
	return co, nil
}

// load reads the cache file into memory. A missing file is not an error.
func (c *Cache) load() {
	data, err := os.ReadFile(c.path)
	if err != nil {
		return
	}
	json.Unmarshal(data, &c.entries)
}

// save writes the in-memory cache to disk, creating the directory if needed.
func (c *Cache) save() {
	os.MkdirAll(filepath.Dir(c.path), 0o755)
	data, err := json.MarshalIndent(c.entries, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(c.path, data, 0o644)
}
