package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"sparrow-ssr/models"
)

// PropertySearchService reads demo listings
// from a separate mock-data root.
type PropertySearchService struct {
	root string
}

// PropertyFilters represents the search parameters.
type PropertyFilters struct {
	Search string
	Pax    int
	Order  int
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

	return &PropertySearchService{
		root: root,
	}, nil
}

// Search selects listings within one site's explicitly supplied storage directory.
// Host-to-directory resolution belongs to the deployment catalog, not this service.
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

	filename := filepath.Join(
		s.root,
		siteDirectory,
		"properties.json",
	)

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

	query := strings.ToLower(
		strings.TrimSpace(filters.Search),
	)

	results := make([]models.Property, 0, len(data.Properties))

	for _, property := range data.Properties {
		if property.MaxGuests < filters.Pax {
			continue
		}

		if query != "" && !matchesProperty(property, query) {
			continue
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
