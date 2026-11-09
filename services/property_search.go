package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"sparrow-ssr/models"
)

type PropertySearchService struct {
	root string
}

type PropertyFilters struct {
	Search    string
	Pax       int
	Order     int
	DateStart string
	DateEnd   string
}

func NewPropertySearchService(
	mockRoot string,
) (*PropertySearchService, error) {
	root, err := filepath.Abs(mockRoot)
	if err != nil {
		return nil, fmt.Errorf(
			"resolve mock data root: %w", err,
		)
	}

	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf(
			"stat mock data root: %w", err,
		)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf(
			"mock data root is not a directory: %s",
			root,
		)
	}

	return &PropertySearchService{root: root}, nil
}

// ParseStayDates accepts either no dates or
// a valid [check-in, checkout) range.
func ParseStayDates(
	start, end string,
) (time.Time, time.Time, error) {
	if start == "" && end == "" {
		return time.Time{}, time.Time{}, nil
	}

	if start == "" || end == "" {
		return time.Time{}, time.Time{},
			fmt.Errorf("dateStart and dateEnd must both be provided")
	}

	checkIn, err := time.Parse("2006-01-02", start)
	if err != nil {
		return time.Time{}, time.Time{},
			fmt.Errorf("dateStart must be YYYY-MM-DD")
	}

	checkOut, err := time.Parse("2006-01-02", end)
	if err != nil {
		return time.Time{}, time.Time{},
			fmt.Errorf("dateEnd must be YYYY-MM-DD")
	}

	if !checkOut.After(checkIn) {
		return time.Time{}, time.Time{},
			fmt.Errorf(
				"dateEnd must be after dateStart",
			)
	}

	return checkIn, checkOut, nil
}

func (s *PropertySearchService) Search(
	siteDirectory string,
	filters PropertyFilters,
) ([]models.Property, error) {
	if !filepath.IsLocal(siteDirectory) || filepath.Clean(siteDirectory) != siteDirectory {
		return nil, fmt.Errorf(
			"invalid site directory %q",
			siteDirectory,
		)
	}

	if filters.Pax < 1 {
		return nil, fmt.Errorf(
			"pax must be at least 1",
		)
	}

	if filters.Order != 1 {
		return nil, fmt.Errorf(
			"unsupported sort order %d",
			filters.Order,
		)
	}

	start, end, err := ParseStayDates(
		filters.DateStart,
		filters.DateEnd,
	)
	if err != nil {
		return nil, err
	}

	// Load this site's existing listing data.
	filename := filepath.Join(s.root, siteDirectory, "properties.json")

	contents, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf(
			"read mock properties: %w", err,
		)
	}

	var data models.PropertyDataset

	if err := json.Unmarshal(contents, &data); err != nil {
		return nil, fmt.Errorf(
			"decode mock properties: %w", err,
		)
	}

	// Load availability only for date searches.
	// Search without dates preserves old behavior.
	var availability models.PropertyAvailability

	if !start.IsZero() {
		filename := filepath.Join(s.root, siteDirectory, "availability.json")

		contents, err := os.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf(
				"read mock availability: %w", err,
			)
		}

		if err := json.Unmarshal(contents, &availability); err != nil {
			return nil, fmt.Errorf(
				"decode mock availability: %w", err,
			)
		}
	}

	query := strings.ToLower(
		strings.TrimSpace(filters.Search),
	)

	results := make([]models.Property, 0, len(data.Properties))

	for _, property := range data.Properties {
		// Use actual capacity when known.
		// Otherwise, fall back to the illustrative
		// capacity supplied for this demonstration.
		capacity := property.MaxGuests

		if capacity <= 0 {
			capacity = property.DemoMaxGuests
		}

		if (capacity > 0 && capacity < filters.Pax) ||
			(query != "" && !matchesProperty(property, query)) {
			continue
		}

		// New date availability filter.
		if !start.IsZero() {
			matched, err := coversStay(
				availability.Properties[property.ID],
				start,
				end,
			)

			if err != nil {
				return nil, fmt.Errorf(
					"invalid availability for %s: %w",
					property.ID,
					err,
				)
			}

			if !matched {
				continue
			}
		}

		results = append(results, property)
	}

	// order=1: Most Popular, descending.
	// Property ID resolves equal-score ties.
	sort.Slice(results, func(i, j int) bool {
		if results[i].PopularityScore == results[j].PopularityScore {
			return results[i].ID < results[j].ID
		}

		return results[i].PopularityScore > results[j].PopularityScore
	})

	return results, nil
}

// A stay must fit fully inside one available window.
func coversStay(
	windows []models.AvailabilityWindow,
	start, end time.Time,
) (bool, error) {
	for _, window := range windows {
		first, err := time.Parse("2006-01-02", window.Start)
		if err != nil {
			return false, fmt.Errorf(
				"invalid window start %q",
				window.Start,
			)
		}

		last, err := time.Parse("2006-01-02", window.End)
		if err != nil || !last.After(first) {
			return false, fmt.Errorf(
				"invalid window end %q",
				window.End,
			)
		}

		if !start.Before(first) && !end.After(last) {
			return true, nil
		}
	}

	return false, nil
}

func matchesProperty(
	property models.Property,
	query string,
) bool {
	fields := []string{
		property.Name,
		property.Location.City,
		property.Location.State,
		property.Location.Area,
		property.Location.Country,
	}

	for _, field := range fields {
		if strings.Contains(strings.ToLower(field), query) {
			return true
		}
	}

	return false
}
