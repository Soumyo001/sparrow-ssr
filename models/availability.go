package models

import "time"

type AvailabilityWindow struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type PropertyAvailability struct {
	Properties map[string][]AvailabilityWindow `json:"properties"`
}

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
