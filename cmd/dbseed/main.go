package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"sparrow-ssr/models"
	"sparrow-ssr/services"

	"github.com/beego/beego/v2/client/orm"
	"github.com/joho/godotenv"
)

const siteKey = "beta/preview/test-travel-templates"
const availabilityEnd = "2027-12-31"

type seedProperty struct {
	ID              string
	Slug            string
	Name            string
	Type            string
	Location        models.PropertyLocation
	MaxGuests       int
	SizeSqFt        int
	Rating          float64
	ReviewCount     int
	PriceFrom       int
	Currency        string
	Amenities       []string
	ProviderName    string
	BookingURL      string
	PopularityScore int
	EcoFriendly     bool
	Images          []string
	AvailableFrom   string
}

type seedStats struct {
	Locations  int
	Properties int
	Windows    int
}

func main() {
	if err := godotenv.Load(); err != nil &&
		!errors.Is(err, os.ErrNotExist) {
		log.Fatalf("load .env: %v", err)
	}

	if err := validateSeedData(seedProperties); err != nil {
		log.Fatalf("invalid seed data: %v", err)
	}

	databaseURL := os.Getenv("DATABASE_URL_DIRECT")
	if databaseURL == "" {
		databaseURL = os.Getenv("DATABASE_URL")
	}

	if err := services.InitPropertyDatabase(databaseURL); err != nil {
		log.Fatalf("initialize PostgreSQL: %v", err)
	}

	var stats seedStats

	err := orm.NewOrm().DoTx(
		func(ctx context.Context, tx orm.TxOrmer) error {
			knownLocations := make(
				map[models.PropertyLocation]*models.LocationRecord,
			)

			end, err := time.Parse("2006-01-02", availabilityEnd)
			if err != nil {
				return err
			}

			for _, p := range seedProperties {
				location := knownLocations[p.Location]

				if location == nil {
					candidate := &models.LocationRecord{
						City:    p.Location.City,
						State:   p.Location.State,
						Area:    p.Location.Area,
						Country: p.Location.Country,
					}

					isNew, locationID, err := tx.ReadOrCreate(
						candidate,
						"Country", "State", "City", "Area",
					)
					if err != nil {
						return fmt.Errorf(
							"seed location for %s: %w", p.ID, err,
						)
					}

					candidate.Id = locationID
					if isNew {
						stats.Locations++
					}

					knownLocations[p.Location] = candidate
					location = candidate
				}

				amenities, err := json.Marshal(p.Amenities)
				if err != nil {
					return fmt.Errorf(
						"encode amenities for %s: %w", p.ID, err,
					)
				}

				images, err := json.Marshal(p.Images)
				if err != nil {
					return fmt.Errorf(
						"encode images for %s: %w", p.ID, err,
					)
				}

				record := &models.PropertyRecord{
					SiteKey:         siteKey,
					Slug:            p.Slug,
					Location:        location,
					PropertyID:      p.ID,
					Name:            p.Name,
					PropertyType:    p.Type,
					MaxGuests:       p.MaxGuests,
					SizeSqFt:        p.SizeSqFt,
					Rating:          p.Rating,
					ReviewCount:     p.ReviewCount,
					PriceFrom:       p.PriceFrom,
					Currency:        p.Currency,
					AmenitiesJSON:   string(amenities),
					ProviderName:    p.ProviderName,
					BookingURL:      p.BookingURL,
					PopularityScore: p.PopularityScore,
					EcoFriendly:     p.EcoFriendly,
					ImageJSON:       string(images),
				}

				isNew, propertyID, err := tx.ReadOrCreate(
					record, "SiteKey", "PropertyID",
				)
				if err != nil {
					return fmt.Errorf(
						"seed property %s: %w", p.ID, err,
					)
				}

				record.Id = propertyID
				if isNew {
					stats.Properties++
				}

				start, err := time.Parse(
					"2006-01-02", p.AvailableFrom,
				)
				if err != nil {
					return fmt.Errorf(
						"parse availability for %s: %w", p.ID, err,
					)
				}

				window := &models.AvailabilityRecord{
					Property:  record,
					StartDate: start,
					EndDate:   end,
				}

				isNew, _, err = tx.ReadOrCreate(
					window,
					"Property", "StartDate", "EndDate",
				)
				if err != nil {
					return fmt.Errorf(
						"seed availability for %s: %w", p.ID, err,
					)
				}

				if isNew {
					stats.Windows++
				}
			}

			return nil
		},
	)
	if err != nil {
		log.Fatalf("seed transaction rolled back: %v", err)
	}

	log.Printf(
		"Seed complete. New records: locations=%d properties=%d availability=%d",
		stats.Locations, stats.Properties, stats.Windows,
	)
	log.Println(
		"Existing rows were preserved; rerunning does not overwrite modified property data.",
	)
}

func validateSeedData(properties []seedProperty) error {
	if len(properties) == 0 {
		return errors.New("no properties supplied")
	}

	ids := make(map[string]bool, len(properties))
	slugs := make(map[string]bool, len(properties))

	end, err := time.Parse("2006-01-02", availabilityEnd)
	if err != nil {
		return fmt.Errorf("invalid common availability end: %w", err)
	}

	for _, p := range properties {
		if p.ID == "" || ids[p.ID] {
			return fmt.Errorf(
				"missing or duplicate property ID %q", p.ID,
			)
		}
		ids[p.ID] = true

		if !validSlug(p.Slug) || slugs[p.Slug] {
			return fmt.Errorf(
				"invalid or duplicate property slug %q for %s",
				p.Slug, p.ID,
			)
		}
		slugs[p.Slug] = true

		if p.Name == "" ||
			p.Type == "" ||
			p.Location.City == "" ||
			p.Location.State == "" ||
			p.Location.Country == "" ||
			p.Location.Area == "" {
			return fmt.Errorf(
				"missing identity or location for %s", p.ID,
			)
		}

		if p.MaxGuests < 1 ||
			p.PriceFrom < 1 ||
			p.Currency == "" {
			return fmt.Errorf(
				"invalid capacity or price for %s", p.ID,
			)
		}

		if p.ProviderName == "" ||
			p.BookingURL == "" ||
			len(p.Images) == 0 {
			return fmt.Errorf(
				"missing provider or image for %s", p.ID,
			)
		}

		if len(p.Amenities) == 0 {
			return fmt.Errorf(
				"missing amenities for %s", p.ID,
			)
		}

		start, err := time.Parse(
			"2006-01-02", p.AvailableFrom,
		)
		if err != nil || !end.After(start) {
			return fmt.Errorf(
				"invalid availability range for %s", p.ID,
			)
		}
	}

	return nil
}

// validSlug keeps stable property slugs lowercase and URL-safe.
func validSlug(slug string) bool {
	if slug == "" ||
		len(slug) > 255 ||
		slug[0] == '-' ||
		slug[len(slug)-1] == '-' {
		return false
	}

	previousHyphen := false

	for _, r := range slug {
		switch {
		case r >= 'a' && r <= 'z',
			r >= '0' && r <= '9':
			previousHyphen = false

		case r == '-' && !previousHyphen:
			previousHyphen = true

		default:
			return false
		}
	}

	return true
}
