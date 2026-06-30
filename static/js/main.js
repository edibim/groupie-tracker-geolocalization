const artistHero = document.querySelector(".artist-hero");
const artistCards = document.querySelectorAll(".artist-card");
const artistSearch = document.querySelector("#artist-search");
const creationFilter = document.querySelector("#creation-filter");
const membersFilter = document.querySelector("#members-filter");
const artistCount = document.querySelector("#artist-count");
const noResults = document.querySelector("#no-results");

function formatLocation(raw) {
    const lastDash = raw.lastIndexOf("-");

    if (lastDash === -1) {
        return toTitleCase(raw.replaceAll("_", " "));
    }

    const city = raw.slice(0, lastDash).replaceAll("_", " ").replaceAll("-", " ");
    const country = raw.slice(lastDash + 1).replaceAll("_", " ");
    const countryName = country.length <= 3 ? country.toUpperCase() : toTitleCase(country);

    return `${toTitleCase(city)}, ${countryName}`;
}

function toTitleCase(value) {
    return value.replace(/\b\w/g, (letter) => letter.toUpperCase());
}

function listMarkup(items, formatter) {
    if (!items || items.length === 0) {
        return "<p class=\"empty-state\">No data found.</p>";
    }

    return `
        <ul class="detail-list">
            ${items.map((item) => `<li>${formatter ? formatter(item) : item.replace("*", "")}</li>`).join("")}
        </ul>
    `;
}

function relationsMarkup(relations) {
    const entries = Object.entries(relations || {});

    if (entries.length === 0) {
        return "<p class=\"empty-state\">No relations found.</p>";
    }

    return entries.map(([location, dates]) => `
        <div class="relation-row">
            <strong>${formatLocation(location)}</strong>
            <span>${dates.map((date) => date.replace("*", "")).join(", ")}</span>
        </div>
    `).join("");
}

async function loadArtistDetail(detail, output, render) {
    const artistID = artistHero.dataset.artistId;

    try {
        const response = await fetch(`/artist/${artistID}/${detail}`);

        if (!response.ok) {
            throw new Error(`Request failed with status ${response.status}`);
        }

        const data = await response.json();
        output.innerHTML = render(data);
    } catch (error) {
        output.innerHTML = "<p class=\"error-text\">Could not load this artist data.</p>";
    }
}

function matchesCreationFilter(year, filter) {
    if (filter === "before-1980") return year < 1980;
    if (filter === "1980-1999") return year >= 1980 && year <= 1999;
    if (filter === "2000-2009") return year >= 2000 && year <= 2009;
    if (filter === "2010-plus") return year >= 2010;
    return true;
}

function matchesMembersFilter(count, filter) {
    if (filter === "solo") return count === 1;
    if (filter === "small") return count >= 2 && count <= 4;
    if (filter === "large") return count >= 5;
    return true;
}

function filterArtists() {
    const query = artistSearch.value.trim().toLowerCase();
    const creationValue = creationFilter.value;
    const membersValue = membersFilter.value;
    let visibleCount = 0;

    artistCards.forEach((card) => {
        const text = card.dataset.search.toLowerCase();
        const created = Number(card.dataset.created);
        const members = Number(card.dataset.members);
        const isVisible = text.includes(query)
            && matchesCreationFilter(created, creationValue)
            && matchesMembersFilter(members, membersValue);

        card.classList.toggle("hidden", !isVisible);

        if (isVisible) visibleCount++;
    });

    artistCount.textContent = visibleCount;
    noResults.classList.toggle("hidden", visibleCount !== 0);
}

if (artistCards.length > 0 && artistSearch && creationFilter && membersFilter) {
    [artistSearch, creationFilter, membersFilter]
        .forEach((control) => control.addEventListener("input", filterArtists));
}

if (artistHero) {
    const locationsOutput = document.querySelector("#locations-output");
    const datesOutput = document.querySelector("#dates-output");
    const relationsOutput = document.querySelector("#relations-output");

    loadArtistDetail("locations", locationsOutput, (data) => listMarkup(data, formatLocation));
    loadArtistDetail("dates", datesOutput, (data) => listMarkup(data));
    loadArtistDetail("relations", relationsOutput, relationsMarkup);
}
