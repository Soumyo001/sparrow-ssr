package services

import (
	"context"
	"fmt"
	"net/url"
	"sparrow-ssr/models"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func InitPropertyDatabase(databaseURL string) error {
	databaseURL = strings.TrimSpace(databaseURL)
	if databaseURL == "" {
		return fmt.Errorf("PostgreSQL connection URL is not configured")
	}

	parsed, err := url.Parse(databaseURL)
	if err != nil ||
		(parsed.Scheme != "postgres" && parsed.Scheme != "postgresql") ||
		parsed.Host == "" ||
		parsed.User == nil ||
		strings.TrimPrefix(parsed.Path, "/") == "" {
		return fmt.Errorf("invalid PostgreSQL connection URL")
	}

	switch parsed.Query().Get("sslmode") {
	case "require", "verify-ca", "verify-full":
	default:
		return fmt.Errorf("PostgreSQL connection must require TLS")
	}

	orm.DefaultTimeLoc = time.UTC

	orm.RegisterModel(
		new(models.SiteRecord),
		new(models.LocationRecord),
		new(models.PropertyRecord),
		new(models.AvailabilityRecord),
	)

	if err := orm.RegisterDriver("pgx", orm.DRPostgres); err != nil {
		return fmt.Errorf("register PostgreSQL driver: %w", err)
	}

	if err := orm.RegisterDataBase("default", "pgx", databaseURL); err != nil {
		return fmt.Errorf("register PostgreSQL database: %w", err)
	}

	db, err := orm.GetDB("default")
	if err != nil {
		return fmt.Errorf("obtain PostgreSQL connection pool: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(
		context.Background(), 15*time.Second,
	)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("connect to PostgreSQL: %w", err)
	}

	return nil
}
