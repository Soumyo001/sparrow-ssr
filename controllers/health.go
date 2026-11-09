package controllers

import (
	"net/http"

	beego "github.com/beego/beego/v2/server/web"
)

type HealthController struct {
	beego.Controller
}

type healthResponse struct {
	Status   string `json:"status"`
	API      string `json:"api"`
	Frontend string `json:"frontend"`
}

func (c *HealthController) Get() {
	// This endpoint returns JSON, not an HTML template.
	c.EnableRender = false

	// Frontend readiness: the deployment catalog and
	// content loader have been initialized.
	frontendReady := dependencies.catalog != nil &&
		dependencies.loader != nil &&
		dependencies.catalog.Count() > 0

	// API readiness: the property-search API dependencies
	// and deployment catalog have been initialized.
	apiReady := propertyAPI.catalog != nil &&
		propertyAPI.search != nil &&
		propertyAPI.catalog.Count() > 0

	result := healthResponse{
		Status:   "ok",
		API:      "ready",
		Frontend: "ready",
	}

	status := http.StatusOK

	if !frontendReady || !apiReady {
		result.Status = "unavailable"
		status = http.StatusServiceUnavailable

		if !frontendReady {
			result.Frontend = "not_ready"
		}

		if !apiReady {
			result.API = "not_ready"
		}
	}

	// Health responses must not be served from a browser
	// or intermediary cache.
	c.Ctx.Output.Header("Cache-Control", "no-store")

	c.Ctx.Output.Header(
		"Content-Type",
		"application/json; charset=utf-8",
	)

	c.Ctx.Output.SetStatus(status)

	c.Data["json"] = result
	c.ServeJSON()
}
