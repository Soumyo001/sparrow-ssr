package controllers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"sparrow-ssr/models"
	"sparrow-ssr/services"

	beego "github.com/beego/beego/v2/server/web"
)

// Property API dependencies are initialized before
// the application starts accepting requests.
type propertyAPIDependencies struct {
	catalog *services.Catalog
	search  *services.PropertySearchService
}

var propertyAPI propertyAPIDependencies

func SetPropertyAPIDependencies(
	catalog *services.Catalog,
	search *services.PropertySearchService,
) {
	propertyAPI = propertyAPIDependencies{
		catalog: catalog,
		search:  search,
	}
}

type PropertyAPIController struct {
	beego.Controller
}

// propertySearchResponse defines the JSON API response.
type propertySearchResponse struct {
	Search     string            `json:"search"`
	Pax        int               `json:"pax"`
	Order      int               `json:"order"`
	Total      int               `json:"total"`
	DateStart  string            `json:"dateStart,omitempty"`
	DateEnd    string            `json:"dateEnd,omitempty"`
	Properties []models.Property `json:"properties"`
}

func (c *PropertyAPIController) Get() {
	// This controller returns JSON, not HTML templates.
	c.EnableRender = false

	if propertyAPI.catalog == nil || propertyAPI.search == nil {
		c.writeJSONError(
			http.StatusInternalServerError,
			"Application is not initialized",
		)
		return
	}

	// STEP 1: Read URL query parameters.
	params := c.Ctx.Request.URL.Query()

	search := strings.TrimSpace(params.Get("search"))

	dateStart := strings.TrimSpace(params.Get("dateStart"))
	dateEnd := strings.TrimSpace(params.Get("dateEnd"))

	if _, _, err := services.ParseStayDates(dateStart, dateEnd); err != nil {
		c.writeJSONError(
			http.StatusBadRequest,
			err.Error(),
		)
		return
	}

	// Default search parameters:
	// 2 guests and Most Popular sorting.
	pax := 2
	order := 1

	if params.Has("pax") {
		value, err := strconv.Atoi(params.Get("pax"))
		if err != nil || value < 1 {
			c.writeJSONError(
				http.StatusBadRequest,
				"pax must be a positive integer",
			)
			return
		}
		pax = value
	}

	if params.Has("order") {
		value, err := strconv.Atoi(params.Get("order"))
		if err != nil || value != 1 {
			c.writeJSONError(
				http.StatusBadRequest,
				"unsupported sort order",
			)
			return
		}
		order = value
	}

	// STEP 2: Resolve the registered site.
	// All current demo sites have a homepage route.
	// Its PageLocation contains the deployment-derived
	// ContentDirectory, so we do not guess paths
	// from the requested hostname.
	location, found := propertyAPI.catalog.Resolve(
		"test-travel-templates.beta.123presto.com",
		"/",
	)

	if !found {
		c.writeJSONError(
			http.StatusNotFound,
			"Site not found",
		)
		return
	}

	// STEP 3: Search this site's property dataset.
	properties, err := propertyAPI.search.Search(
		location.ContentDirectory,
		services.PropertyFilters{
			Search:    search,
			Pax:       pax,
			Order:     order,
			DateStart: dateStart,
			DateEnd:   dateEnd,
		},
	)

	if err != nil {
		log.Printf("Property search failed: %v", err)

		c.writeJSONError(
			http.StatusInternalServerError,
			"Failed to search properties",
		)
		return
	}

	// STEP 4: Return the JSON response.
	c.Data["json"] = propertySearchResponse{
		Search:     search,
		Pax:        pax,
		Order:      order,
		Total:      len(properties),
		DateStart:  dateStart,
		DateEnd:    dateEnd,
		Properties: properties,
	}

	c.ServeJSON()
}

func (c *PropertyAPIController) writeJSONError(
	status int,
	message string,
) {
	c.EnableRender = false

	// Set the content type before committing
	// a non-200 HTTP status.
	c.Ctx.Output.Header(
		"Content-Type",
		"application/json; charset=utf-8",
	)
	c.Ctx.Output.SetStatus(status)

	c.Data["json"] = map[string]string{
		"error": message,
	}

	c.ServeJSON()
}
