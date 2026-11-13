package main

import (
	"errors"
	"log"
	"os"
	"strings"

	"sparrow-ssr/routers"
	"sparrow-ssr/services"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Fatalf("load .env file: %v", err)
		}
	}

	const contentRoot string = "content"

	// Existing deployment configuration.
	deploymentFiles := []string{"conf/deployment.json"}

	// Optionally override the deployment metadata files.
	if configured := strings.TrimSpace(os.Getenv("SPARROW_DEPLOYMENTS")); configured != "" {
		deploymentFiles = strings.Split(configured, ",")

		for i, file := range deploymentFiles {
			deploymentFiles[i] = strings.TrimSpace(file)
		}
	}

	catalog, err := services.NewMultiCatalog(
		contentRoot,
		deploymentFiles,
	)
	if err != nil {
		log.Fatalf("initialize site catalog: %v", err)
	}

	// An optional, explicitly registered preview hostname for render
	previewHost := strings.TrimSpace(
		os.Getenv("SPARROW_PREVIEW_HOST"),
	)
	previewSource := strings.TrimSpace(
		os.Getenv("SPARROW_PREVIEW_SOURCE_HOST"),
	)

	if previewHost != "" || previewSource != "" {
		if err := catalog.RegisterPreviewAlias(
			previewSource,
			previewHost,
		); err != nil {
			log.Fatalf(
				"register preview hostname: %v",
				err,
			)
		}
	}

	// Initialize the frontend content loader.
	loader, err := services.NewContentLoader(contentRoot)
	if err != nil {
		log.Fatalf(
			"initialize content loader: %v", err,
		)
	}

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL_POOLED"))
	if databaseURL == "" {
		databaseURL = strings.TrimSpace(os.Getenv("DATABASE_URL"))
	}

	if err := services.InitPropertyDatabase(databaseURL); err != nil {
		log.Fatalf("initialize PostgreSQL: %v", err)
	}

	propertySearch := services.NewDatabasePropertySearchService()

	routers.Register(catalog, loader)
	routers.RegisterPropertyAPI(catalog, propertySearch)

	beego.BConfig.WebConfig.ViewsPath = "views"

	if err := beego.AddViewPath("views"); err != nil {
		log.Fatalf(
			"initialize HTML templates: %v", err,
		)
	}

	log.Printf(
		"Catalog ready: %d routes from %d deployment(s)",
		catalog.Count(),
		len(deploymentFiles),
	)

	// Default local development address.
	address := "127.0.0.1:8080"

	// Render provides PORT as an environment variable.
	// Bind to all interfaces so Render's proxy can reach
	// the application.
	if port := strings.TrimSpace(os.Getenv("PORT")); port != "" {
		address = "0.0.0.0:" + port
	}

	log.Printf("Listening on %s", address)

	beego.Run(address)
}
