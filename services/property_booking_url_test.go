package services

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"sparrow-ssr/models"
)

// This test uses a reserved .invalid hostname.
// It never makes a network request.
func TestPropertyBookingURLRoundTrip(t *testing.T) {
	root := t.TempDir()

	const siteDirectory = "beta/preview/booking-link-test"

	folder := filepath.Join(
		root,
		filepath.FromSlash(siteDirectory),
	)

	if err := os.MkdirAll(folder, 0755); err != nil {
		t.Fatal(err)
	}

	const sampleURL = "https://booking.example.invalid/hotel/fixture"

	data := models.PropertyDataset{
		Properties: []models.Property{
			{
				ID:        "linked",
				MaxGuests: 2,
				Location: models.PropertyLocation{
					Country: "USA",
				},
				ProviderName:    "Booking.com",
				BookingURL:      sampleURL,
				PopularityScore: 20,
			},
			{
				ID:        "unlinked",
				MaxGuests: 2,
				Location: models.PropertyLocation{
					Country: "USA",
				},
				ProviderName:    "Expedia",
				PopularityScore: 10,
			},
		},
	}

	content, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(
		filepath.Join(folder, "properties.json"),
		content,
		0644,
	); err != nil {
		t.Fatal(err)
	}

	service, err := NewPropertySearchService(root)
	if err != nil {
		t.Fatal(err)
	}

	results, err := service.Search(
		siteDirectory,
		PropertyFilters{
			Search: "usa",
			Pax:    2,
			Order:  1,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 2 ||
		results[0].BookingURL != sampleURL ||
		results[1].BookingURL != "" {
		t.Fatalf(
			"unexpected booking URLs in search results: %#v",
			results,
		)
	}

	encoded, err := json.Marshal(results)
	if err != nil {
		t.Fatal(err)
	}

	var objects []map[string]json.RawMessage

	if err := json.Unmarshal(encoded, &objects); err != nil {
		t.Fatal(err)
	}

	var receivedURL string

	if err := json.Unmarshal(
		objects[0]["booking_url"],
		&receivedURL,
	); err != nil {
		t.Fatal(err)
	}

	if receivedURL != sampleURL {
		t.Fatalf(
			"booking_url JSON = %q, want %q",
			receivedURL,
			sampleURL,
		)
	}

	if _, found := objects[1]["booking_url"]; found {
		t.Fatal("unlinked property should omit booking_url")
	}
}
