// Renders an artist's concert locations on a Leaflet map.

// Point Leaflet's default marker icons at our vendored copies, otherwise the
// markers fail to load (a well-known Leaflet gotcha).
delete L.Icon.Default.prototype._getIconUrl;
L.Icon.Default.mergeOptions({
    iconRetinaUrl: "/static/vendor/leaflet/images/marker-icon-2x.png",
    iconUrl: "/static/vendor/leaflet/images/marker-icon.png",
    shadowUrl: "/static/vendor/leaflet/images/marker-shadow.png",
});

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

        places.forEach((place) => {
            const marker = L.marker([place.lat, place.lng]);
            marker.bindPopup(`<strong>${place.address}</strong><br>${place.dates.join(", ")}`);
            marker.addTo(map);
        });
        // Markers stay on the full-world view; the user can zoom in manually.
    } catch (error) {
        mapEl.insertAdjacentHTML(
            "afterend",
            '<p class="error-text">Could not load the concerts map.</p>'
        );
    }
}

initConcertMap();
