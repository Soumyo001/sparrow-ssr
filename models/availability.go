package models

type AvailabilityWindow struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type PropertyAvailability struct {
	Properties map[string][]AvailabilityWindow `json:"properties"`
}
