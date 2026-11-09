
<!DOCTYPE html>
<html lang="{{.Language}}">
<head>
    <meta charset="UTF-8">
    <meta name="viewport"
          content="width=device-width, initial-scale=1.0">

    <title>Find a Place to Stay | {{.SiteName}}</title>
    <link rel="canonical" href="{{.CanonicalURL}}">

    {{if .GlobalCSS}}
    <link rel="stylesheet" href="{{.GlobalCSS}}">
    {{end}}

    {{if .TemplateCSS}}
    <link rel="stylesheet" href="{{.TemplateCSS}}">
    {{end}}

    <link rel="stylesheet" href="/static/css/refine.css">

    <link rel="stylesheet" href="/static/css/property_search_popup.css">

    {{.WidgetHTML}}
</head>

<body class="refine-route">
    {{.HeaderHTML}}

    <main id="refine-app" class="refine-app">
        <div class="refine-inner">

            <p class="refine-breadcrumb">
                Vacation Rentals in
                <span id="refine-breadcrumb-destination">
                    your destination
                </span>
            </p>

            <h1>
                Find a Place to Stay in
                <span id="refine-heading-destination">
                    your destination
                </span>
            </h1>

            <p class="refine-subtitle">
                Plan Your Perfect Getaway
            </p>

            <form id="refine-search-form"
                  class="refine-search-form"
                  role="search">

                <label for="refine-destination">
                    Destination
                    <input id="refine-destination"
                           type="search"
                           placeholder="Search destination"
                           autocomplete="off">
                </label>

                <label for="refine-guests">
                    Guests
                    <input id="refine-guests"
                           type="number"
                           min="1"
                           max="20"
                           step="1"
                           value="2"
                           required>
                </label>

                <button type="submit"
                        class="refine-search-submit">
                    Search Stays
                </button>
            </form>

            <div class="refine-toolbar">
                <div class="refine-filters">
                    <details id="refine-date-dropdown"
                             class="refine-date-dropdown">

                        <summary id="refine-dates"
                                 class="refine-date-button">
                            Dates
                        </summary>

                        <div class="refine-date-panel">
                            <label for="refine-date-start">
                                Check-in
                                <input id="refine-date-start"
                                       type="date">
                            </label>

                            <label for="refine-date-end">
                                Checkout
                                <input id="refine-date-end"
                                       type="date">
                            </label>

                            <div class="refine-date-actions">
                                <button id="refine-date-clear"
                                        type="button">
                                    Clear
                                </button>

                                <button id="refine-date-apply"
                                        type="button">
                                    Apply Dates
                                </button>
                            </div>
                        </div>
                    </details>

                    <label for="refine-max-price">
                        Maximum Price (৳)
                        <input id="refine-max-price"
                               type="number"
                               min="0"
                               placeholder="Any price">
                    </label>

                    <label class="refine-filter-checkbox"
                           for="refine-pool-only">
                        <input id="refine-pool-only"
                               type="checkbox">
                        Pool only
                    </label>
                </div>

                <label for="refine-sort"
                       class="refine-sort-label">
                    Sort by
                    <select id="refine-sort">
                        <option value="popular">
                            Most Reviewed
                        </option>
                        <option value="price_low">
                            Price: Low to High
                        </option>
                        <option value="price_high">
                            Price: High to Low
                        </option>
                    </select>
                </label>
            </div>

            <p id="refine-status"
               class="refine-status"
               role="status"
               aria-live="polite">
                Loading properties...
            </p>

            <!-- JavaScript creates the property cards here. -->
            <section id="refine-grid"
                     class="refine-grid"
                     aria-label="Property listings">
            </section>

        </div>
    </main>

    {{.FooterHTML}}

    <dialog id="refine-dialog"
            class="refine-dialog"
            aria-labelledby="refine-dialog-title">
        <h2 id="refine-dialog-title"></h2>
        <p id="refine-dialog-message"></p>
        <button id="refine-dialog-close" type="button">
            Close
        </button>
    </dialog>

    {{if .GlobalJS}}
    <script src="{{.GlobalJS}}" defer></script>
    {{end}}

    <script src="/static/js/refine.js" defer></script>

    <script src="/static/js/property_search_popup.js" defer></script>
</body>
</html>
