package services

import (
	"encoding/json"
	"fmt"
	"strings"

	"sparrow-ssr/models"

	"github.com/beego/beego/v2/client/orm"
)

// DatabasePropertySearchService reads accommodation data exclusively from PostgreSQL.
type DatabasePropertySearchService struct{}

func NewDatabasePropertySearchService() *DatabasePropertySearchService {
	return &DatabasePropertySearchService{}
}

func (s *DatabasePropertySearchService) Search(
	siteDirectory string,
	filters PropertyFilters,
) ([]models.Property, error) {
	if strings.TrimSpace(siteDirectory) == "" {
		return nil, fmt.Errorf("site directory must not be empty")
	}

	if filters.Pax < 1 {
		return nil, fmt.Errorf("pax must be at least 1")
	}

	if filters.Order != 1 {
		return nil, fmt.Errorf("unsupported sort order %d", filters.Order)
	}

	start, end, err := ParseStayDates(
		filters.DateStart,
		filters.DateEnd,
	)
	if err != nil {
		return nil, err
	}

	o := orm.NewOrm()

	// Restrict every search to the deployment-derived site
	// and properties capable of accommodating the guest count.
	condition := orm.NewCondition().
		And("SiteKey", siteDirectory).
		And("MaxGuests__gte", filters.Pax)

	// Search across the property name and related location fields.
	if query := strings.TrimSpace(filters.Search); query != "" {
		destination := orm.NewCondition().
			Or("Name__icontains", query).
			Or("Location__City__icontains", query).
			Or("Location__State__icontains", query).
			Or("Location__Area__icontains", query).
			Or("Location__Country__icontains", query)

		condition = condition.AndCond(destination)
	}

	// A stay must fit entirely inside one availability window.
	// Fetch matching property IDs in one query, not one query
	// per property.
	if !start.IsZero() {
		var windows []models.AvailabilityRecord

		_, err := o.QueryTable(new(models.AvailabilityRecord)).
			Filter("Property__SiteKey", siteDirectory).
			Filter("StartDate__lte", start).
			Filter("EndDate__gte", end).
			RelatedSel("Property").
			Limit(-1).All(&windows)

		if err != nil {
			return nil, fmt.Errorf(
				"query property availability: %w",
				err,
			)
		}

		if len(windows) == 0 {
			return []models.Property{}, nil
		}

		// Multiple matching windows may reference the same property.
		// Remove duplicate database IDs before filtering.
		ids := make([]int64, 0, len(windows))
		seen := make(map[int64]bool, len(windows))

		for _, window := range windows {
			if window.Property == nil || window.Property.Id < 1 {
				return nil, fmt.Errorf(
					"availability record has no property ID",
				)
			}

			if !seen[window.Property.Id] {
				seen[window.Property.Id] = true
				ids = append(ids, window.Property.Id)
			}
		}

		values := make([]interface{}, len(ids))
		for i, id := range ids {
			values[i] = id
		}

		condition = condition.And("Id__in", values...)
	}

	// Load matching properties and their related location records.
	// Preserve existing most-popular-first ordering.
	var rows []models.PropertyRecord

	_, err = o.QueryTable(new(models.PropertyRecord)).
		SetCond(condition).
		RelatedSel("Location").
		OrderBy("-PopularityScore", "PropertyID").
		Limit(-1).
		All(&rows)

	if err != nil {
		return nil, fmt.Errorf("query properties: %w", err)
	}

	// Convert database records into the existing public API shape.
	// This preserves the frontend's expected JSON structure.
	results := make([]models.Property, 0, len(rows))

	for _, row := range rows {
		if row.Location == nil {
			return nil, fmt.Errorf(
				"property %s has no related location",
				row.PropertyID,
			)
		}

		var amenities []string

		if err := json.Unmarshal([]byte(row.AmenitiesJSON), &amenities); err != nil {
			return nil, fmt.Errorf(
				"decode amenities for %s: %w",
				row.PropertyID,
				err,
			)
		}

		var images []string

		if err := json.Unmarshal([]byte(row.ImageJSON), &images); err != nil {
			return nil, fmt.Errorf(
				"decode images for %s: %w",
				row.PropertyID,
				err,
			)
		}

		results = append(results, models.Property{
			ID:           row.PropertyID,
			Slug:         row.Slug,
			Name:         row.Name,
			PropertyType: row.PropertyType,
			Location: models.PropertyLocation{
				City:    row.Location.City,
				State:   row.Location.State,
				Area:    row.Location.Area,
				Country: row.Location.Country,
			},
			MaxGuests:       row.MaxGuests,
			SizeSqFt:        row.SizeSqFt,
			Rating:          row.Rating,
			ReviewCount:     row.ReviewCount,
			PriceFrom:       row.PriceFrom,
			Currency:        row.Currency,
			Amenities:       amenities,
			ProviderName:    row.ProviderName,
			BookingURL:      row.BookingURL,
			PopularityScore: row.PopularityScore,
			EcoFriendly:     row.EcoFriendly,
			Images:          images,
		})
	}

	return results, nil
}
