package services

import (
	"errors"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

// Run explicitly with SPARROW_TEST_NEON=1.
// This test reads the seeded Neon database only.
func TestDatabasePropertySearchSeededNeon(t *testing.T) {
	if os.Getenv("SPARROW_TEST_NEON") != "1" {
		t.Skip("set SPARROW_TEST_NEON=1 to run the read-only Neon test")
	}

	if err := godotenv.Load("../.env"); err != nil &&
		!errors.Is(err, os.ErrNotExist) {
		t.Fatalf("load environment: %v", err)
	}

	databaseURL := os.Getenv("DATABASE_URL_DIRECT")
	if databaseURL == "" {
		databaseURL = os.Getenv("DATABASE_URL")
	}

	if databaseURL == "" {
		t.Fatal("DATABASE_URL_DIRECT or DATABASE_URL is required")
	}

	if err := InitPropertyDatabase(databaseURL); err != nil {
		t.Fatalf("initialize database: %v", err)
	}

	const site = "beta/preview/test-travel-templates"

	search := NewDatabasePropertySearchService()

	for _, tc := range []struct {
		name    string
		site    string
		filters PropertyFilters
		want    int
		first   string
	}{
		{
			"all", site,
			PropertyFilters{Search: "UsA", Pax: 2, Order: 1},
			15, "prop-002",
		},
		{
			"beverly", site,
			PropertyFilters{Search: "bEvErLy", Pax: 2, Order: 1},
			1, "prop-007",
		},
		{
			"eight_guests", site,
			PropertyFilters{Search: "usa", Pax: 8, Order: 1},
			2, "prop-003",
		},
		{
			"twenty_guests", site,
			PropertyFilters{Search: "usa", Pax: 20, Order: 1},
			0, "",
		},
		{
			"unknown_site", "beta/preview/unknown",
			PropertyFilters{Search: "usa", Pax: 2, Order: 1},
			0, "",
		},
		{
			"unknown_destination", site,
			PropertyFilters{Search: "Atlantis", Pax: 2, Order: 1},
			0, "",
		},
		{
			"early_dates", site,
			PropertyFilters{
				Search:    "usa",
				Pax:       2,
				Order:     1,
				DateStart: "2026-11-06",
				DateEnd:   "2026-11-13",
			},
			10, "prop-002",
		},
		{
			"later_dates", site,
			PropertyFilters{
				Search:    "usa",
				Pax:       2,
				Order:     1,
				DateStart: "2026-11-16",
				DateEnd:   "2026-11-20",
			},
			15, "prop-002",
		},
		{
			"out_of_window", site,
			PropertyFilters{
				Search:    "usa",
				Pax:       2,
				Order:     1,
				DateStart: "2027-12-30",
				DateEnd:   "2028-01-01",
			},
			0, "",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := search.Search(tc.site, tc.filters)
			if err != nil {
				t.Fatal(err)
			}

			if len(got) != tc.want {
				t.Fatalf(
					"want %d properties, got %d",
					tc.want,
					len(got),
				)
			}

			if tc.first != "" && got[0].ID != tc.first {
				t.Fatalf(
					"first property = %q, want %q",
					got[0].ID,
					tc.first,
				)
			}

			for _, p := range got {
				if p.Slug == "" ||
					p.MaxGuests < 1 ||
					p.PriceFrom < 1 ||
					p.Location.Country == "" ||
					len(p.Images) == 0 ||
					len(p.Amenities) == 0 {
					t.Fatalf(
						"incomplete database property %q: %+v",
						p.ID,
						p,
					)
				}
			}
		})
	}
}
