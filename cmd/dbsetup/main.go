package main

import (
	"errors"
	"log"
	"os"
	"sparrow-ssr/services"

	"github.com/beego/beego/v2/client/orm"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatalf("load .env: %v", err)
	}

	if err := services.InitPropertyDatabase(os.Getenv("DATABASE_URL")); err != nil {
		log.Fatalf("initialize property database: %v", err)
	}

	if err := orm.RunSyncdb("default", false, true); err != nil {
		log.Fatalf("create property schema: %v", err)
	}

	log.Println(
		"PostgreSQL schema ready: sites, locations, properties, property_availability",
	)
}
