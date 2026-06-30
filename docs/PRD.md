# Product Requirements Document — Groupie Tracker Geolocalization

> Status: Draft · Last updated: 2026-06-30 · Owner: gtzimoka

## 1. Overview

Groupie Tracker Geolocalization extends the existing **groupie-tracker** project
(`../groupie-tracker`). The base already fetches the Groupie Trackers API,
lists/searches artists, shows an artist page, and exposes per-artist JSON
sub-endpoints. This project **adds geolocalization**: it converts each concert
location (a text address such as `north_carolina-usa`) into geographic
coordinates via **geocoding** and renders them as **map markers**, connected by
a line in chronological order (tour path).

The backend stays in **Go, standard library only**. Geocoding uses OpenStreetMap
**Nominatim** (free, key-less); the map uses **Leaflet**, vendored locally so it
works without internet access during the audit.

## 2. Relationship to the base project

We keep the base project's conventions and code, and add a thin geolocalization
layer on top. Import paths use the module `groupie-tracker-geolocalization`.

### Inherited from the base (do not rebuild)

| Capability | Where it lives in the base |
|---|---|
| Artist listing + **search & filters** (name/creation/members) | `static/js/main.js`, `templates/index.html` |
| Artist page (name, image, members, first album, creation date) | `templates/artist.html` |
| Per-artist JSON sub-endpoints `/artist/{id}/{members,locations,dates,relations}` | `handlers/handlers.go` |
| Concurrent startup cache of all 4 API endpoints | `handlers/cacheAPI()` |
| External API client | `api/client.go` |
| Data models | `models/models.go` |
| Error pages `404/405/500` + `checkMethod` | `handlers/template.go`, `templates/` |
| Go 1.22 routing with `{id}` + `r.PathValue` | `handlers/handlers.go` |

> The search/filter requirement is therefore already satisfied; this PRD does not
> re-specify it.

### Added by this project

| Addition | New location |
|---|---|
| Geocoding (Nominatim) + throttling | `geocode/geocode.go` |
| Coordinate cache (memory + disk) | `geocode/cache.go`, `data/geocache.json` |
| Coordinates JSON endpoint `/artist/{id}/coordinates` | `handlers/handlers.go` |
| Leaflet map + markers + tour path | `static/js/map.js`, vendored `static/vendor/leaflet/` |
| Map container on the artist page | `templates/artist.html` (edit) |

## 3. Goals & Non-Goals

### Goals
- Convert each concert location to coordinates through real geocoding.
- Show an artist's concert locations as markers on a map on the artist page.
- Connect markers with a line in chronological order (tour path).
- Never crash; reuse the base's graceful error handling.
- Standard library only; cover new logic with unit tests.

### Non-Goals
- Re-implementing anything already provided by the base (listing, search, JSON
  sub-endpoints, error pages).
- User accounts, editing data, or paid map providers / third-party Go modules.

## 4. Constraints (from the exercise)

| Constraint | Source |
|---|---|
| Backend in Go | exercise.md:11 |
| Handle all website errors; never crash | exercise.md:13, principles.md:27 |
| Respect Go good practices | exercise.md:15 |
| Unit tests recommended (we treat as required where logic exists) | exercise.md:17 |
| **Only standard Go packages** | exercise.md:20 |
| Free choice of Map API | exercise.md:8 |

> The "only standard packages" rule applies to **Go imports**. Calling Nominatim
> over HTTP with `net/http` is allowed; importing a third-party Go SDK is not.
> Leaflet is browser-side JavaScript, not a Go dependency.

## 5. Functional Requirements

### FR-1 Geocoding (new)
- `geocode.Normalize(raw)` converts the raw API location string
  (e.g. `north_carolina-usa`) into a human-readable query: replace `_` with
  spaces, split on the **last** `-` into `City, Country`, title-cased — country
  upper-cased when ≤3 letters (e.g. `North Carolina, USA`; `Mainz, Germany`).
  This is the form **shown to users** and **sent to Nominatim**. The base does
  the equivalent client-side in `main.js` (`formatLocation`); `geocode` does it
  server-side.
- `geocode.Geocode(query)` returns `(lat, lng)` for that normalized query via
  Nominatim.
- Send a valid `User-Agent` (Nominatim policy) and **throttle to ~1 request/sec**
  for uncached lookups to respect the usage policy.
- Addresses that cannot be geocoded are skipped with a logged warning; the rest
  still render.

### FR-2 Geocoding cache (new)
- An in-memory map backed by `data/geocache.json`, **keyed by the normalized
  query** (e.g. `Mainz, Germany`), valued `{lat, lng}`.
- On lookup: serve from memory, else from disk, else call Nominatim and persist.
- This keeps repeat loads instant and keeps Nominatim traffic within policy.

### FR-3 Coordinates endpoint (new)
- `GET /artist/{id}/coordinates` returns JSON for that artist's concerts:
  `[{ "address": "...", "lat": 0.0, "lng": 0.0, "dates": ["..."] }]`.
- Built from the artist's **`relation.DatesLocations`** (map of location →
  its dates), which already pairs each location with its concert dates.
- The `address` field is the **normalized, human-readable form**
  (e.g. `Mainz, Germany`), not the raw API string.
- Dates arrive as `DD-MM-YYYY`, sometimes prefixed with `*`: strip the `*` and
  parse with the layout `02-01-2006` before comparing.
- The array is **sorted by the earliest concert date of each location** (the
  ordering key for the tour path; ties broken by address for determinism).
- Follows the base's existing `/artist/{id}/…` JSON pattern and `checkMethod`.

### FR-4 Map with markers (new)
- The artist page renders a Leaflet map container.
- `static/js/map.js` fetches `/artist/{id}/coordinates`, shows a loading state
  while it resolves, then drops one marker per location, fitting the map bounds
  to the markers.

### FR-5 Tour path — bonus (new)
- Draw a Leaflet polyline through the markers in the chronological order from
  FR-3, so the touring path is visible.

## 6. Non-Functional Requirements

- **Reliability:** reuse the base's error handling; geocoding failures never
  crash the server or block the page.
- **Correctness:** the audit artists (Queen, ACDC, Imagine Dragons, Guns N'
  Roses, Post Malone, Red Hot Chili Peppers) render markers at the expected
  cities (see audit.md).
- **Performance:** coordinates are cached on disk; uncached geocoding is
  throttled but only happens once per location ever.
- **Consistency:** match the base's package layout, naming, and routing style.
- **Offline-safe audit:** Leaflet is vendored, not loaded from a CDN.

## 7. Architecture

Same hybrid model as the base: Go serves HTML via `html/template` and exposes
JSON endpoints that the frontend JavaScript consumes.

```
Browser ──GET /────────────────────► homeHandler ──► cache.Artists ──► index.html
                                       (search/filter handled in main.js)
Browser ──GET /artist/{id}──────────► artistHandler ─► artist.html (+ empty map)
map.js  ──GET /artist/{id}/coordinates─► coordinatesHandler
                                          ├─ cache: artist's relation.DatesLocations (location → dates)
                                          ├─ geocode: normalize raw→"City, Country" ─(memory→disk→Nominatim)→ lat/lng
                                          └─ JSON: [{address:"City, Country", lat, lng, dates}] sorted by earliest date
map.js  ──► Leaflet: markers + polyline (tour path)
```

### Project structure

Bootstrapped from the base project, then extended. New/changed items are marked.

```
groupie-tracker-geolocalization/
├── main.go                         # newServer(addr) + ListenAndServe (from base)
├── go.mod                          # module groupie-tracker-geolocalization, go 1.22
├── api/
│   ├── client.go                   # GetArtists/Locations/Dates/Relations (from base)
│   └── client_test.go
├── models/
│   └── models.go                   # Artist, Location, Date, Relation, Cache (structs only, from base)
├── geocode/                        # NEW: geocoding layer
│   ├── geocode.go                  # Normalize(raw) + Geocode(query) via Nominatim + throttle + User-Agent
│   ├── geocode_test.go             # parses saved fixtures from testdata/
│   ├── cache.go                    # in-memory + data/geocache.json persistence
│   ├── cache_test.go
│   └── testdata/                   # saved Nominatim responses (input fixtures)
├── handlers/
│   ├── handlers.go                 # base routes + NEW coordinatesHandler & route
│   ├── handlers_test.go            # + tests for /artist/{id}/coordinates
│   ├── template.go                 # renderTemplate, render404/405/500 (from base)
│   └── template_test.go
├── templates/
│   ├── index.html                  # from base (listing + search/filters)
│   ├── artist.html                 # EDIT: add #concert-map container
│   ├── 404.html
│   ├── 405.html
│   └── 500.html
├── static/
│   ├── css/style.css               # + map styling
│   ├── js/
│   │   ├── main.js                 # from base (search/filters, detail fetch)
│   │   └── map.js                  # NEW: Leaflet init, fetch coordinates, markers + path
│   ├── vendor/leaflet/             # NEW: vendored leaflet.js, leaflet.css, marker images
│   └── images/
├── data/geocache.json              # NEW: auto-generated coordinate cache
├── docs/
│   ├── PRD.md
│   └── MILESTONES.md
├── README.md
└── LICENSE
```

### Package responsibilities

- **api** — fetch + parse the Groupie Trackers API (base). Depends on `net/http`,
  `encoding/json`.
- **models** — typed structs for API data + the startup `Cache` (base).
- **geocode** *(new)* — `Normalize(raw)` produces the human-readable query;
  `Geocode(query)` returns `(lat, lng)`, throttled and cached. Depends on
  `net/http`, `encoding/json`, `os`, `strings`, `sync`, `time`.
- **handlers** — routing, page rendering, JSON endpoints, error pages (base);
  adds `coordinatesHandler`. Depends on `api`, `models`, `geocode`,
  `html/template`, `net/http`.

## 8. Error Handling

Reuse the base's mechanisms; add geocoding-specific paths.

| Situation | Behavior |
|---|---|
| Unknown artist / bad id | `render404` (base) |
| Non-GET method | `render405` via `checkMethod` (base) |
| Template failure | `render500` (base) |
| Address not geocodable | skip that location, log warning, return the rest |
| Nominatim unreachable for an uncached address | omit that marker; the endpoint still returns the locations it could resolve; `map.js` shows a non-blocking notice (like the base's `loadArtistDetail` catch) |

## 9. Testing

Table-driven unit tests for every package with logic; no network in tests.

- **api** — JSON parsing against an `httptest.Server` (base).
- **geocode** *(new)* — Nominatim response parsing from saved fixtures in
  `geocode/testdata/` (offline) and cache hit/miss/persist behavior.
- **handlers** *(new coverage)* — `/artist/{id}/coordinates` returns the right
  fields and is sorted by earliest date; bad id → 404; non-GET → 405.
- **template** — templates render with data (base).

`testdata/` holds only captured input fixtures, not generated golden output.

## 10. Milestones

Delivery is broken into M1–M7. See [MILESTONES.md](MILESTONES.md).

## 11. Open Questions

- Whether to commit a small pre-seeded `data/geocache.json` (covering the audit
  artists) so the first demo needs no live Nominatim calls.
