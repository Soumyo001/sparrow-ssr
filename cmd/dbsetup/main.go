package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"sparrow-ssr/services"

	"github.com/beego/beego/v2/client/orm"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil &&
		!errors.Is(err, os.ErrNotExist) {
		log.Fatalf("load .env: %v", err)
	}

	if err := services.InitPropertyDatabase(os.Getenv("DATABASE_URL")); err != nil {
		log.Fatalf("initialize property database: %v", err)
	}

	if err := orm.RunSyncdb("default", false, true); err != nil {
		log.Fatalf("create property schema: %v", err)
	}

	db, err := orm.GetDB("default")
	if err != nil {
		log.Fatalf("get database connection: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := ensureForeignKeys(ctx, db); err != nil {
		log.Fatalf("create foreign-key constraints: %v", err)
	}

	log.Println("PostgreSQL schema ready: tables, indexes, and foreign keys")
}

// ensureForeignKeys adds the two required relationships.
// Existing named constraints are skipped.
//
// All additions run in one transaction, so a failure rolls back
// any constraints created during this invocation.
func ensureForeignKeys(
	ctx context.Context,
	db *sql.DB,
) error {
	constraints := []struct {
		tableName      string
		constraintName string
		statement      string
	}{
		{
			tableName:      "public.properties",
			constraintName: "properties_location_id_fkey",
			statement: `
				ALTER TABLE public.properties
				ADD CONSTRAINT properties_location_id_fkey
				FOREIGN KEY (location_id)
				REFERENCES public.locations(id)
				ON DELETE RESTRICT
			`,
		},
		{
			tableName:      "public.property_availability",
			constraintName: "property_availability_property_id_fkey",
			statement: `
				ALTER TABLE public.property_availability
				ADD CONSTRAINT property_availability_property_id_fkey
				FOREIGN KEY (property_id)
				REFERENCES public.properties(id)
				ON DELETE CASCADE
			`,
		},
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin schema transaction: %w", err)
	}
	defer tx.Rollback()

	for _, constraint := range constraints {
		var exists bool

		err := tx.QueryRowContext(
			ctx,
			`
			SELECT EXISTS (
				SELECT 1
				FROM pg_constraint
				WHERE conrelid = $1::regclass
				  AND conname = $2
				  AND contype = 'f'
			)
			`,
			constraint.tableName,
			constraint.constraintName,
		).Scan(&exists)

		if err != nil {
			return fmt.Errorf(
				"check constraint %s: %w",
				constraint.constraintName,
				err,
			)
		}

		if exists {
			log.Printf(
				"Foreign key already exists: %s",
				constraint.constraintName,
			)
			continue
		}

		if _, err := tx.ExecContext(
			ctx,
			constraint.statement,
		); err != nil {
			return fmt.Errorf(
				"create constraint %s: %w",
				constraint.constraintName,
				err,
			)
		}

		log.Printf(
			"Foreign key created: %s",
			constraint.constraintName,
		)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit schema transaction: %w", err)
	}

	return nil
}
