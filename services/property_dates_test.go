package services

import "testing"

func TestPropertySearchDateAvailability(t *testing.T) {
	service, err := NewPropertySearchService("../mock_data")
	if err != nil {
		t.Fatal(err)
	}

	const site = "beta/preview/test-travel-templates"

	for _, tc := range []struct {
		start string
		end   string
		want  int
	}{
		{"2026-11-06", "2026-11-13", 10},
		{"2026-11-16", "2026-11-20", 15},
		{"2027-12-30", "2028-01-01", 0},
	} {
		got, err := service.Search(
			site,
			PropertyFilters{
				Search:    "usa",
				Pax:       2,
				Order:     1,
				DateStart: tc.start,
				DateEnd:   tc.end,
			},
		)

		if err != nil {
			t.Fatal(err)
		}

		if len(got) != tc.want {
			t.Fatalf(
				"%s to %s: want %d, got %d",
				tc.start, tc.end,
				tc.want, len(got),
			)
		}
	}

	// Invalid and incomplete date ranges.
	for _, tc := range []struct {
		start string
		end   string
	}{
		{"2026-11-06", ""},
		{"", "2026-11-13"},
		{"2026-11-13", "2026-11-06"},
		{"2026-02-30", "2026-03-03"},
	} {
		_, err := service.Search(
			site,
			PropertyFilters{
				Search:    "usa",
				Pax:       2,
				Order:     1,
				DateStart: tc.start,
				DateEnd:   tc.end,
			},
		)

		if err == nil {
			t.Errorf(
				"expected invalid date error: %+v",
				tc,
			)
		}
	}
}
