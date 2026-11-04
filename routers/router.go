package routers

import (
	"sparrow-ssr/controllers"
	"sparrow-ssr/services"

	beego "github.com/beego/beego/v2/server/web"
)

func Register(
	catalog *services.Catalog,
	loader *services.ContentLoader,
) {
	controllers.SetDependencies(catalog, loader)

	beego.Router("/", &controllers.PageController{})
	beego.Router("/*", &controllers.PageController{})
}
