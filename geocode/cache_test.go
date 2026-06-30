package geocode

import (
	"path/filepath"
	"testing"
)

func TestCacheServesSecondLookupFromMemory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "geocache.json")
	c := New(path)

	// Inject a fake geocoder that counts how many times it is called.
	calls := 0
	c.geocode = func(query string) (float64, float64, error) {
		calls++
		return 49.99, 8.25, nil
	}

	first, err := c.Coordinates("Mainz, Germany")
	if err != nil {
		t.Fatalf("first lookup: %v", err)
	}
	second, err := c.Coordinates("Mainz, Germany")
	if err != nil {
		t.Fatalf("second lookup: %v", err)
	}

	if first != second {
		t.Errorf("coords differ: %v vs %v", first, second)
	}
	if calls != 1 {
		t.Errorf("geocoder called %d times, want 1 (second lookup should hit the cache)", calls)
	}
}

func TestCachePersistsToDisk(t *testing.T) {
	path := filepath.Join(t.TempDir(), "geocache.json")

	// First cache instance: resolve and store one entry.
	c1 := New(path)
	c1.geocode = func(query string) (float64, float64, error) {
		return 49.99, 8.25, nil
	}
	want, err := c1.Coordinates("Mainz, Germany")
	if err != nil {
		t.Fatalf("populate: %v", err)
	}

	// A fresh cache from the same file must load the entry from disk,
	// without calling the geocoder.
	c2 := New(path)
	c2.geocode = func(query string) (float64, float64, error) {
		t.Fatal("geocoder must not be called: entry should come from disk")
		return 0, 0, nil
	}
	got, err := c2.Coordinates("Mainz, Germany")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got != want {
		t.Errorf("loaded %v, want %v", got, want)
	}
}
