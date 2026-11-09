
(() => {
    "use strict";

    const popup = document.querySelector(
        ".property-search-popup-container"
    );

    if (!popup ||
        popup.dataset.sparrowSearchReady === "true") {
        return;
    }

    const destination = popup.querySelector("#location_search");
    const dates = popup.querySelector("#date_picker");
    const guestDisplay = popup.querySelector("#guest_picker");
    const guestPanel = popup.querySelector("#search-guest-panel");
    const searchButton = popup.querySelector(".search-property-btn");

    const originalGuestCount =
        guestPanel?.querySelector(".guest-count");

    if (!destination || !dates || !guestDisplay ||
        !guestPanel || !originalGuestCount || !searchButton) {
        return;
    }

    const dateWrap = dates.closest(".datepicker-wrap");
    const guestWrap = guestDisplay.closest(".guest-container");

    if (!dateWrap || !guestWrap) return;

    popup.dataset.sparrowSearchReady = "true";
    popup.classList.add("sparrow-search-ready");

    let selectedStart = "";
    let selectedEnd = "";

    // The original readonly field displays the chosen
    // dates. The new panel contains actual date inputs.
    dates.setAttribute("aria-haspopup", "dialog");
    dates.setAttribute("aria-expanded", "false");
    dates.setAttribute("role", "button");
    dates.tabIndex = 0;

    const datePanel = document.createElement("div");
    datePanel.className = "sparrow-date-panel";
    datePanel.hidden = true;

    datePanel.innerHTML = `
        <label>
            Check-in
            <input type="date"
                   class="sparrow-date-start"
                   aria-label="Check-in date">
        </label>

        <label>
            Checkout
            <input type="date"
                   class="sparrow-date-end"
                   aria-label="Checkout date">
        </label>

        <div class="sparrow-date-actions">
            <button type="button"
                    class="sparrow-date-clear">
                Clear
            </button>

            <button type="button"
                    class="sparrow-date-apply">
                Apply
            </button>
        </div>
    `;

    dateWrap.append(datePanel);

    const startInput = datePanel.querySelector(
        ".sparrow-date-start"
    );
    const endInput = datePanel.querySelector(
        ".sparrow-date-end"
    );
    const applyDates = datePanel.querySelector(
        ".sparrow-date-apply"
    );
    const clearDates = datePanel.querySelector(
        ".sparrow-date-clear"
    );

    // Retain the original count span for compatibility
    // with publisher code, but show an editable input.
    const guestNumber = document.createElement("input");

    guestNumber.type = "number";
    guestNumber.className = "sparrow-guest-number";
    guestNumber.min = "1";
    guestNumber.max = "20";
    guestNumber.step = "1";
    guestNumber.value = "2";
    guestNumber.setAttribute(
        "aria-label",
        "Number of guests"
    );

    originalGuestCount.hidden = true;
    originalGuestCount.insertAdjacentElement(
        "afterend",
        guestNumber
    );

    const decrease = guestPanel.querySelector(".decrease-nav");
    const increase = guestPanel.querySelector(".increase-nav");
    const guestClear = guestPanel.querySelector(".guest-clear-btn");
    const guestSave = guestPanel.querySelector(".guest-save-btn");

    if (!decrease || !increase || !guestClear || !guestSave) {
        return;
    }

    // The original publisher CSS draws the +/- symbols.
    // We do not add duplicate symbols using textContent.

    // Keep validation feedback outside the horizontal
    // search row to prevent it becoming another flex item.
    const error = document.createElement("p");
    error.className = "sparrow-search-error";
    error.setAttribute("role", "alert");
    error.hidden = true;

    popup.querySelector(".property-search-content")?.insertAdjacentElement(
        "afterend", error
    );

    function setError(message) {
        error.textContent = message;
        error.hidden = !message;
    }

    function closeDates() {
        datePanel.hidden = true;
        dates.setAttribute("aria-expanded", "false");
    }

    function closeGuests() {
        guestPanel.classList.remove("sparrow-open");
        guestDisplay.setAttribute("aria-expanded", "false");
    }

    function guestCount() {
        const count = Number(guestNumber.value);

        if (!Number.isInteger(count) ||
            count < 1 || count > 20) {
            return null;
        }

        return count;
    }

    function setGuests(count) {
        const value = Math.max(1, Math.min(20, count));

        guestNumber.value = String(value);

        guestDisplay.value =
            `${value} ${value === 1 ? "Person" : "Persons"}`;

        setError("");
    }

    function formatDate(value) {
        const date = new Date(`${value}T00:00:00Z`);

        return date.toLocaleDateString("en-US", {
            month: "short",
            day: "2-digit",
            timeZone: "UTC"
        });
    }

    function validateDates(start, end) {
        if ((start && !end) || (!start && end)) {
            return "Choose both check-in and checkout dates.";
        }

        if (start && end <= start) {
            return "Checkout must be after check-in.";
        }

        return "";
    }

    // Open the date panel using the original date field.
    dates.addEventListener("click", (event) => {
        event.preventDefault();
        event.stopImmediatePropagation();

        const shouldOpen = datePanel.hidden;

        closeGuests();
        datePanel.hidden = !shouldOpen;

        dates.setAttribute(
            "aria-expanded",
            String(shouldOpen)
        );

        if (shouldOpen) startInput.focus();
    }, true);

    dates.addEventListener("keydown", (event) => {
        if (event.key === "Enter" || event.key === " ") {
            event.preventDefault();
            dates.click();
        }
    });

    // Open the existing guest panel.
    guestDisplay.addEventListener("click", (event) => {
        event.preventDefault();
        event.stopImmediatePropagation();

        const shouldOpen =
            !guestPanel.classList.contains("sparrow-open");

        closeDates();

        guestPanel.classList.toggle(
            "sparrow-open",
            shouldOpen
        );

        guestDisplay.setAttribute(
            "aria-expanded",
            String(shouldOpen)
        );

        if (shouldOpen) guestNumber.focus();
    }, true);

    guestDisplay.addEventListener("keydown", (event) => {
        if (event.key === "Enter" || event.key === " ") {
            event.preventDefault();
            guestDisplay.click();
        }
    });

    // Commit selected dates.
    applyDates.addEventListener("click", () => {
        const start = startInput.value;
        const end = endInput.value;

        const message = validateDates(start, end);

        if (message) {
            setError(message);
            return;
        }

        selectedStart = start;
        selectedEnd = end;

        dates.value = start && end
            ? `${formatDate(start)} – ${formatDate(end)}`
            : "";

        setError("");
        closeDates();
    });

    // Remove the date range.
    clearDates.addEventListener("click", () => {
        selectedStart = "";
        selectedEnd = "";

        startInput.value = "";
        endInput.value = "";
        dates.value = "";

        setError("");
        closeDates();
    });

    // Connect the original +/- buttons to the number input.
    for (const [button, amount] of [
        [decrease, -1],
        [increase, 1]
    ]) {
        button.addEventListener("click", (event) => {
            event.preventDefault();
            event.stopImmediatePropagation();

            setGuests((guestCount() || 2) + amount);
        }, true);
    }

    guestNumber.addEventListener("change", () => {
        const value = guestCount();

        if (value === null) {
            setError(
                "Guests must be a whole number between 1 and 20."
            );
            return;
        }

        setGuests(value);
    });

    guestClear.addEventListener("click", (event) => {
        event.preventDefault();
        event.stopImmediatePropagation();

        // Restore the original default of two guests.
        setGuests(2);
    }, true);

    guestSave.addEventListener("click", (event) => {
        event.preventDefault();
        event.stopImmediatePropagation();

        const value = guestCount();

        if (value === null) {
            setError(
                "Guests must be a whole number between 1 and 20."
            );
            return;
        }

        setGuests(value);
        closeGuests();
    }, true);

    // Convert selections into the exact URL format
    // expected by the existing /refine page.
    function submitSearch() {
        const count = guestCount();

        if (count === null) {
            setError(
                "Guests must be a whole number between 1 and 20."
            );
            return;
        }

        const dateError = validateDates(
            selectedStart,
            selectedEnd
        );

        if (dateError) {
            setError(dateError);
            return;
        }

        const search = destination.value.trim();

        if (!search) {
            setError("Enter a destination to search.");
            destination.focus();
            return;
        }

        const params = new URLSearchParams({
            search: search.toLowerCase(),
            pax: String(count),
            order: "1"
        });

        if (selectedStart && selectedEnd) {
            params.set("dateStart", selectedStart);
            params.set("dateEnd", selectedEnd);
        }

        // Relative URL: preserve the current host and port.
        window.location.assign(
            `/refine?${params.toString()}`
        );
    }

    searchButton.addEventListener("click", (event) => {
        event.preventDefault();
        event.stopImmediatePropagation();

        submitSearch();
    }, true);

    // Close pickers when clicking elsewhere.
    document.addEventListener("click", (event) => {
        if (!dateWrap.contains(event.target)) {
            closeDates();
        }

        if (!guestWrap.contains(event.target)) {
            closeGuests();
        }
    });

    document.addEventListener("keydown", (event) => {
        if (event.key === "Escape") {
            closeDates();
            closeGuests();
        }
    });

    setGuests(2);
})();
