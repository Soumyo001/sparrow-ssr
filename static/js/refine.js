
(() => {
    "use strict";

    const app = document.getElementById("refine-app") ||
        document.getElementById("country-results");
    if (!app) return;

    const country = (app.dataset.country || "").trim();
    const isCountry = country.length > 0;

    const $ = id => document.getElementById(id);
    const PAGE_SIZE = 6;

    const ui = {
        heading: $("refine-heading-destination"), breadcrumb: $("refine-breadcrumb-destination"),
        grid: $("refine-grid"), status: $("refine-status"), sort: $("refine-sort"),
        paginationWrap: $("refine-pagination-wrap"),
        pagination: $("refine-pagination"),
        pageSummary: $("refine-page-summary"),
        filter: $("refine-filter-dialog"), filterClose: $("refine-filter-close"),
        filterClear: $("refine-filter-clear"), filterSearch: $("refine-filter-search"),
        eco: $("refine-eco-only"), modalDates: $("refine-modal-dates"),
        modalDateSummary: $("refine-modal-date-summary"),
        guestSection: $("refine-guest-section"), guestNumber: $("refine-guest-number"),
        guestMinus: $("refine-guest-minus"), guestPlus: $("refine-guest-plus"),
        priceSection: $("refine-price-section"), priceMin: $("refine-min-price"),
        priceMax: $("refine-max-price"), rangeMin: $("refine-range-min"),
        rangeMax: $("refine-range-max"), rangeTrack: $("refine-range-track"),
        currencyMin: $("refine-price-currency-min"), currencyMax: $("refine-price-currency-max"),
        amenities: $("refine-amenities-grid"),
        calendar: $("refine-calendar-dialog"), calendarClose: $("refine-calendar-close"),
        calendarHint: $("refine-calendar-hint"), calendarMonths: $("refine-calendar-months"),
        monthPrev: $("refine-month-prev"), monthNext: $("refine-month-next"),
        calendarClear: $("refine-calendar-clear"), calendarApply: $("refine-calendar-apply"),
        message: $("refine-dialog"), messageTitle: $("refine-dialog-title"),
        messageText: $("refine-dialog-message"), messageClose: $("refine-dialog-close")
    };

    const state = {
        search: "", pax: 2, dateStart: "", dateEnd: "",
        properties: [], currentPage: 1,
        priceMin: null, priceMax: null, eco: false, amenities: new Set(),
        favorites: new Set()
    };
    let draft = null;
    let latestRequest = 0;
    let calendarOrigin = "standalone";
    let chosenStart = "", chosenEnd = "", visibleMonth = new Date();

    function node(tag, className = "", text = undefined) {
        const n = document.createElement(tag);
        if (className) n.className = className;
        if (text !== undefined) n.textContent = String(text);
        return n;
    }

    function parseURL() {
        const params = new URLSearchParams(location.search);
    
        state.search = isCountry
            ? country
            : (params.get("search") || "").trim();
    
        const pax = Number(params.get("pax") || 2);
        state.pax = Number.isInteger(pax) && pax >= 1 && pax <= 20
            ? pax
            : 2;
    
        state.dateStart = params.get("dateStart") || "";
        state.dateEnd = params.get("dateEnd") || "";
    
        if (!isCountry) {
            const name = state.search || "your destination";
            ui.heading.textContent = name;
            ui.breadcrumb.textContent = name;
        }
    
        updatePills();
    }

    function updateURL() {
        const params = new URLSearchParams({
            pax: String(state.pax),
            order: "1"
        });

        if (!isCountry) {
            params.set("search", state.search);
        }

        if (state.dateStart && state.dateEnd) {
            params.set("dateStart", state.dateStart);
            params.set("dateEnd", state.dateEnd);
        }

        const path = isCountry
            ? window.location.pathname
            : "/refine";

        history.pushState(
            {},
            "",
            `${path}?${params.toString()}`
        );
    }

    function updatePills() {
        const moreCount = (state.eco ? 1 : 0) + state.amenities.size;

        const active = {
            dates: Boolean(state.dateStart && state.dateEnd),
            price: state.priceMin !== null || state.priceMax !== null,
            guests: state.pax !== 2,
            more: moreCount > 0
        };

        let priceText = "Price";

        if (active.price) {
            const currency = priceBounds().currency;
            const symbols = {
                USD: "$",
                BDT: "৳",
                EUR: "€",
                GBP: "£"
            };

            priceText = `Price ${symbols[currency] || currency}`;
        }

        const labels = {
            dates: active.dates
                ? `${shortDate(state.dateStart)} – ${shortDate(state.dateEnd)}`
                : "Dates",
            price: priceText,
            guests: active.guests
                ? `${state.pax} Guests`
                : "Guests",
            more: "More"
        };

        document.querySelectorAll("[data-filter-chip]").forEach(chip => {
            const type = chip.dataset.filterChip;
            const opener = chip.querySelector("[data-filter-open]");
            const remove = chip.querySelector("[data-filter-remove]");

            opener.textContent = labels[type];
            opener.setAttribute("aria-pressed", String(active[type]));

            if (remove) {
                remove.hidden = !active[type];
            }
        });

        const countBadge = document.getElementById("refine-more-count");

        countBadge.textContent = String(moreCount);
        countBadge.hidden = moreCount === 0;
    }
    function priceAmount(p) {
        const n = Number(p.price_from);
        return Number.isFinite(n) && n > 0 ? n : 0;
    }
    function priceLabel(p) {
        const amount = priceAmount(p);
        if (!amount) return "";
        return p.currency === "BDT"
            ? `৳${amount.toLocaleString("en-US")}`
            : `${p.currency || ""} ${amount.toLocaleString("en-US")}`.trim();
    }
    function priceBounds() {
        const values = state.properties.map(priceAmount).filter(n => n > 0);
        const low = values.length ? Math.floor(Math.min(...values)) : 0;
        const high = values.length ? Math.ceil(Math.max(...values)) : 100;
        const currencies = new Set(state.properties.map(p => p.currency).filter(Boolean));
        const currency = currencies.size === 1 ? [...currencies][0] : "USD";
        return { low, high: Math.max(low + 1, high), currency };
    }
    function imageURL(property) {
        const raw = Array.isArray(property.images) ? property.images[0] : "";
        if (typeof raw !== "string" || !raw) return "";
        try {
            const url = new URL(raw);
            return url.protocol === "https:" || url.protocol === "http:" ? url.href : "";
        } catch { return ""; }
    }
    function bookingURL(property) {
        const raw = property.booking_url;
        if (typeof raw !== "string" || !raw.trim()) return "";
        const domains = { "booking.com": "booking.com", "expedia": "expedia.com", "expedia.com": "expedia.com" };
        const domain = domains[String(property.provider_name || "").trim().toLowerCase()];
        if (!domain) return "";
        try {
            const url = new URL(raw.trim());
            const host = url.hostname.toLowerCase();
            if (url.protocol !== "https:" ||
                !(host === domain || host.endsWith(`.${domain}`)) ||
                url.username || url.password || url.port || url.pathname === "/") return "";
            return url.href;
        } catch { return ""; }
    }
    function showMessage(title, message) {
        ui.messageTitle.textContent = title;
        ui.messageText.textContent = message;
        ui.message.showModal();
    }
    function propertyCard(property) {
        const card = node("article", "refine-card");
        const picture = node("div", "refine-card-picture");
        const fallback = node("div", "refine-image-fallback", "Photo unavailable");
        const photoURL = imageURL(property);
        if (photoURL) {
            const img = node("img", "refine-card-image");
            img.src = photoURL;
            img.alt = `Illustrative accommodation photograph for ${property.name}`;
            img.loading = "lazy";
            img.addEventListener("error", () => { img.remove(); picture.prepend(fallback); }, { once: true });
            picture.append(img);
        } else picture.append(fallback);

        picture.append(node("span", "refine-price", priceLabel(property) ? `From ${priceLabel(property)}` : "Check provider for price"));
        if (property.eco_friendly) picture.append(node("span", "refine-eco", "Eco"));
        const fav = node("button", "refine-favorite", state.favorites.has(property.id) ? "♥" : "♡");
        fav.type = "button";
        fav.setAttribute("aria-label", `Favorite ${property.name}`);
        fav.setAttribute("aria-pressed", String(state.favorites.has(property.id)));
        fav.addEventListener("click", () => {
            if (state.favorites.has(property.id)) state.favorites.delete(property.id);
            else state.favorites.add(property.id);
            fav.textContent = state.favorites.has(property.id) ? "♥" : "♡";
            fav.setAttribute("aria-pressed", String(state.favorites.has(property.id)));
        });
        picture.append(fav);

        const body = node("div", "refine-card-body");
        const meta = node("div", "refine-card-meta");
        const rating = node("span", "refine-rating");
        rating.append(node("span", "refine-rating-icon", "●"));
        rating.append(document.createTextNode(`${Number(property.rating || 0).toFixed(1)} (${property.review_count || 0} Reviews)`));
        meta.append(rating, node("span", "refine-property-type", property.property_type || ""));
        body.append(meta, node("h2", "refine-card-title", property.name));
        const details = [property.size_sqft > 0 ? `${property.size_sqft} ft²` : "", property.property_type,
            ...(Array.isArray(property.amenities) ? property.amenities : [])].filter(Boolean).join(" · ");
        body.append(node("p", "refine-card-details", details));
        const loc = property.location || {};
        body.append(node("p", "refine-card-location", [loc.city, loc.area, loc.state].filter(Boolean).join(" › ")));
        const bottom = node("div", "refine-card-bottom");
        const view = node("button", "refine-availability", "View Availability");
        view.type = "button";
        view.addEventListener("click", () => {
            const dest = bookingURL(property);
            if (dest) location.assign(dest);
            else showMessage(property.name, "Live booking information is not connected for this listing. Check with the provider.");
        });
        bottom.append(node("span", "refine-provider", property.provider_name || ""), view);
        body.append(bottom);
        card.append(picture, body);
        return card;
    }

    function pageButton(text, destination, disabled = false, selected = false) {
        const button = node("button", "refine-page-button", text);

        button.type = "button";
        button.disabled = disabled;

        if (selected) {
            button.classList.add("is-current");
            button.setAttribute("aria-current", "page");
            button.setAttribute(
                "aria-label",
                `Page ${destination}, current page`
            );
        } else {
            button.setAttribute(
                "aria-label",
                /^\d+$/.test(String(text))
                    ? `Go to page ${destination}`
                    : `${text} page`
            );
        }

        button.addEventListener("click", () => {
            if (button.disabled || state.currentPage === destination) {
                return;
            }

            state.currentPage = destination;

            // Filter and sort state remain unchanged.
            render();

            // Return the visitor to the beginning of the new results.
            ui.grid.focus({ preventScroll: true });
            ui.grid.scrollIntoView({
                behavior: "smooth",
                block: "start"
            });
        });

        return button;
    }

    function renderPagination(total) {
        if (total === 0) {
            ui.paginationWrap.hidden = true;
            ui.pagination.replaceChildren();
            ui.pageSummary.textContent = "";
            state.currentPage = 1;

            return { start: 0, end: 0 };
        }

        const pageCount = Math.ceil(total / PAGE_SIZE);

        // Prevent an invalid page after the result count changes.
        state.currentPage = Math.max(
            1,
            Math.min(state.currentPage, pageCount)
        );

        const start = (state.currentPage - 1) * PAGE_SIZE;
        const end = Math.min(start + PAGE_SIZE, total);

        ui.paginationWrap.hidden = false;
        ui.pageSummary.textContent =
            `Showing ${start + 1}–${end} of ${total} properties`;

        const controls = document.createDocumentFragment();

        if (pageCount > 1) {
            controls.append(
                pageButton(
                    "Previous",
                    state.currentPage - 1,
                    state.currentPage === 1
                )
            );

            for (let page = 1; page <= pageCount; page++) {
                controls.append(
                    pageButton(
                        String(page),
                        page,
                        false,
                        page === state.currentPage
                    )
                );
            }

            controls.append(
                pageButton(
                    "Next",
                    state.currentPage + 1,
                    state.currentPage === pageCount
                )
            );
        }

        ui.pagination.replaceChildren(controls);

        return { start, end };
    }

    function render() {
        updatePills();
        const results = state.properties.filter(p => {
            const price = priceAmount(p);
            if (state.priceMin !== null && (!price || price < state.priceMin)) return false;
            if (state.priceMax !== null && (!price || price > state.priceMax)) return false;
            if (state.eco && !p.eco_friendly) return false;
            if (![...state.amenities].every(a => Array.isArray(p.amenities) && p.amenities.includes(a))) return false;
            return true;
        });
        if (ui.sort.value === "price_low") results.sort((a,b) => priceAmount(a) - priceAmount(b) || a.id.localeCompare(b.id));
        else if (ui.sort.value === "price_high") results.sort((a,b) => priceAmount(b) - priceAmount(a) || a.id.localeCompare(b.id));
        else results.sort((a,b) => Number(b.popularity_score) - Number(a.popularity_score) || a.id.localeCompare(b.id));

        const { start, end } = renderPagination(results.length);

        const fragment = document.createDocumentFragment();

        if (!results.length) {
            fragment.append(
                node(
                    "p",
                    "refine-empty",
                    "No properties match your search or filters."
                )
            );
        } else {
            results.slice(start, end).forEach(property => {
                fragment.append(propertyCard(property));
            });
        }

        ui.grid.replaceChildren(fragment);

        ui.status.textContent = results.length
            ? `Showing ${start + 1} through ${end} of ${results.length} matching properties`
            : "No properties match your search or filters.";

        ui.status.classList.remove("is-visible");
    }

    let countryHydrated = false;

    async function load() {
        const request = ++latestRequest;
        const preserveInitialSSR = isCountry && !countryHydrated;

        state.currentPage = 1;
        ui.paginationWrap.hidden = true;
        ui.pagination.replaceChildren();

        ui.status.textContent = "Loading properties...";
        ui.status.classList.add("is-visible");

        // Keep the initial server-rendered cards visible until
        // the first successful client-side enhancement.
        if (!preserveInitialSSR) {
            ui.grid.replaceChildren();
        }

        // PostgreSQL Search() applies guest and date conditions.
        // Its destination search is keyword-based, so we enforce
        // exact country equality on the returned results below.
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
                `/api/properties?${params}`,
                { headers: { Accept: "application/json" } }
            );

            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.error || `HTTP ${response.status}`);
            }

            if (!Array.isArray(data.properties)) {
                throw new Error("Unexpected API response");
            }

            if (request !== latestRequest) return;

            state.properties = isCountry
                ? data.properties.filter(property =>
                    String(
                        property.location?.country || ""
                    ).toLowerCase() === country.toLowerCase()
                )
                : data.properties;

            countryHydrated = true;
            render();
        } catch (err) {
            if (request !== latestRequest) return;

            ui.status.textContent =
                `Unable to load properties: ${err.message}`;
            ui.status.classList.add("is-visible");
        }
    }

    function cloneApplied() {
        return { pax: state.pax, dateStart: state.dateStart, dateEnd: state.dateEnd,
            priceMin: state.priceMin, priceMax: state.priceMax,
            eco: state.eco, amenities: new Set(state.amenities) };
    }
    function amendGuest(delta) {
        const entered = Number(ui.guestNumber.value);
        const value = Number.isInteger(entered) && entered >= 1 && entered <= 20 ? entered : draft.pax;
        draft.pax = Math.max(1, Math.min(20, value + delta));
        ui.guestNumber.value = String(draft.pax);
    }
    function numericPrice(raw, bound, fallback) {
        if (String(raw).trim() === "") return fallback;
        const n = Number(raw);
        return Number.isFinite(n) ? Math.max(bound.low, Math.min(bound.high, Math.round(n))) : fallback;
    }
    function refreshPriceFields() {
        const bounds = priceBounds();
        const min = draft.priceMin === null ? bounds.low : Math.max(bounds.low, Math.min(bounds.high, draft.priceMin));
        const max = draft.priceMax === null ? bounds.high : Math.max(bounds.low, Math.min(bounds.high, draft.priceMax));
        const low = Math.min(min, max), high = Math.max(min, max);
        for (const slider of [ui.rangeMin, ui.rangeMax]) { slider.min = String(bounds.low); slider.max = String(bounds.high); slider.step = "1"; }
        ui.rangeMin.value = String(low); ui.rangeMax.value = String(high);
        ui.priceMin.value = String(low); ui.priceMax.value = String(high);
        ui.currencyMin.textContent = bounds.currency;
        ui.currencyMax.textContent = bounds.currency;
        const width = bounds.high - bounds.low;
        ui.rangeTrack.style.background = `linear-gradient(to right, #d9e2e6 0%, #d9e2e6 ${((low-bounds.low)/width)*100}%, #ff790b ${((low-bounds.low)/width)*100}%, #ff790b ${((high-bounds.low)/width)*100}%, #d9e2e6 ${((high-bounds.low)/width)*100}%)`;
    }
    function setPrice(kind, raw) {
        const bounds = priceBounds();
        const fallback = kind === "min" ? bounds.low : bounds.high;
        const value = numericPrice(raw, bounds, fallback);
        const other = kind === "min"
            ? (draft.priceMax ?? bounds.high) : (draft.priceMin ?? bounds.low);
        const adjusted = kind === "min" ? Math.min(value, other) : Math.max(value, other);
        if (kind === "min") draft.priceMin = adjusted <= bounds.low ? null : adjusted;
        else draft.priceMax = adjusted >= bounds.high ? null : adjusted;
        refreshPriceFields();
    }
    function availableAmenities() {
        return [...new Set(state.properties.flatMap(p => Array.isArray(p.amenities) ? p.amenities : []))]
            .filter(Boolean).sort((a,b) => a.localeCompare(b));
    }
    function populateAmenities() {
        const fragment = document.createDocumentFragment();
        for (const amenity of availableAmenities()) {
            const label = node("label", "refine-amenity");
            const check = node("input");
            check.type = "checkbox";
            check.checked = draft.amenities.has(amenity);
            check.addEventListener("change", () => {
                if (check.checked) draft.amenities.add(amenity);
                else draft.amenities.delete(amenity);
            });
            label.append(check, node("span", "", amenity));
            fragment.append(label);
        }
        ui.amenities.replaceChildren(fragment);
    }
    function paintDraft() {
        ui.eco.checked = draft.eco;
        ui.guestNumber.value = String(draft.pax);
        ui.modalDateSummary.textContent = draft.dateStart && draft.dateEnd
            ? `${shortDate(draft.dateStart)} – ${shortDate(draft.dateEnd)}` : "Select a date";
        refreshPriceFields();
        populateAmenities();
    }
    function openFilters(focus) {
        draft = cloneApplied();
        ui.guestSection.classList.toggle("is-highlighted", focus === "guests");
        ui.priceSection.classList.toggle("is-highlighted", focus === "price");
        paintDraft();
        ui.filter.showModal();
        if (focus === "guests") ui.guestNumber.focus();
        else if (focus === "price") ui.priceMax.focus();
        else ui.filterClose.focus();
    }
    function applyDraft() {
        const pax = Number(ui.guestNumber.value);
        if (!Number.isInteger(pax) || pax < 1 || pax > 20) {
            ui.guestNumber.focus(); ui.guestNumber.reportValidity(); return;
        }
        draft.pax = pax;
        const mustFetch = draft.pax !== state.pax || draft.dateStart !== state.dateStart || draft.dateEnd !== state.dateEnd;
        Object.assign(state, { pax: draft.pax, dateStart: draft.dateStart, dateEnd: draft.dateEnd,
            priceMin: draft.priceMin, priceMax: draft.priceMax, eco: draft.eco, amenities: new Set(draft.amenities) });
        ui.filter.close();

        if (mustFetch) {
            updateURL();
            load();
        } else {
            state.currentPage = 1;
            render();
        }
    }

    function isoDate(year, month, day) {
        return `${year}-${String(month+1).padStart(2,"0")}-${String(day).padStart(2,"0")}`;
    }
    function todayISO() { const d = new Date(); return isoDate(d.getFullYear(),d.getMonth(),d.getDate()); }
    function shortDate(value) {
        const d = new Date(`${value}T00:00:00Z`);
        return Number.isNaN(d.getTime()) ? value : d.toLocaleDateString("en-US", { month:"short", day:"2-digit", timeZone:"UTC" });
    }
    function calendarPanel(year, month) {
        const box = node("section", "refine-month");
        box.append(node("h3", "", new Intl.DateTimeFormat("en-US", { month: "long", year: "numeric" }).format(new Date(year,month,1))));
        const grid = node("div", "refine-month-days");
        for (const day of ["Su","Mo","Tu","We","Th","Fr","Sa"]) grid.append(node("span", "refine-day-name", day));
        for (let i=0; i<new Date(year,month,1).getDay(); i++) grid.append(node("span"));
        for (let day=1; day<=new Date(year,month+1,0).getDate(); day++) {
            const iso = isoDate(year,month,day);
            const btn = node("button", "refine-day", day);
            btn.type = "button";
            btn.disabled = iso < todayISO();
            btn.setAttribute("aria-label", new Date(year, month, day).toDateString());
            if (iso === chosenStart || iso === chosenEnd) btn.classList.add("selected");
            else if (chosenStart && chosenEnd && iso>chosenStart && iso<chosenEnd) btn.classList.add("between");
            btn.addEventListener("click", () => {
                if (!chosenStart || chosenEnd || iso <= chosenStart) { chosenStart = iso; chosenEnd = ""; }
                else chosenEnd = iso;
                paintCalendar();
            });
            grid.append(btn);
        }
        box.append(grid);
        return box;
    }
    function paintCalendar() {
        const year = visibleMonth.getFullYear(), month = visibleMonth.getMonth();
        const second = new Date(year, month+1, 1);
        ui.calendarMonths.replaceChildren(calendarPanel(year, month), calendarPanel(second.getFullYear(),second.getMonth()));
        const today = new Date();
        ui.monthPrev.disabled = year < today.getFullYear() || (year === today.getFullYear() && month <= today.getMonth());
        ui.calendarHint.textContent = chosenStart && chosenEnd
            ? `${shortDate(chosenStart)} – ${shortDate(chosenEnd)}`
            : chosenStart ? "Choose checkout" : "Choose check-in and checkout";
        ui.calendarApply.disabled = Boolean(chosenStart && !chosenEnd);
    }
    function openCalendar(fromFilters) {
        calendarOrigin = fromFilters ? "filters" : "standalone";
        chosenStart = fromFilters ? draft.dateStart : state.dateStart;
        chosenEnd = fromFilters ? draft.dateEnd : state.dateEnd;
        const date = chosenStart ? new Date(`${chosenStart}T12:00:00`) : new Date();
        visibleMonth = new Date(date.getFullYear(),date.getMonth(),1);
        paintCalendar();
        ui.calendar.showModal();
    }
    function acceptDates() {
        if (chosenStart && !chosenEnd) return;
        if (calendarOrigin === "filters") {
            draft.dateStart = chosenStart;
            draft.dateEnd = chosenEnd;
            paintDraft();
        } else {
            state.dateStart = chosenStart;
            state.dateEnd = chosenEnd;
            updateURL(); updatePills(); load();
        }
        ui.calendar.close();
    }

    // Filters are intentionally one modal, not four independent <details> panels.
    document.querySelectorAll("[data-filter-open]").forEach(button => {
        button.addEventListener("click", () => {
            const type = button.dataset.filterOpen;
            if (type === "dates") openCalendar(false);
            else openFilters(type);
        });
    });

    document.querySelectorAll("[data-filter-remove]").forEach(button => {
        button.addEventListener("click", event => {
            event.stopPropagation();

            const type = button.dataset.filterRemove;
            let needsFetch = false;

            switch (type) {
                case "dates":
                    state.dateStart = "";
                    state.dateEnd = "";
                    needsFetch = true;
                    break;

                case "price":
                    state.priceMin = null;
                    state.priceMax = null;
                    break;

                case "guests":
                    state.pax = 2;
                    needsFetch = true;
                    break;

                case "more":
                    state.eco = false;
                    state.amenities.clear();
                    break;

                default:
                    return;
            }

            if (needsFetch) {
                updateURL();
                updatePills();
                load();
            } else {
                state.currentPage = 1;
                render();
            }

            button.parentElement
                .querySelector("[data-filter-open]")
                .focus();
        });
    });

    ui.filterClose.addEventListener("click", () => ui.filter.close());
    ui.filterClear.addEventListener("click", () => { draft = { pax:2, dateStart:"", dateEnd:"", priceMin:null,priceMax:null,eco:false,amenities:new Set() }; paintDraft(); });
    ui.filterSearch.addEventListener("click", applyDraft);
    ui.modalDates.addEventListener("click", () => openCalendar(true));
    ui.guestNumber.addEventListener("input", () => {
        const n = Number(ui.guestNumber.value);
        if (Number.isInteger(n) && n >= 1 && n <= 20) draft.pax = n;
    });
    ui.guestMinus.addEventListener("click", () => amendGuest(-1));
    ui.guestPlus.addEventListener("click", () => amendGuest(1));
    ui.rangeMin.addEventListener("input", () => setPrice("min", ui.rangeMin.value));
    ui.rangeMax.addEventListener("input", () => setPrice("max", ui.rangeMax.value));
    ui.priceMin.addEventListener("change", () => setPrice("min", ui.priceMin.value));
    ui.priceMax.addEventListener("change", () => setPrice("max", ui.priceMax.value));
    ui.eco.addEventListener("change", () => { draft.eco = ui.eco.checked; });
    ui.calendarClose.addEventListener("click", () => ui.calendar.close());
    ui.monthPrev.addEventListener("click", () => { visibleMonth.setMonth(visibleMonth.getMonth()-1); paintCalendar(); });
    ui.monthNext.addEventListener("click", () => { visibleMonth.setMonth(visibleMonth.getMonth()+1); paintCalendar(); });
    ui.calendarClear.addEventListener("click", () => { chosenStart=""; chosenEnd=""; paintCalendar(); });
    ui.calendarApply.addEventListener("click", acceptDates);
    ui.messageClose.addEventListener("click", () => ui.message.close());
    ui.sort.addEventListener("change", () => {
        state.currentPage = 1;
        render();
    });
    window.addEventListener("popstate", () => { parseURL(); load(); });
    
    function alignWithPublisherHeader() {
        const navigation = document.getElementById("main-navigation");
        const inner = app.querySelector(".refine-inner");
    
        // Other publisher sites might use different header markup.
        // CSS provides a fallback when this navigation is absent.
        if (!navigation || !inner) return;
    
        const navRect = navigation.getBoundingClientRect();
        const pageRect = app.getBoundingClientRect();
    
        if (navRect.width <= 0 || pageRect.width <= 0) return;
    
        const left = Math.max(0, navRect.left - pageRect.left);
        const right = Math.max(0, pageRect.right - navRect.right);
    
        if (left + right >= pageRect.width) return;
    
        inner.style.width = "auto";
        inner.style.maxWidth = "none";
        inner.style.marginLeft = `${left}px`;
        inner.style.marginRight = `${right}px`;
        inner.style.paddingInline = "0";
    }
    
    window.addEventListener("resize", alignWithPublisherHeader);
    
    alignWithPublisherHeader();
    parseURL();
    load();
})();

