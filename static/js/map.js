// Renders an artist's concert locations on a Leaflet map.

async function initConcertMap() {
    const mapEl = document.querySelector("#concert-map");
    const hero = document.querySelector(".artist-hero");
    if (!mapEl || !hero) {
        return; // not on the artist page
    }

    const artistID = hero.dataset.artistId;

    // Create the map with a neutral world view until markers arrive.
    const worldBounds = [[-85, -180], [85, 180]];
    const map = L.map(mapEl, {
        maxBounds: worldBounds,
        maxBoundsViscosity: 1.0,
        zoomSnap: 0, // allow a fractional zoom so the world fills the box snugly
    });
    L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
        attribution: "&copy; OpenStreetMap contributors",
        maxZoom: 19,
        noWrap: true,
    }).addTo(map);

    // Fill the box edge-to-edge horizontally (the world spans the full width);
    // top/bottom are cropped. World pixel width at zoom z is 256 * 2^z, so the
    // zoom that fills the container width is log2(width / 256).
    const fillZoom = Math.log2(map.getSize().x / 256);
    map.setView([25, 0], fillZoom);
    map.setMinZoom(fillZoom);

    try {
        const response = await fetch(`/artist/${artistID}/coordinates`);
        if (!response.ok) {
            throw new Error(`request failed: ${response.status}`);
        }
        const places = await response.json();

        if (places.length === 0) {
            mapEl.insertAdjacentHTML(
                "afterend",
                '<p class="map-caption">No mappable concert locations for this artist.</p>'
            );
            return;
        }

        // Draw the tour path connecting concerts in chronological order.
        const latlngs = places.map((place) => [place.lat, place.lng]);
        if (latlngs.length > 1) {
            L.polyline(latlngs, { color: "#3b82f6", weight: 3, opacity: 0.7 }).addTo(map);
        }

        // One coloured dot per concert: first = blue, last = green, rest = grey.
        places.forEach((place, i) => {
            let fillColor = "#777";
            if (i === places.length - 1) fillColor = "green";
            if (i === 0) fillColor = "red";

            L.circleMarker([place.lat, place.lng], {
                radius: 7,
                weight: 2,
                color: "#fff",
                fillColor: fillColor,
                fillOpacity: 1,
            })
                .bindPopup(`<strong>${place.address}</strong><br>${place.dates.join(", ")}`)
                .addTo(map);
        });
    } catch (error) {
        mapEl.insertAdjacentHTML(
            "afterend",
            '<p class="error-text">Could not load the concerts map.</p>'
        );
    }
}

initConcertMap();
