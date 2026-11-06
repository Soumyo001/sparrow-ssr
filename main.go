package main

import (
	"log"
	"os"
	"strings"

	"sparrow-ssr/routers"
	"sparrow-ssr/services"

	beego "github.com/beego/beego/v2/server/web"
)

func main() {
	const contentRoot string = "content"

	// Existing single-deployment startup remains the default.
	deploymentFiles := []string{"conf/deployment.json"}

	// Optional, explicit comma-separated deployment metadata file paths.
	if configured := os.Getenv("SPARROW_DEPLOYMENTS"); configured != "" {
		deploymentFiles = strings.Split(configured, ",")
		for i, file := range deploymentFiles {
			deploymentFiles[i] = strings.TrimSpace(file)
		}
	}

	catalog, err := services.NewMultiCatalog(contentRoot, deploymentFiles)
	if err != nil {
		log.Fatalf("initialize site catalog: %v", err)
	}

	loader, err := services.NewContentLoader(contentRoot)
	if err != nil {
		log.Fatalf("initialize content loader: %v", err)
	}

	routers.Register(catalog, loader)

	beego.BConfig.WebConfig.ViewsPath = "views"

	if err := beego.AddViewPath("views"); err != nil {
		log.Fatalf("initialize HTML templates: %v", err)
	}

	log.Printf("Catalog ready: %d pages from %d deployment(s)", catalog.Count(), len(deploymentFiles))
	beego.Run("127.0.0.1:8080")
}
