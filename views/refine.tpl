
<!DOCTYPE html>
<html lang="{{.Language}}">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Find a Place to Stay | {{.SiteName}}</title>
    <link rel="canonical" href="{{.CanonicalURL}}">
    {{if .GlobalCSS}}<link rel="stylesheet" href="{{.GlobalCSS}}">{{end}}
    {{if .TemplateCSS}}<link rel="stylesheet" href="{{.TemplateCSS}}">{{end}}
    <link rel="stylesheet" href="/static/css/refine.css">
    {{if .IsCountry}}
        <link rel="stylesheet" href="/static/css/all.css">
    {{end}}
    <link rel="stylesheet" href="/static/css/property_search_popup.css">
    {{.WidgetHTML}}
</head>
<body class="refine-route{{if .IsCountry}} country-route{{end}}">
    {{.HeaderHTML}}


    <main id="{{if .IsCountry}}country-results{{else}}refine-app{{end}}"
          class="refine-app"{{if .IsCountry}} data-country="{{.Country}}"{{end}}>
        <div class="refine-inner">
            {{if .IsCountry}}
            <p class="refine-breadcrumb">
                {{.Total}} Vacation Rentals Near {{.Country}} | {{.Country}}
            </p>
            <h1>{{.Country}} Resorts, Hotels &amp; Rentals</h1>
            <p class="refine-subtitle">
                Plan your next getaway to {{.Country}}
            </p>
            {{else}}
            <p class="refine-breadcrumb">
                Vacation Rentals in
                <span id="refine-breadcrumb-destination">your destination</span>
            </p>
            <h1>
                Find a Place to Stay in
                <span id="refine-heading-destination">your destination</span>
            </h1>
            <p class="refine-subtitle">Plan Your Perfect Getaway to</p>
            {{end}}

            <div class="refine-toolbar">
                <div class="refine-filters" role="group" aria-label="Property filters">
                    <div class="refine-filter-chip" data-filter-chip="dates">
                        <button type="button" class="refine-filter-pill"
                                data-filter-open="dates">Dates</button>
                        <button type="button" class="refine-filter-remove"
                                data-filter-remove="dates"
                                aria-label="Clear dates filter"
                                hidden>×</button>
                    </div>

                    <div class="refine-filter-chip" data-filter-chip="price">
                        <button type="button" class="refine-filter-pill"
                                data-filter-open="price">Price</button>
                        <button type="button" class="refine-filter-remove"
                                data-filter-remove="price"
                                aria-label="Clear price filter"
                                hidden>×</button>
                    </div>

                    <div class="refine-filter-chip" data-filter-chip="guests">
                        <button type="button" class="refine-filter-pill"
                                data-filter-open="guests">Guests</button>
                        <button type="button" class="refine-filter-remove"
                                data-filter-remove="guests"
                                aria-label="Reset guests to two"
                                hidden>×</button>
                    </div>

                    <div class="refine-filter-chip" data-filter-chip="more">
                        <button type="button" class="refine-filter-pill"
                                data-filter-open="more">More</button>
                        <span id="refine-more-count"
                              class="refine-filter-count"
                              aria-hidden="true"
                              hidden></span>
                    </div>
                </div>
                <label class="refine-sort-label" for="refine-sort">
                    Sort by
                    <select id="refine-sort">
                        <option value="popular">Most Popular</option>
                        <option value="price_low">Price: Low to High</option>
                        <option value="price_high">Price: High to Low</option>
                    </select>
                </label>
            </div>

            <p id="refine-status" class="refine-status" role="status" aria-live="polite">Loading properties...</p>

            <section id="refine-grid"
                     class="refine-grid"
                     aria-label="{{if .IsCountry}}{{.Country}} property listings{{else}}Property listings{{end}}"
                     tabindex="-1">

                {{if .IsCountry}}
                {{range .Cards}}
                <article class="refine-card">
                    <div class="refine-card-picture">
                        {{if .Photo}}
                        <img class="refine-card-image"
                             src="{{.Photo}}"
                             loading="lazy"
                             alt="Illustrative accommodation photograph for {{.Name}}">
                        {{else}}
                        <div class="refine-image-fallback">
                            Photo unavailable
                        </div>
                        {{end}}

                        <span class="refine-price">{{.Price}}</span>

                        {{if .Eco}}
                        <span class="refine-eco">Eco</span>
                        {{end}}
                    </div>

                    <div class="refine-card-body">
                        <div class="refine-card-meta">
                            <span class="refine-rating">
                                <span class="refine-rating-icon"
                                      aria-hidden="true">●</span>
                                {{printf "%.1f" .Rating}}
                                ({{.Reviews}} Reviews)
                            </span>

                            <span class="refine-property-type">
                                {{.Type}}
                            </span>
                        </div>

                        <h2 class="refine-card-title">
                            {{.Name}}
                        </h2>

                        <p class="refine-card-details">
                            {{.Amenities}}
                        </p>

                        <p class="refine-card-location">
                            {{.Location}}
                        </p>

                        <div class="refine-card-bottom">
                            <span class="refine-provider">
                                {{.Provider}}
                            </span>

                            {{if .BookingURL}}
                            <a class="refine-availability country-availability"
                               href="{{.BookingURL}}"
                               rel="noopener noreferrer">
                                View Availability
                            </a>
                            {{else}}
                            <span class="country-unavailable">
                                Provider link unavailable
                            </span>
                            {{end}}
                        </div>
                    </div>
                </article>

                {{else}}
                <p class="refine-empty">
                    No properties are currently listed in {{$.Country}}.
                </p>
                {{end}}
                {{end}}
            </section>

            <div id="refine-pagination-wrap"
                 class="refine-pagination-wrap"
                 hidden>
                <p id="refine-page-summary"
                   class="refine-page-summary"></p>

                <nav id="refine-pagination"
                     class="refine-pagination"
                     aria-label="Property result pages"></nav>
            </div>

            <p class="refine-disclaimer">
                Sample prices, guest capacities, photographs and availability are
                illustrative. Confirm details with the booking provider.
            </p>
        </div>
    </main>

    {{.FooterHTML}}

    <!-- All of Price, Guests and More open this SAME centered filter modal. -->
    <dialog id="refine-filter-dialog" class="refine-filter-dialog" aria-labelledby="refine-filter-title">
        <div class="refine-modal-heading">
            <h2 id="refine-filter-title">Filters</h2>
            <button type="button" id="refine-filter-close" class="refine-modal-x" aria-label="Close filters">×</button>
        </div>
        <div class="refine-filter-modal-body">
            <div class="refine-modal-two-cols">
                <label class="refine-modal-tile refine-modal-check disabled" title="Pet-friendly data is not available in this assessment dataset">
                    <span>🐾 Pet-friendly only</span>
                    <input type="checkbox" disabled aria-label="Pet-friendly filter unavailable">
                </label>
                <label class="refine-modal-tile refine-modal-check">
                    <span>🍃 Eco-friendly only</span>
                    <input id="refine-eco-only" type="checkbox">
                </label>

                <button type="button"
                        id="refine-modal-dates"
                        class="refine-modal-tile refine-modal-action">
                    <span id="refine-modal-date-summary">Select a date</span>

                    <span class="refine-modal-calendar-icon" aria-hidden="true">
                        <svg xmlns="http://www.w3.org/2000/svg"
                            viewBox="0 0 24 24"
                            width="19"
                            height="19"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                            stroke-linecap="round"
                            stroke-linejoin="round">
                            <rect x="3" y="5" width="18" height="16" rx="2"/>
                            <path d="M7 3v4M17 3v4M3 10h18"/>
                            <path d="m9 16 2 2 4-4"/>
                        </svg>
                    </span>
                </button>
                <div id="refine-guest-section" class="refine-modal-tile refine-modal-guests">
                    <span>Guests</span>
                    <div class="refine-guest-controls">
                        <button id="refine-guest-minus" type="button" aria-label="Decrease guests">−</button>
                        <input id="refine-guest-number" type="number" min="1" max="20" step="1" value="2" aria-label="Number of guests">
                        <button id="refine-guest-plus" type="button" aria-label="Increase guests">+</button>
                    </div>
                </div>
            </div>

            <section id="refine-price-section" class="refine-price-section" aria-labelledby="refine-price-heading">
                <h3 id="refine-price-heading">Price range</h3>
                <div class="refine-range-wrap">
                    <div class="refine-range-track" id="refine-range-track"></div>
                    <input id="refine-range-min" type="range" min="0" max="100" value="0" aria-label="Minimum price slider">
                    <input id="refine-range-max" type="range" min="0" max="100" value="100" aria-label="Maximum price slider">
                </div>
                <div class="refine-price-fields">
                    <label for="refine-min-price">Min price</label>
                    <span class="refine-money-field"><span id="refine-price-currency-min">USD</span><input id="refine-min-price" type="number" min="0" step="1" inputmode="numeric" aria-label="Minimum price"></span>
                    <span class="refine-price-dash">–</span>
                    <span class="refine-money-field"><span id="refine-price-currency-max">USD</span><input id="refine-max-price" type="number" min="0" step="1" inputmode="numeric" aria-label="Maximum price"></span>
                    <label for="refine-max-price">Max price</label>
                </div>
            </section>

            <section class="refine-amenities-section" aria-labelledby="refine-amenities-heading">
                <h3 id="refine-amenities-heading">Amenities</h3>
                <div id="refine-amenities-grid" class="refine-amenities-grid"></div>
            </section>
        </div>
        <div class="refine-modal-footer">
            <button type="button" id="refine-filter-clear" class="refine-clear-button">Clear</button>
            <button type="button" id="refine-filter-search" class="refine-search-button">Search</button>
        </div>
    </dialog>

    <!-- Date selection uses an independent modal, not an inline details panel. -->
    <dialog id="refine-calendar-dialog" class="refine-calendar-dialog" aria-labelledby="refine-calendar-title">
        <div class="refine-modal-heading">
            <h2 id="refine-calendar-title">Select your dates</h2>
            <button type="button" id="refine-calendar-close" class="refine-modal-x" aria-label="Close calendar">×</button>
        </div>
        <div class="refine-calendar-content">
            <div class="refine-calendar-toolbar">
                <button id="refine-month-prev" type="button" aria-label="Previous month">‹</button>
                <strong id="refine-calendar-hint">Choose check-in and checkout</strong>
                <button id="refine-month-next" type="button" aria-label="Next month">›</button>
            </div>
            <div id="refine-calendar-months" class="refine-calendar-months"></div>
        </div>
        <div class="refine-modal-footer">
            <button type="button" id="refine-calendar-clear" class="refine-clear-button">Clear</button>
            <button type="button" id="refine-calendar-apply" class="refine-search-button">Apply dates</button>
        </div>
    </dialog>

    <dialog id="refine-dialog" class="refine-message-dialog" aria-labelledby="refine-dialog-title">
        <h2 id="refine-dialog-title"></h2>
        <p id="refine-dialog-message"></p>
        <button type="button" id="refine-dialog-close">Close</button>
    </dialog>

    {{if .GlobalJS}}<script src="{{.GlobalJS}}" defer></script>{{end}}
    <script src="/static/js/refine.js" defer></script>
    <script src="/static/js/property_search_popup.js" defer></script>
</body>
</html>
