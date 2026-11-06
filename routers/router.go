package routers

import (
	"sparrow-ssr/controllers"
	"sparrow-ssr/services"

	beego "github.com/beego/beego/v2/server/web"
)

// Register keeps the existing server-rendered routes.
func Register(
	catalog *services.Catalog,
	loader *services.ContentLoader,
) {
	controllers.SetDependencies(catalog, loader)

	beego.Router("/", &controllers.PageController{})
	beego.Router("/*", &controllers.PageController{})
}

// RegisterPropertyAPI adds the property-search endpoint.
// It is separate from Register so the existing SSR
// registration interface remains unchanged.
func RegisterPropertyAPI(
	catalog *services.Catalog,
	search *services.PropertySearchService,
) {
	controllers.SetPropertyAPIDependencies(
		catalog,
		search,
	)

	beego.Router(
		"/api/properties",
		&controllers.PropertyAPIController{},
		"get:Get",
	)
}
