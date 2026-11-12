package models

import "time"

// AvailabilityRecord links available [check-in, checkout) windows to
// the property's database primary key, not its public string identifier.
type AvailabilityRecord struct {
	Id        int64           `orm:"auto"`
	Property  *PropertyRecord `orm:"rel(fk);on_delete(cascade)"`
	StartDate time.Time       `orm:"type(date)"`
	EndDate   time.Time       `orm:"type(date)"`
}

func (*AvailabilityRecord) TableName() string { return "property_availability" }

func (*AvailabilityRecord) TableUnique() [][]string {
	return [][]string{
		{"Property", "StartDate", "EndDate"},
	}
}

func (*AvailabilityRecord) TableIndex() [][]string {
	return [][]string{
		{"StartDate", "EndDate"},
	}
}
