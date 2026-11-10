package main

import (
	"testing"

	"sparrow-ssr/models"
)

func TestSeedData(t *testing.T) {
	if err := validateSeedData(seedProperties); err != nil {
		t.Fatalf("invalid seed data: %v", err)
	}
	if got := len(seedProperties); got != 15 {
		t.Fatalf("want 15 properties, got %d", got)
	}
	locations := make(map[models.PropertyLocation]bool)
	ids := make(map[string]bool)
	for _, property := range seedProperties {
		locations[property.Location] = true
		ids[property.ID] = true
	}
	if got := len(locations); got != 8 {
		t.Fatalf("want 8 distinct locations, got %d", got)
	}
	if !ids["prop-001"] || !ids["prop-015"] {
		t.Fatal("expected first and last property IDs")
	}
}

func TestRejectDuplicateSeedID(t *testing.T) {
	duplicate := append([]seedProperty(nil), seedProperties...)
	duplicate = append(duplicate, seedProperties[0])
	if err := validateSeedData(duplicate); err == nil {
		t.Fatal("expected duplicate ID to be rejected")
	}
}
