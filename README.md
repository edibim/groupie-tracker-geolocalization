# groupie-tracker-geolocalization

A web application that maps the concert locations of a chosen artist/band. It
extends the base [groupie-tracker](../groupie-tracker) project by converting each
concert location into geographic coordinates (geocoding) and plotting them on an
interactive map, connected by a line in chronological tour order.

The backend is written in **Go using only the standard library**.

## Features

- Browse and search artists (inherited from the base project).
- Per-artist concert **map** with one marker per location; hovering over a
  marker shows the city and its concert dates (one per line) — no click needed.
- **Tour path**: markers are joined by a line in chronological order — the first
  concert is shown in **green**, the last in **red**, as explained by the on-map
  legend — with a travelling-light animation showing the tour direction.
- **Geocoding** via OpenStreetMap Nominatim, with an on-disk **cache** so repeat
  loads are instant and requests stay within Nominatim's usage policy.
- Concert **dates** and **dates & locations** are listed in chronological order.
- Graceful handling: locations that cannot be geocoded are skipped (and logged);
  the rest still render.

## Requirements

- Go 1.22 or newer.
- Internet access at runtime: the app fetches the Groupie Trackers API, geocodes
  via Nominatim, and loads OpenStreetMap map tiles.

## Run

```bash
go run .
```

Then open <http://localhost:8080>.

> The first time you open an artist, uncached locations are geocoded one per
> second (Nominatim policy), so the markers may take a few seconds to appear.
> Subsequent loads are instant because the results are cached in
> `data/geocache.json`.

## Test

```bash
go test ./...
```

## Routes

| Route | Description |
|---|---|
| `/` | Artist list + search/filters |
| `/artist/{id}` | Artist page (info + map) |
| `/artist/{id}/coordinates` | JSON: geocoded concert locations, sorted by earliest date |
| `/artist/{id}/{members,locations,dates,relations}` | JSON sub-resources (base) |
| `/static/...` | CSS, JS, vendored Leaflet, images |

## Project structure

```
main.go            HTTP server entry point
api/               client for the Groupie Trackers API
models/            data structs
geocode/           Nominatim geocoding + on-disk cache + request throttling
handlers/          routes, page rendering, JSON endpoints, error pages
templates/         HTML (index, artist, 404/405/500)
static/            CSS, JS (main.js, map.js), vendored Leaflet
data/geocache.json auto-generated geocoding cache
docs/              PRD and milestones
```

## Notes on the map

- The Leaflet **library** is vendored under `static/vendor/leaflet/`, so it loads
  without a CDN. The map **tiles**, however, are served by OpenStreetMap and
  require internet access.
- Geocoding uses Nominatim with a descriptive `User-Agent` and ~1 request/second
  throttling, as required by its usage policy.

## Credits

- Concert data: [Groupie Trackers API](https://groupietrackers.herokuapp.com/api)
- Geocoding: [OpenStreetMap Nominatim](https://nominatim.openstreetmap.org/)
- Map rendering: [Leaflet](https://leafletjs.com/) + OpenStreetMap tiles
