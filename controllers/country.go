package controllers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"sparrow-ssr/services"

	beego "github.com/beego/beego/v2/server/web"
)

var countrySearcher services.PropertySearcher

func SetCountrySearchDependencies(
	search services.PropertySearcher,
) {
	countrySearcher = search
}

type CountryController struct {
	beego.Controller
}

type countryCard struct {
	Name       string
	Type       string
	Photo      string
	Price      string
	Location   string
	Amenities  string
	Provider   string
	BookingURL string
	Rating     float64
	Reviews    int
	Eco        bool
}

var countrySlugPattern = regexp.MustCompile(
	`^[a-z]+(?:-[a-z]+)*$`,
)

func countryFromSlug(slug string) (string, bool) {
	if len(slug) < 2 ||
		len(slug) > 80 ||
		!countrySlugPattern.MatchString(slug) {
		return "", false
	}

	if slug == "usa" {
		return "USA", true
	}

	words := strings.Split(slug, "-")
	for i, word := range words {
		words[i] = strings.ToUpper(word[:1]) + word[1:]
	}

	return strings.Join(words, " "), true
}

func safePhoto(images []string) string {
	if len(images) == 0 {
		return ""
	}

	u, err := url.Parse(images[0])
	if err != nil ||
		(u.Scheme != "https" && u.Scheme != "http") ||
		u.Host == "" ||
		u.User != nil {
		return ""
	}

	return u.String()
}

// Match the provider restrictions from static/js/refine.js.
func safeBooking(provider, raw string) string {
	provider = strings.ToLower(strings.TrimSpace(provider))

	domain := ""

	switch provider {
	case "booking.com":
		domain = "booking.com"
	case "expedia", "expedia.com":
		domain = "expedia.com"
	default:
		return ""
	}

	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil ||
		u.Scheme != "https" ||
		u.User != nil ||
		u.Port() != "" ||
		u.Path == "" ||
		u.Path == "/" {
		return ""
	}

	host := strings.ToLower(u.Hostname())

	if host != domain &&
		!strings.HasSuffix(host, "."+domain) {
		return ""
	}

	return u.String()
}

func (c *CountryController) Get() {
	if dependencies.catalog == nil ||
		dependencies.loader == nil ||
		countrySearcher == nil {
		c.writeCountryError(
			http.StatusInternalServerError,
			"Application is not initialized",
		)
		return
	}

	location, found := dependencies.catalog.Resolve(
		c.Ctx.Request.Host,
		"/",
	)
	if !found {
		c.writeCountryError(
			http.StatusNotFound,
			"Site not found",
		)
		return
	}

	slug := c.Ctx.Input.Param(":country")
	country, valid := countryFromSlug(slug)
	if !valid {
		c.writeCountryError(
			http.StatusNotFound,
			"Country not found",
		)
		return
	}

	properties, err := countrySearcher.SearchCountry(
		location.ContentDirectory, country,
	)
	if err != nil {
		log.Printf("Country property search failed: %v", err)
		c.writeCountryError(
			http.StatusInternalServerError,
			"Failed to search country properties",
		)
		return
	}

	content, err := dependencies.loader.Load(location)
	if err != nil {
		log.Printf(
			"Country page content loading failed: %v", err,
		)
		c.writeCountryError(
			http.StatusInternalServerError,
			"Failed to load site content",
		)
		return
	}

	cards := make([]countryCard, 0, len(properties))

	for _, p := range properties {
		detailParts := make(
			[]string, 0, len(p.Amenities)+1,
		)

		if p.SizeSqFt > 0 {
			detailParts = append(
				detailParts,
				fmt.Sprintf("%d ft²", p.SizeSqFt),
			)
		}

		detailParts = append(detailParts, p.Amenities...)

		price := "Check provider for price"
		if p.PriceFrom > 0 {
			price = fmt.Sprintf(
				"From %s %d", p.Currency, p.PriceFrom,
			)
		}

		cards = append(cards, countryCard{
			Name:  p.Name,
			Type:  p.PropertyType,
			Photo: safePhoto(p.Images),
			Price: price,
			Location: strings.Join(
				nonempty(
					p.Location.City,
					p.Location.Area,
					p.Location.State,
				),
				" › ",
			),
			Amenities: strings.Join(
				detailParts, " · ",
			),
			Provider: p.ProviderName,
			BookingURL: safeBooking(
				p.ProviderName, p.BookingURL,
			),
			Rating:  p.Rating,
			Reviews: p.ReviewCount,
			Eco:     p.EcoFriendly,
		})
	}

	language := content.Header.Language
	if language == "" {
		language = "en"
	}

	c.Data["Language"] = language
	c.Data["SiteName"] = content.Header.SiteName
	c.Data["Country"] = country
	c.Data["Total"] = len(cards)
	c.Data["Cards"] = cards

	c.Data["CanonicalURL"] =
		"https://" + location.Domain + "/all/" + slug

	c.Data["RefineURL"] =
		"/refine?search=" +
			url.QueryEscape(country) + "&pax=2&order=1"

	c.Data["GlobalCSS"] = content.Header.CSSFile
	c.Data["TemplateCSS"] =
		content.Header.TemplateCSSFile
	c.Data["GlobalJS"] = content.Header.JSFile

	// Trusted publisher-provided fragments only.
	c.Data["HeaderHTML"] = template.HTML(
		content.Header.HeaderHTML,
	)
	c.Data["FooterHTML"] = template.HTML(
		content.Footer.FooterHTML,
	)
	c.Data["WidgetHTML"] = template.HTML(
		content.Header.WidgetCode,
	)

	c.TplName = "all.tpl"

	if err := c.Render(); err != nil {
		log.Printf("Country page rendering failed: %v", err)
		c.writeCountryError(
			http.StatusInternalServerError,
			"Failed to render country page",
		)
		return
	}

	c.EnableRender = false
}

func nonempty(values ...string) []string {
	result := make([]string, 0, len(values))

	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			result = append(result, value)
		}
	}

	return result
}

func (c *CountryController) writeCountryError(
	status int,
	message string,
) {
	c.EnableRender = false
	c.Ctx.Output.SetStatus(status)
	_ = c.Ctx.Output.Body([]byte(message))
}
