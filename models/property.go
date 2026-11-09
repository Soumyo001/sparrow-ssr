package models

type PropertyLocation struct {
	City    string `json:"city"`
	State   string `json:"state"`
	Area    string `json:"area"`
	Country string `json:"country"`
}

type Property struct {
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	PropertyType    string           `json:"property_type"`
	Location        PropertyLocation `json:"location"`
	MaxGuests       int              `json:"max_guests"`
	DemoMaxGuests   int              `json:"demo_max_guests,omitempty"`
	SizeSqFt        int              `json:"size_sqft"`
	Rating          float64          `json:"rating"`
	ReviewCount     int              `json:"review_count"`
	PriceFrom       int              `json:"price_from"`
	DemoPriceFrom   int              `json:"demo_price_from,omitempty"`
	Currency        string           `json:"currency"`
	Amenities       []string         `json:"amenities"`
	ProviderName    string           `json:"provider_name"`
	BookingURL      string           `json:"booking_url,omitempty"`
	PopularityScore int              `json:"popularity_score"`
	EcoFriendly     bool             `json:"eco_friendly"`
	Images          []string         `json:"images"`
}

type PropertyDataset struct {
	Properties []Property `json:"properties"`
}
