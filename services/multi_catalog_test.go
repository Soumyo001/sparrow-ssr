package services

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sparrow-ssr/models"
)

func writeMultiSiteJSON(t *testing.T, filename string, value any) {
	t.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filename, data, 0644); err != nil {
		t.Fatal(err)
	}
}

// testDeployment builds an isolated site fixture using the existing storage layout.
func testDeployment(t *testing.T, contentRoot, site, host string) string {
	t.Helper()

	relativeDir := "beta/preview/" + site
	folder := filepath.Join(contentRoot, filepath.FromSlash(relativeDir))
	if err := os.MkdirAll(folder, 0755); err != nil {
		t.Fatal(err)
	}

	writeMultiSiteJSON(t, filepath.Join(folder, "header.json"), models.SiteHeader{
		HeaderHTML: "<header>" + site + "</header>",
	})
	writeMultiSiteJSON(t, filepath.Join(folder, "footer.json"), models.SiteFooter{
		FooterHTML: "<footer>" + site + "</footer>",
	})
	writeMultiSiteJSON(t, filepath.Join(folder, "home.json"), models.PageDocument{
		Head: models.PageHead{CanonicalURL: "https://" + host + "/"},
		Body: models.PageBody{Content: "<main>" + site + "</main>"},
	})

	cdnBase := "https://cdn.123presto.com/" + relativeDir
	metadata := models.DeploymentResponse{
		Data: models.DeploymentData{
			PublishedDomain: "https://" + host + "/",
			UploadedFilePaths: []string{
				cdnBase + "/header.json",
				cdnBase + "/footer.json",
				cdnBase + "/home.json",
			},
		},
	}

	metadataFile := filepath.Join(t.TempDir(), "deployment.json")
	writeMultiSiteJSON(t, metadataFile, metadata)
	return metadataFile
}

func TestMultiCatalogResolvesDistinctSites(t *testing.T) {
	root := t.TempDir()
	travelFile := testDeployment(t, root, "travel-demo", "travel.example.com")
	hotelFile := testDeployment(t, root, "hotel-demo", "hotel.example.com")

	catalog, err := NewMultiCatalog(root, []string{travelFile, hotelFile})
	if err != nil {
		t.Fatal(err)
	}
	if catalog.Count() != 2 {
		t.Fatalf("expected 2 routes, got %d", catalog.Count())
	}

	loader, err := NewContentLoader(root)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		host string
		site string
	}{
		{"travel.example.com", "travel-demo"},
		{"hotel.example.com", "hotel-demo"},
	} {
		location, found := catalog.Resolve(tc.host, "/")
		if !found {
			t.Fatalf("no homepage route for %q", tc.host)
		}

		content, err := loader.Load(location)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(content.Header.HeaderHTML, tc.site) ||
			!strings.Contains(content.Page.Body.Content, tc.site) ||
			!strings.Contains(content.Footer.FooterHTML, tc.site) {
			t.Fatalf("content mixed between sites for %q", tc.host)
		}
	}

	if _, found := catalog.Resolve("unknown.example.com", "/"); found {
		t.Fatal("unknown domain should not resolve")
	}
	if _, found := catalog.Resolve("travel.example.com", "/missing"); found {
		t.Fatal("unknown path should not resolve")
	}
}

func TestMultiCatalogRejectsDuplicateRoute(t *testing.T) {
	root := t.TempDir()
	file := testDeployment(t, root, "travel-demo", "travel.example.com")

	catalog, err := NewMultiCatalog(root, []string{file, file})
	if err == nil || catalog != nil {
		t.Fatal("expected duplicate route to reject the catalog")
	}
	if !strings.Contains(err.Error(), "duplicate route") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMultiCatalogRejectsInvalidInput(t *testing.T) {
	root := t.TempDir()

	for _, files := range [][]string{
		nil,
		{""},
		{"missing-deployment.json"},
	} {
		catalog, err := NewMultiCatalog(root, files)
		if err == nil || catalog != nil {
			t.Fatalf("expected error for metadata files %q", files)
		}
	}
}
