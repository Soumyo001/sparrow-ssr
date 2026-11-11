package services

import (
	"fmt"
	"time"
)

// PropertyFilters contains the filters accepted by the property search API.
type PropertyFilters struct {
	Search    string
	Pax       int
	Order     int
	DateStart string
	DateEnd   string
}

// ParseStayDates accepts either no dates or a valid [check-in, checkout) range.
func ParseStayDates(start, end string) (time.Time, time.Time, error) {
	if start == "" && end == "" {
		return time.Time{}, time.Time{}, nil
	}
	if start == "" || end == "" {
		return time.Time{}, time.Time{}, fmt.Errorf("dateStart and dateEnd must both be provided")
	}
	checkIn, err := time.Parse("2006-01-02", start)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("dateStart must be YYYY-MM-DD")
	}
	checkOut, err := time.Parse("2006-01-02", end)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("dateEnd must be YYYY-MM-DD")
	}
	if !checkOut.After(checkIn) {
		return time.Time{}, time.Time{}, fmt.Errorf("dateEnd must be after dateStart")
	}
	return checkIn, checkOut, nil
}
