package main

import (
	"log"

	"sparrow-ssr/routers"
	"sparrow-ssr/services"

	beego "github.com/beego/beego/v2/server/web"
)

func main() {
	const contentRoot = "content"

	catalog, err := services.NewCatalog(
		contentRoot,
		"conf/deployment.json",
	)
	if err != nil {
		log.Fatalf(
			"initialize site catalog: %v", err,
		)
	}

	loader, err := services.NewContentLoader(contentRoot)
	if err != nil {
		log.Fatalf(
			"initialize content loader: %v", err,
		)
	}

	routers.Register(catalog, loader)

	beego.BConfig.WebConfig.ViewsPath = "views"

	if err := beego.AddViewPath("views"); err != nil {
		log.Fatalf(
			"initialize HTML templates: %v", err,
		)
	}

	log.Printf("Catalog ready: %d pages", catalog.Count())

	log.Println("Server-side HTML renderer initialized")
	beego.Run()
}
