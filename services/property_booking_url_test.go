package services

import (
	"encoding/json"
	"testing"

	"sparrow-ssr/models"
)

// Check the public response's booking_url contract without using mock JSON files.
func TestPropertyBookingURLRoundTrip(t *testing.T) {
	const sampleURL = "https://booking.example.invalid/hotel/fixture"

	properties := []models.Property{
		{
			ID:           "linked",
			ProviderName: "Booking.com",
			BookingURL:   sampleURL,
		},
		{
			ID:           "unlinked",
			ProviderName: "Expedia",
		},
	}

	encoded, err := json.Marshal(properties)
	if err != nil {
		t.Fatal(err)
	}

	var objects []map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &objects); err != nil {
		t.Fatal(err)
	}

	if len(objects) != 2 {
		t.Fatalf("want 2 properties, got %d", len(objects))
	}

	var receivedURL string
	if err := json.Unmarshal(objects[0]["booking_url"], &receivedURL); err != nil {
		t.Fatal(err)
	}

	if receivedURL != sampleURL {
		t.Fatalf("booking_url JSON = %q, want %q", receivedURL, sampleURL)
	}

	if _, found := objects[1]["booking_url"]; found {
		t.Fatal("unlinked property should omit booking_url")
	}
}
