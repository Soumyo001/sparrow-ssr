package tests

import (
	"strings"
	"time"

	"sparrow-ssr/models"
	"sparrow-ssr/services"
)

// fixturePropertySearcher exercises controller wiring and site-key propagation.
// It never reads property JSON or connects to PostgreSQL. The real database
// query behavior is covered by the opt-in Neon integration tests.
type fixturePropertySearcher map[string]fixturePropertyRecord

type fixturePropertyRecord struct {
	property      models.Property
	availableFrom string
}

func (s fixturePropertySearcher) Search(
	siteDirectory string,
	filters services.PropertyFilters,
) ([]models.Property, error) {
	start, end, err := services.ParseStayDates(
		filters.DateStart,
		filters.DateEnd,
	)
	if err != nil {
		return nil, err
	}

	record, found := s[siteDirectory]
	if !found {
		return []models.Property{}, nil
	}

	property := record.property
	query := strings.ToLower(strings.TrimSpace(filters.Search))

	if filters.Pax > property.MaxGuests ||
		(query != "" &&
			!strings.Contains(strings.ToLower(property.Name), query) &&
			!strings.Contains(strings.ToLower(property.Location.Country), query)) {
		return []models.Property{}, nil
	}

	if !start.IsZero() {
		availableFrom, err := time.Parse(
			"2006-01-02",
			record.availableFrom,
		)
		if err != nil {
			return nil, err
		}

		availableUntil := time.Date(
			2027, 12, 31, 0, 0, 0, 0, time.UTC,
		)

		if start.Before(availableFrom) || end.After(availableUntil) {
			return []models.Property{}, nil
		}
	}

	return []models.Property{property}, nil
}

// SearchCountry is an exact-country test fixture.
// It never accesses the production database.
func (s fixturePropertySearcher) SearchCountry(
	siteDirectory, country string,
) ([]models.Property, error) {
	record, found := s[siteDirectory]

	if !found ||
		!strings.EqualFold(
			record.property.Location.Country,
			country,
		) {
		return []models.Property{}, nil
	}

	return []models.Property{record.property}, nil
}
