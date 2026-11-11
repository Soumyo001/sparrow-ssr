package services

import "testing"

func TestParseStayDates(t *testing.T) {
	for _, tc := range []struct {
		name      string
		start     string
		end       string
		wantError bool
	}{
		{"no_dates", "", "", false},
		{"valid_dates", "2026-11-06", "2026-11-13", false},
		{"same_day", "2026-11-06", "2026-11-06", true},
		{"missing_start", "", "2026-11-13", true},
		{"missing_end", "2026-11-06", "", true},
		{"reversed_dates", "2026-11-13", "2026-11-06", true},
		{"invalid_start", "2026-02-30", "2026-03-03", true},
		{"invalid_end", "2026-11-06", "invalid", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			start, end, err := ParseStayDates(tc.start, tc.end)
			if (err != nil) != tc.wantError {
				t.Fatalf("ParseStayDates(%q, %q): error=%v, wantError=%t", tc.start, tc.end, err, tc.wantError)
			}
			if err == nil {
				if tc.start == "" && (!start.IsZero() || !end.IsZero()) {
					t.Fatal("empty dates must produce zero times")
				}
				if tc.start != "" && !end.After(start) {
					t.Fatal("checkout must follow check-in")
				}
			}
		})
	}
}
