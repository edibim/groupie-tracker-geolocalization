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

    // Temporary world view while the concert locations are loading.
    const fillZoom = Math.log2(map.getSize().x / 256);
    map.setView([25, 0], fillZoom);

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
            L.polyline(latlngs, {
                color: "#3b82f6",
                weight: 3,
                opacity: 0.7,
            }).addTo(map);
        }

        // Keep track of all marker positions so the map can automatically
        // zoom and center to fit every concert location.
        const bounds = L.latLngBounds();

        // One coloured dot per concert: first = red, last = green, rest = grey.
        places.forEach((place, i) => {
            let fillColor = "#777";
            if (i === places.length - 1) fillColor = "red";
            if (i === 0) fillColor = "green";

            bounds.extend([place.lat, place.lng]);

            L.circleMarker([place.lat, place.lng], {
                radius: 7,
                weight: 2,
                color: "#fff",
                fillColor: fillColor,
                fillOpacity: 1,
            })
                    .bindTooltip(`<strong>${place.address}</strong><br>${place.dates.join(", ")}`, {
                    direction: "top",
                    offset: [0, -6],
                })

                .addTo(map);
        });

        // Automatically fit the map to include all concert locations.
        map.fitBounds(bounds, {
            padding: [40, 40],
        });

        // Legend explaining the marker colours.
        const legend = L.control({ position: "bottomright" });

        legend.onAdd = () => {
            const div = L.DomUtil.create("div", "map-legend");
            div.innerHTML =
                '<span><i style="background: green"></i> First concert</span>' +
                '<span><i style="background: red"></i> Last concert</span>' +
                '<span><i style="background: #777"></i> Other</span>';
            return div;
        };

        legend.addTo(map);
    } catch (error) {
        mapEl.insertAdjacentHTML(
            "afterend",
            '<p class="error-text">Could not load the concerts map.</p>'
        );
    }
}

initConcertMap();