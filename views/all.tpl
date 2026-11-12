
<!DOCTYPE html>
<html lang="{{.Language}}">
<head>
    <meta charset="UTF-8">
    <meta name="viewport"
          content="width=device-width, initial-scale=1.0">

    <title>
        {{.Country}} Resorts, Hotels &amp; Rentals | {{.SiteName}}
    </title>

    <link rel="canonical" href="{{.CanonicalURL}}">

    {{if .GlobalCSS}}
    <link rel="stylesheet" href="{{.GlobalCSS}}">
    {{end}}

    {{if .TemplateCSS}}
    <link rel="stylesheet" href="{{.TemplateCSS}}">
    {{end}}

    <link rel="stylesheet" href="/static/css/refine.css">
    <link rel="stylesheet" href="/static/css/all.css">
    <link rel="stylesheet"
          href="/static/css/property_search_popup.css">

    {{.WidgetHTML}}
</head>

<body class="refine-route country-route">
    {{.HeaderHTML}}

    <main id="country-results" class="refine-app">
        <div class="refine-inner">

            <p class="refine-breadcrumb">
                {{.Total}} Vacation Rentals Near
                {{.Country}} | {{.Country}}
            </p>

            <h1>
                {{.Country}} Resorts, Hotels &amp; Rentals
            </h1>

            <p class="refine-subtitle">
                Plan your next getaway to {{.Country}}
            </p>

            <div class="refine-toolbar">
                <a class="refine-filter-pill country-refine-link"
                   href="{{.RefineURL}}">
                    Dates · Price · Guests · More filters
                </a>

                <span class="country-count">
                    {{.Total}} properties
                </span>
            </div>

            <section class="refine-grid"
                     aria-label="{{.Country}} property listings">

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

                        <span class="refine-price">
                            {{.Price}}
                        </span>

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
                    No properties are currently listed in
                    {{$.Country}}.
                </p>
                {{end}}

            </section>

            <p class="refine-disclaimer">
                Sample prices, guest capacities, photographs and
                availability are illustrative. Confirm details
                with the booking provider.
            </p>
        </div>
    </main>

    {{.FooterHTML}}

    {{if .GlobalJS}}
    <script src="{{.GlobalJS}}" defer></script>
    {{end}}

    <script src="/static/js/property_search_popup.js"
            defer></script>
</body>
</html>
