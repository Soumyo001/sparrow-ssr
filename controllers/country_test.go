package controllers

import "testing"

func TestCountryFromSlug(t *testing.T) {
	for _, tc := range []struct {
		slug, want string
		valid      bool
	}{
		{"usa", "USA", true},
		{"canada", "Canada", true},
		{"united-kingdom", "United Kingdom", true},
		{"USA", "", false},
		{"usa%2fcanada", "", false},
		{"us--ca", "", false},
		{"", "", false},
	} {
		got, valid := countryFromSlug(tc.slug)
		if got != tc.want || valid != tc.valid {
			t.Errorf(
				"countryFromSlug(%q) = (%q, %v), want (%q, %v)",
				tc.slug, got, valid, tc.want, tc.valid,
			)
		}
	}
}

func TestSafeBooking(t *testing.T) {
	if got := safeBooking(
		"Booking.com",
		"https://www.booking.com/hotel/us/example.html",
	); got == "" {
		t.Fatal("expected a valid booking-provider URL")
	}

	for _, raw := range []string{
		"https://booking.com.evil.example/hotel/steal",
		"javascript:alert(1)",
		"https://booking.com/",
		"https://user:pass@booking.com/hotel/example",
	} {
		if got := safeBooking("Booking.com", raw); got != "" {
			t.Errorf("unsafe provider URL was accepted: %q", raw)
		}
	}
}
