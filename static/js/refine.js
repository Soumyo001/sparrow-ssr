(() => {
    "use strict";

    const app = document.getElementById("refine-app");
    if (!app) return;

    const elements = {
        searchForm: document.getElementById("refine-search-form"),
        destination: document.getElementById("refine-destination"),
        guests: document.getElementById("refine-guests"),
        headingDestination: document.getElementById("refine-heading-destination"),
        breadcrumbDestination: document.getElementById("refine-breadcrumb-destination"),
        maxPrice: document.getElementById("refine-max-price"),
        poolOnly: document.getElementById("refine-pool-only"),
        sort: document.getElementById("refine-sort"),
        dates: document.getElementById("refine-dates"),
        dateDropdown: document.getElementById("refine-date-dropdown"),
        dateStart: document.getElementById("refine-date-start"),
        dateEnd: document.getElementById("refine-date-end"),
        dateApply: document.getElementById("refine-date-apply"),
        dateClear: document.getElementById("refine-date-clear"),
        status: document.getElementById("refine-status"),
        grid: document.getElementById("refine-grid"),
        dialog: document.getElementById("refine-dialog"),
        dialogTitle: document.getElementById("refine-dialog-title"),
        dialogMessage: document.getElementById("refine-dialog-message"),
        closeDialog: document.getElementById("refine-dialog-close")
    };

    const state = {
        search: "",
        pax: 2,
        dateStart: "",
        dateEnd: "",
        properties: [],
        favorites: new Set()
    };

    let latestRequest = 0;

    // Read the search criteria from the browser URL.
    function readURL() {
        const params = new URLSearchParams(window.location.search);

        state.search = (params.get("search") || "").trim();

        const pax = Number(params.get("pax") || "2");
        state.pax =
            Number.isInteger(pax) && pax >= 1 && pax <= 20
                ? pax
                : 2;

        elements.destination.value = state.search;
        elements.guests.value = String(state.pax);

        state.dateStart = params.get("dateStart") || "";
        state.dateEnd = params.get("dateEnd") || "";

        elements.dateStart.value = state.dateStart;
        elements.dateEnd.value = state.dateEnd;

        updateDateSummary();

        const name = state.search || "your destination";

        elements.headingDestination.textContent = name;
        elements.breadcrumbDestination.textContent = name;
    }

    function formatDateLabel(isoDate) {
        if (!/^\d{4}-\d{2}-\d{2}$/.test(isoDate)) {
            return isoDate;
        }

        const date = new Date(`${isoDate}T00:00:00Z`);

        if (Number.isNaN(date.getTime())) {
            return isoDate;
        }

        return date.toLocaleDateString("en-US", {
            month: "short",
            day: "2-digit",
            timeZone: "UTC"
        });
    }

    function updateDateSummary() {
        elements.dates.textContent =
            state.dateStart && state.dateEnd
                ? `${formatDateLabel(state.dateStart)} – ` +
                  `${formatDateLabel(state.dateEnd)}`
                : "Dates";
    }

    // Update the page URL without a full reload.
    function updateURL() {
        const params = new URLSearchParams({
            search: state.search,
            pax: String(state.pax),
            order: "1"
        });

        if (state.dateStart && state.dateEnd) {
            params.set("dateStart", state.dateStart);
            params.set("dateEnd", state.dateEnd);
        }

        window.history.pushState(
            {},
            "",
            `/refine?${params.toString()}`
        );
    }

    // Create DOM nodes safely using textContent.
    function el(tag, className, text) {
        const node = document.createElement(tag);

        if (className) {
            node.className = className;
        }

        if (text !== undefined) {
            node.textContent = String(text);
        }

        return node;
    }

    function formatPrice(property) {
        const amount = Number(property.price_from);

        if (!Number.isFinite(amount) || amount <= 0) {
            return "";
        }

        const formatted = amount.toLocaleString("en-BD");

        if (property.currency === "BDT") {
            return `৳${formatted}`;
        }

        return `${property.currency} ${formatted}`;
    }

    function imageURL(property) {
        const raw = Array.isArray(property.images)
            ? property.images[0]
            : "";

        if (!raw) return "";

        try {
            const url = new URL(raw);

            if (url.protocol === "https:" ||
                url.protocol === "http:") {
                return url.href;
            }
        } catch {
            return "";
        }

        return "";
    }

    // Accept only manually verified property-page links from the
    // provider named on the listing. Never infer a URL from a name.
    function allowedBookingURL(property) {
        const raw = property.booking_url;

        if (typeof raw !== "string" || !raw.trim()) {
            return "";
        }

        const provider = String(
            property.provider_name || ""
        ).trim().toLowerCase();

        const domainByProvider = {
            "booking.com": "booking.com",
            "expedia": "expedia.com",
            "expedia.com": "expedia.com"
        };

        const domain = domainByProvider[provider];
        if (!domain) return "";

        try {
            const url = new URL(raw.trim());
            const hostname = url.hostname.toLowerCase();

            const allowedHost =
                hostname === domain ||
                hostname.endsWith(`.${domain}`);

            if (url.protocol !== "https:" ||
                !allowedHost ||
                url.username || url.password ||
                url.port || url.pathname === "/") {
                return "";
            }

            return url.href;
        } catch {
            return "";
        }
    }

    function showDialog(title, message) {
        elements.dialogTitle.textContent = title;
        elements.dialogMessage.textContent = message;

        if (typeof elements.dialog.showModal === "function") {
            elements.dialog.showModal();
        } else {
            window.alert(`${title}\n\n${message}`);
        }
    }

    // Construct one property card from one API object.
    function propertyCard(property) {
        const card = el("article", "refine-card");
        const picture = el("div", "refine-card-picture");

        const fallback = el(
            "div",
            "refine-image-fallback",
            "Photo unavailable"
        );

        const photo = imageURL(property);

        if (photo) {
            const img = el("img", "refine-card-image");

            img.src = photo;
            img.alt =
                `Illustrative accommodation photograph for ${property.name}`;
            img.loading = "lazy";

            img.addEventListener(
                "error",
                () => {
                    img.remove();
                    picture.append(fallback);
                },
                { once: true }
            );

            picture.append(img);
        } else {
            picture.append(fallback);
        }

        const displayedPrice = formatPrice(property);
        picture.append(
            el(
                "span",
                "refine-price",
                displayedPrice
                    ? `From ${displayedPrice}`
                    : "Check provider for price"
            )
        );

        if (property.eco_friendly) {
            picture.append(
                el("span", "refine-eco", "Eco")
            );
        }

        // Favorites are kept in browser memory only.
        const favorite = el(
            "button",
            "refine-favorite",
            state.favorites.has(property.id) ? "♥" : "♡"
        );

        favorite.type = "button";
        favorite.setAttribute(
            "aria-label",
            `Favorite ${property.name}`
        );
        favorite.setAttribute(
            "aria-pressed",
            String(state.favorites.has(property.id))
        );

        favorite.addEventListener("click", () => {
            if (state.favorites.has(property.id)) {
                state.favorites.delete(property.id);
            } else {
                state.favorites.add(property.id);
            }

            const selected = state.favorites.has(property.id);

            favorite.textContent = selected ? "♥" : "♡";
            favorite.setAttribute(
                "aria-pressed",
                String(selected)
            );
        });

        picture.append(favorite);

        const body = el("div", "refine-card-body");
        const meta = el("div", "refine-card-meta");

        meta.append(
            el(
                "span",
                "refine-rating",
                `● ${Number(property.rating).toFixed(1)} ` +
                `(${property.review_count} Reviews)`
            ),
            el(
                "span",
                "refine-property-type",
                property.property_type
            )
        );

        body.append(meta);

        body.append(
            el("h2", "refine-card-title", property.name)
        );

        const details = [
            property.size_sqft
                ? `${property.size_sqft} ft²`
                : "",
            property.property_type,
            ...(Array.isArray(property.amenities)
                ? property.amenities
                : [])
        ].filter(Boolean).join(" · ");

        body.append(
            el("p", "refine-card-details", details)
        );

        const location = [
            property.location?.city,
            property.location?.area,
            property.location?.state
        ].filter(Boolean).join(" › ");

        body.append(
            el("p", "refine-card-location", location)
        );

        const bottom = el("div", "refine-card-bottom");

        bottom.append(
            el(
                "span",
                "refine-provider",
                property.provider_name
            )
        );

        const availability = el(
            "button",
            "refine-availability",
            "View Availability"
        );

        availability.type = "button";

        availability.addEventListener("click", () => {
            const bookingURL = allowedBookingURL(property);

            if (bookingURL) {
                // Navigate to the supplied provider URL.
                // Its property identity must be verified separately.
                // Dates and prices are not transferred here.
                window.location.assign(bookingURL);
                return;
            }

            showDialog(
                property.name,
                "This is a demo listing. Live dates, prices, " +
                "booking links and availability are not connected."
            );
        });

        bottom.append(availability);
        body.append(bottom);

        card.append(picture, body);

        return card;
    }

    // Apply additional browser-side filters and sorting.
    function render() {
        const pricesAvailable = state.properties.some(
            property => Number(property.price_from) > 0
        );
    
        elements.maxPrice.disabled = !pricesAvailable;
        elements.maxPrice.placeholder = pricesAvailable
            ? "Any price"
            : "Price unavailable";
    
        for (const value of ["price_low", "price_high"]) {
            const option = elements.sort.querySelector(
                `option[value="${value}"]`
            );
        
            if (option) {
                option.disabled = !pricesAvailable;
            }
        }
    
        if (!pricesAvailable) {
            elements.sort.value = "popular";
        }
    
        const ceiling = pricesAvailable
            ? elements.maxPrice.value.trim()
            : "";
    
        const maxPrice = ceiling === ""
            ? Infinity
            : Math.max(0, Number(ceiling));
    
        const poolOnly = elements.poolOnly.checked;

        const filtered = state.properties.filter(property => {
            const price = Number(property.price_from);
            if (maxPrice !== Infinity &&
                (!Number.isFinite(price) ||
                 price <= 0 ||
                 price > maxPrice)) {
                return false;
            }

            if (poolOnly &&
                !property.amenities?.includes("Pool")) {
                return false;
            }

            return true;
        });

        switch (elements.sort.value) {
            case "price_low":
                filtered.sort((a, b) =>
                    a.price_from - b.price_from ||
                    a.id.localeCompare(b.id)
                );
                break;

            case "price_high":
                filtered.sort((a, b) =>
                    b.price_from - a.price_from ||
                    a.id.localeCompare(b.id)
                );
                break;

            default:
                filtered.sort((a, b) =>
                    b.popularity_score - a.popularity_score ||
                    a.id.localeCompare(b.id)
                );
        }

        elements.grid.replaceChildren();

        elements.status.textContent =
            `${filtered.length} of ${state.properties.length} properties shown. ` +
            `Demo date filtering only; confirm price, guests, and availability on the provider site.`;

        if (filtered.length === 0) {
            elements.grid.append(
                el(
                    "p",
                    "refine-empty",
                    "No properties match your search or filters."
                )
            );
            return;
        }

        const fragment = document.createDocumentFragment();

        filtered.forEach(property => {
            fragment.append(propertyCard(property));
        });

        elements.grid.append(fragment);
    }

    // Request JSON from the verified Beego endpoint.
    async function load() {
        const currentRequest = ++latestRequest;

        elements.status.textContent = "Loading properties...";
        elements.grid.replaceChildren();

        const params = new URLSearchParams({
            search: state.search,
            pax: String(state.pax),
            order: "1"
        });

        if (state.dateStart || state.dateEnd) {
            params.set("dateStart", state.dateStart);
            params.set("dateEnd", state.dateEnd);
        }

        try {
            const response = await fetch(
                `/api/properties?${params.toString()}`,
                {
                    headers: {
                        Accept: "application/json"
                    }
                }
            );

            const data = await response.json();

            if (!response.ok) {
                throw new Error(
                    data.error || `HTTP ${response.status}`
                );
            }

            if (!Array.isArray(data.properties)) {
                throw new Error("Unexpected API response");
            }

            // Ignore responses from older requests.
            if (currentRequest !== latestRequest) {
                return;
            }

            state.properties = data.properties;
            render();

        } catch (error) {
            if (currentRequest !== latestRequest) {
                return;
            }

            elements.status.textContent =
                `Unable to load properties: ${error.message}`;
        }
    }

    // New destination search.
    elements.searchForm.addEventListener(
        "submit",
        event => {
            event.preventDefault();

            state.search = elements.destination.value.trim();
            state.pax = Number(elements.guests.value);

            updateURL();
            readURL();
            load();
        }
    );

    // Guest-count changes request new API results.
    elements.guests.addEventListener("change", () => {
        if (!elements.guests.checkValidity()) {
            elements.guests.reportValidity();
            return;
        }

        state.search = elements.destination.value.trim();
        state.pax = Number(elements.guests.value);

        updateURL();
        readURL();
        load();
    });

    // These controls work on data already fetched.
    elements.maxPrice.addEventListener("input", render);
    elements.poolOnly.addEventListener("change", render);
    elements.sort.addEventListener("change", render);

    elements.dateApply.addEventListener("click", () => {
        const start = elements.dateStart.value;
        const end = elements.dateEnd.value;

        // Both dates must be selected.
        if ((start && !end) || (!start && end)) {
            showDialog(
                "Invalid dates",
                "Select both check-in and checkout."
            );
            return;
        }

        // Checkout must follow check-in.
        if (start && end <= start) {
            showDialog(
                "Invalid dates",
                "Checkout must be after check-in."
            );
            return;
        }

        state.dateStart = start;
        state.dateEnd = end;

        updateURL();
        updateDateSummary();

        elements.dateDropdown.open = false;

        // Re-fetch because date filtering happens
        // in the Beego API, not only in the browser.
        load();
    });

    elements.dateClear.addEventListener("click", () => {
        state.dateStart = "";
        state.dateEnd = "";

        elements.dateStart.value = "";
        elements.dateEnd.value = "";

        updateURL();
        updateDateSummary();

        elements.dateDropdown.open = false;

        load();
    });

    elements.closeDialog.addEventListener(
        "click",
        () => elements.dialog.close()
    );

    // Support browser Back/Forward navigation.
    window.addEventListener("popstate", () => {
        readURL();
        load();
    });

    readURL();
    load();
})();
