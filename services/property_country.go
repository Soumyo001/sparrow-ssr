package services

import (
	"encoding/json"
	"fmt"
	"sparrow-ssr/models"
	"strings"

	"github.com/beego/beego/v2/client/orm"
)

func (s *DatabasePropertySearchService) SearchCountry(
	siteDirectory, country string,
) ([]models.Property, error) {
	if strings.TrimSpace(siteDirectory) == "" {
		return nil, fmt.Errorf("site directory must not be empty")
	}
	if strings.TrimSpace(country) == "" {
		return nil, fmt.Errorf("country must not be empty")
	}

	var rows []models.PropertyRecord

	_, err := orm.NewOrm().QueryTable(new(models.PropertyRecord)).
		Filter("SiteKey", siteDirectory).
		Filter("Location__Country__iexact", country).
		RelatedSel("Location").
		OrderBy("-PopularityScore", "PropertyID").
		Limit(-1).All(&rows)

	if err != nil {
		return nil, fmt.Errorf("querty contains properties: %w", err)
	}

	properties := make([]models.Property, 0, len(rows))

	for _, row := range rows {
		if row.Location == nil {
			return nil, fmt.Errorf("property %s has no location", row.PropertyID)
		}

		var amenities []string
		if err := json.Unmarshal([]byte(row.AmenitiesJSON), &amenities); err != nil {
			return nil, fmt.Errorf("decode amenities for %s: %w", row.PropertyID, err)
		}

		var images []string
		if err := json.Unmarshal([]byte(row.ImageJSON), &images); err != nil {
			return nil, fmt.Errorf("decode images for %s: %w", row.PropertyID, err)
		}

		properties = append(properties, models.Property{
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

	return properties, nil
}
