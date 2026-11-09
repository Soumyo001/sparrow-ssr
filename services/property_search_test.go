package services

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"sparrow-ssr/models"
)

func TestPropertySearchDemoData(t *testing.T) {
	service, err := NewPropertySearchService(
		"../mock_data",
	)
	if err != nil {
		t.Fatal(err)
	}

	const site = "beta/preview/test-travel-templates"

	// Search for USA with two guests.
	all, err := service.Search(
		site,
		PropertyFilters{
			Search: "UsA",
			Pax:    2,
			Order:  1,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(all) != 15 {
		t.Fatalf(
			"want 15 USA properties, got %d",
			len(all),
		)
	}

	// The real-hotel demo ranks by review count.
	// The two highest-ranking fixtures are now
	// prop-002 and prop-003.
	if all[0].ID != "prop-002" ||
		all[1].ID != "prop-003" {
		t.Fatalf(
			"incorrect review-count order: %q, %q",
			all[0].ID,
			all[1].ID,
		)
	}

	// Destination search remains case-insensitive.
	// The new catalog has a Beverly Hills hotel
	// instead of the previous fictional Austin listing.
	beverly, err := service.Search(
		site,
		PropertyFilters{
			Search: "BeVeRlY",
			Pax:    2,
			Order:  1,
		},
	)

	if err != nil ||
		len(beverly) != 1 ||
		beverly[0].ID != "prop-007" {
		t.Fatalf(
			"Beverly filter mismatch: %#v, err=%v",
			beverly, err,
		)
	}

	// Capacity is unknown for these real hotels.
	// max_guests = 0 means unknown, not zero.
	// The service retains these records instead
	// of falsely excluding them by guest count.
	large, err := service.Search(
		site,
		PropertyFilters{
			Search: "usa",
			Pax:    8,
			Order:  1,
		},
	)

	if err != nil || len(large) != 2 ||
		large[0].ID != "prop-003" ||
		large[1].ID != "prop-011" {
		t.Fatalf(
			"8-guest demo capacity mismatch: %#v, err=%v",
			large, err,
		)
	}

	noLargeGroup, err := service.Search(
		site,
		PropertyFilters{
			Search: "usa",
			Pax:    20,
			Order:  1,
		},
	)

	if err != nil || len(noLargeGroup) != 0 {
		t.Fatalf(
			"20-guest demo capacity mismatch: %#v, err=%v",
			noLargeGroup, err,
		)
	}

	// Unknown destination returns no properties.
	absent, err := service.Search(
		site,
		PropertyFilters{
			Search: "Atlantis",
			Pax:    2,
			Order:  1,
		},
	)

	if err != nil || len(absent) != 0 {
		t.Fatalf(
			"unknown destination: got %d, err=%v",
			len(absent), err,
		)
	}

	// Invalid directories and filter values.
	for _, tc := range []struct {
		directory string
		filters   PropertyFilters
	}{
		{
			"../content",
			PropertyFilters{Pax: 2, Order: 1},
		},
		{
			site,
			PropertyFilters{Pax: 0, Order: 1},
		},
		{
			site,
			PropertyFilters{Pax: 2, Order: 9},
		},
	} {
		if _, err := service.Search(
			tc.directory, tc.filters,
		); err == nil {
			t.Fatalf(
				"expected error for %q and %+v",
				tc.directory, tc.filters,
			)
		}
	}
}

func TestPropertySearchSiteIsolation(t *testing.T) {
	root := t.TempDir()

	// Construct independent property data for two sites.
	sites := []struct {
		site string
		id   string
	}{
		{"beta/preview/alpha", "alpha-001"},
		{"beta/preview/bravo", "bravo-001"},
	}

	for _, tc := range sites {
		folder := filepath.Join(
			root,
			filepath.FromSlash(tc.site),
		)

		if err := os.MkdirAll(folder, 0755); err != nil {
			t.Fatal(err)
		}

		data := models.PropertyDataset{
			Properties: []models.Property{
				{
					ID:        tc.id,
					MaxGuests: 2,
					Location: models.PropertyLocation{
						Country: "USA",
					},
				},
			},
		}

		contents, err := json.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}

		filename := filepath.Join(
			folder, "properties.json",
		)

		if err := os.WriteFile(
			filename, contents, 0644,
		); err != nil {
			t.Fatal(err)
		}
	}

	service, err := NewPropertySearchService(root)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range sites {
		got, err := service.Search(
			tc.site,
			PropertyFilters{
				Search: "usa",
				Pax:    2,
				Order:  1,
			},
		)

		if err != nil ||
			len(got) != 1 ||
			got[0].ID != tc.id {
			t.Fatalf(
				"site %q returned %#v, err=%v",
				tc.site, got, err,
			)
		}
	}
}
