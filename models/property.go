package models

// PropertyLocation is the public API representation of a location.
// ORM location records are shared by multiple properties.
type PropertyLocation struct {
	City    string `json:"city"`
	State   string `json:"state"`
	Area    string `json:"area"`
	Country string `json:"country"`
}

// Property is the API response shape consumed by static/js/refine.js.
// Do not register this struct with Beego ORM.
type Property struct {
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	PropertyType    string           `json:"property_type"`
	Location        PropertyLocation `json:"location"`
	MaxGuests       int              `json:"max_guests"`
	DemoMaxGuests   int              `json:"demo_max_guests,omitempty"` // Transitional: remove when the API switches to PostgreSQL.
	SizeSqFt        int              `json:"size_sqft"`
	Rating          float64          `json:"rating"`
	ReviewCount     int              `json:"review_count"`
	PriceFrom       int              `json:"price_from"`
	DemoPriceFrom   int              `json:"demo_price_from,omitempty"` // Transitional: remove when the API switches to PostgreSQL.
	Currency        string           `json:"currency"`
	Amenities       []string         `json:"amenities"`
	ProviderName    string           `json:"provider_name"`
	BookingURL      string           `json:"booking_url,omitempty"`
	PopularityScore int              `json:"popularity_score"`
	EcoFriendly     bool             `json:"eco_friendly"`
	Images          []string         `json:"images"`
}

// PropertyDataset exists only for the old JSON-backed search service and
// its tests. It will be removed during the database cutover.
type PropertyDataset struct {
	Properties []Property `json:"properties"`
}

// SiteRecord identifies one deployment's property catalog. Key is the
// deployment-derived ContentDirectory, never a guessed hostname or path.
type SiteRecord struct {
	Id  int64  `orm:"auto"`
	Key string `orm:"size(255);unique"`
}

func (*SiteRecord) TableName() string { return "sites" }

// LocationRecord stores a reusable city/state/area/country combination.
type LocationRecord struct {
	Id      int64  `orm:"auto"`
	City    string `orm:"size(128)"`
	State   string `orm:"size(128)"`
	Area    string `orm:"size(128)"`
	Country string `orm:"size(128)"`
}

func (*LocationRecord) TableName() string { return "locations" }
func (*LocationRecord) TableUnique() [][]string {
	return [][]string{
		{"Country", "State", "City", "Area"},
	}
}

type PropertyRecord struct {
	Id              int64           `orm:"auto"`
	Site            *SiteRecord     `orm:"rel(fk);on_delete(cascade)"`
	Location        *LocationRecord `orm:"rel(fk);on_delete(cascade)"`
	PropertyID      string          `orm:"column(property_id);size(100)"`
	Name            string          `orm:"size(512)"`
	PropertyType    string          `orm:"size(100)"`
	MaxGuests       int
	SizeSqFt        int `orm:"column(size_sqft)"`
	Rating          float64
	ReviewCount     int
	PriceFrom       int
	Currency        string `orm:"size(12)"`
	AmenitiesJSON   string `orm:"column(amenities);type(jsonb)"`
	ProviderName    string `orm:"size(100)"`
	BookingURL      string `orm:"column(booking_url);type(text)"`
	PopularityScore int
	EcoFriendly     bool
	ImageJSON       string `orm:"column(images);type(jsonb)"`
}

func (*PropertyRecord) TableName() string { return "properties" }
func (*PropertyRecord) TableUnique() [][]string {
	return [][]string{{"Site", "PropertyID"}}
}
func (*PropertyRecord) TableIndex() [][]string {
	return [][]string{
		{"Site", "MaxGuests"},
		{"Site", "PopularityScore"},
	}
}
