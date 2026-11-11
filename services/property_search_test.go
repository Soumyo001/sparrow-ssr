package services

import "testing"

// Validation happens before the database is accessed, so these tests
// do not require a running PostgreSQL instance.
func TestDatabasePropertySearchRejectsInvalidFilters(t *testing.T) {
	const site = "beta/preview/test-travel-templates"

	cases := []struct {
		name    string
		site    string
		filters PropertyFilters
	}{
		{"empty_site", "", PropertyFilters{Pax: 2, Order: 1}},
		{"invalid_guests", site, PropertyFilters{Pax: 0, Order: 1}},
		{"invalid_order", site, PropertyFilters{Pax: 2, Order: 9}},
		{"missing_end_date", site, PropertyFilters{Pax: 2, Order: 1, DateStart: "2026-11-06"}},
		{"reversed_dates", site, PropertyFilters{Pax: 2, Order: 1, DateStart: "2026-11-13", DateEnd: "2026-11-06"}},
	}

	search := NewDatabasePropertySearchService()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := search.Search(tc.site, tc.filters); err == nil {
				t.Fatalf("expected error for site=%q filters=%+v", tc.site, tc.filters)
			}
		})
	}
}
