package services

import (
	"strings"
	"testing"
)

func TestContentLoader(t *testing.T) {
	catalog, err := NewCatalog(
		"../content",
		"../conf/deployment.json",
	)
	if err != nil {
		t.Fatal(err)
	}

	loader, err := NewContentLoader("../content")
	if err != nil {
		t.Fatal(err)
	}

	location, found := catalog.Resolve(
		"test-travel-templates.beta.123presto.com",
		"/",
	)
	if !found {
		t.Fatal("homepage route not found")
	}

	content, err := loader.Load(location)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(
		content.Header.HeaderHTML,
		"<header",
	) {
		t.Error("header content not loaded")
	}

	if !strings.Contains(
		content.Page.Body.Content,
		"Experience Mexico",
	) {
		t.Error("homepage body not loaded")
	}

	if !strings.Contains(
		content.Footer.FooterHTML,
		"<footer",
	) {
		t.Error("footer content not loaded")
	}

	if content.Page.Head.CSSFile == nil {
		t.Error("homepage CSS reference missing")
	}

	if content.Header.JSFile == "" {
		t.Error("shared JavaScript reference missing")
	}
}
