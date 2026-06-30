# Milestones — Groupie Tracker Geolocalization

> Companion to [PRD.md](PRD.md). This project builds on the finished
> **groupie-tracker** base, so milestones cover only the geolocalization layer.
> "Done" means: code compiles, `go test ./...` passes, and the exit criteria are
> met.

## M1 — Bootstrap from the base

**Goal:** start from a working copy of the base project.

- Copy the base `groupie-tracker` code into this repo.
- Rename the module to `groupie-tracker-geolocalization` and fix import paths.
- Verify the inherited site runs (`go run .`) and the base tests pass.

**Exit criteria:** artist listing, search/filters, artist page, JSON sub-endpoints
and 404/405/500 pages all work, unchanged from the base.
**Tests:** existing base tests stay green.

## M2 — Geocoding package

**Goal:** turn a text address into coordinates, throttled and cached.

- `geocode/geocode.go`: `Normalize(raw)` (`_`→space, split on last `-` into
  `City, Country`), then `Geocode(query)` via Nominatim (`net/http`) with a valid
  `User-Agent` and ~1 request/sec throttle for uncached lookups.
- `geocode/cache.go`: in-memory map keyed by the normalized query + load/save
  `data/geocache.json`.

**Exit criteria:** `Geocode("Mainz, Germany")` returns plausible coords; a second
call is served from cache with no network request.
**Tests:** `geocode_test.go` (parse saved Nominatim fixtures from `testdata/`),
`cache_test.go` (hit/miss/persist).

## M3 — Coordinates endpoint

**Goal:** expose an artist's geocoded concerts as JSON.

- Add `coordinatesHandler` and register `GET /artist/{id}/coordinates` in
  `handlers/handlers.go`, following the base's `/artist/{id}/…` pattern and
  `checkMethod`.
- Build the response from the artist's `relation.DatesLocations`; for each
  location, normalize it to the display form (`Mainz, Germany`) used as the
  `address` field, geocode it, then sort by the **earliest date per location**
  (strip `*`, parse `02-01-2006`; ties by address).

**Exit criteria:** `/artist/{id}/coordinates` returns `[{address, lat, lng,
dates}]` sorted chronologically; bad id → 404; non-GET → 405.
**Tests:** `handlers_test.go` — fields present, ordering correct, error codes.

## M4 — Map UI

**Goal:** show the markers on the artist page.

- Vendor Leaflet into `static/vendor/leaflet/` (js, css, marker images).
- Edit `templates/artist.html` to add a `#concert-map` container and load Leaflet
  + `static/js/map.js`.
- `static/js/map.js`: init the map, fetch `/artist/{id}/coordinates`, drop one
  marker per location, fit bounds.

**Exit criteria:** opening "Queen" shows markers at the audit cities; the map
works with no internet (vendored Leaflet).
**Tests:** covered by M3 endpoint tests (JS is exercised manually).

## M5 — Tour path (bonus)

**Goal:** connect markers chronologically.

- In `map.js`, draw a Leaflet polyline through the markers in the order returned
  by the endpoint (already sorted by earliest date in M3).

**Exit criteria:** markers are joined by a line that follows concert date order.

## M6 — Error handling & polish

**Goal:** geolocalization degrades gracefully.

- Skip ungeocodable addresses with a logged warning; the rest still render.
- Handle Nominatim being unreachable: omit that marker, keep the page working;
  `map.js` shows a non-blocking notice.
- Manual audit pass for all listed artists.

**Exit criteria:** an artist with one bad address still maps the others; a
simulated Nominatim outage does not crash or block the page.
**Tests:** `handlers_test.go` — endpoint still returns resolved locations when
some addresses fail.

## M7 — Tests, docs & cleanup

**Goal:** ship-ready quality.

- Fill any test gaps for the new logic.
- `README.md`: run instructions, screenshots, Nominatim + Leaflet credits, note
  on the geocache.
- `gofmt ./...`, `go vet ./...`, remove dead code, final review.

**Exit criteria:** `go test ./...` green; `go vet ./...` clean; README complete.

## Dependency order

```
M1 → M2 → M3 → M4 → M5
                └──→ M6 → M7
```

M6 starts once the map (M4) exists; M5 (bonus) can run in parallel with early
M6 work.
