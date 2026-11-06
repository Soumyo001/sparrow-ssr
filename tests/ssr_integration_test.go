package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sparrow-ssr/models"
	"sparrow-ssr/routers"
	"sparrow-ssr/services"

	beego "github.com/beego/beego/v2/server/web"
)

// writeFixture saves one JSON document in the test directory.
func writeFixture(t *testing.T, filename string, value any) {
	t.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		t.Fatal(err)
	}
}

// createSiteFixture creates a complete, independent test site.
func createSiteFixture(
	t *testing.T,
	root, site, host string,
) string {
	t.Helper()

	relativeDir := "beta/preview/" + site
	folder := filepath.Join(root, filepath.FromSlash(relativeDir))

	if err := os.MkdirAll(folder, 0755); err != nil {
		t.Fatal(err)
	}

	cdnBase := "https://cdn.123presto.com/" + relativeDir

	writeFixture(t, filepath.Join(folder, "header.json"),
		models.SiteHeader{
			SiteName:   site,
			Language:   "en",
			HeaderHTML: "<header>" + site + " header</header>",
			CSSFile:    cdnBase + "/global.css",
			JSFile:     cdnBase + "/global.js",
		},
	)

	writeFixture(t, filepath.Join(folder, "footer.json"),
		models.SiteFooter{
			FooterHTML: "<footer>" + site + " footer</footer>",
		},
	)

	writeFixture(t, filepath.Join(folder, "home.json"),
		models.PageDocument{
			Head: models.PageHead{
				Title:        site + " Home",
				CanonicalURL: "https://" + host + "/",
			},
			Body: models.PageBody{
				Content: "<main>" + site + " home</main>",
			},
		},
	)

	writeFixture(t, filepath.Join(folder, "details.json"),
		models.PageDocument{
			Head: models.PageHead{
				Title:        site + " Details",
				CanonicalURL: "https://" + host + "/details",
			},
			Body: models.PageBody{
				Content: "<main>" + site + " details</main>",
			},
		},
	)

	metadata := models.DeploymentResponse{
		Data: models.DeploymentData{
			PublishedDomain: "https://" + host + "/",
			UploadedFilePaths: []string{
				cdnBase + "/header.json",
				cdnBase + "/footer.json",
				cdnBase + "/home.json",
				cdnBase + "/details.json",
			},
		},
	}

	metadataFile := filepath.Join(
		t.TempDir(), site+"-deployment.json",
	)

	writeFixture(t, metadataFile, metadata)
	return metadataFile
}

// requestPage passes a real HTTP request through Beego's router.
func requestPage(host, path string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(
		http.MethodGet,
		"http://127.0.0.1"+path,
		nil,
	)
	request.Host = host

	recorder := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(recorder, request)

	return recorder
}

func TestMultiSiteSSRHTTP(t *testing.T) {
	root := t.TempDir()

	alphaFile := createSiteFixture(
		t, root, "alpha", "alpha.example.com",
	)
	bravoFile := createSiteFixture(
		t, root, "bravo", "bravo.example.com",
	)

	catalog, err := services.NewMultiCatalog(
		root,
		[]string{alphaFile, bravoFile},
	)
	if err != nil {
		t.Fatal(err)
	}

	if catalog.Count() != 4 {
		t.Fatalf(
			"expected 4 registered routes, got %d",
			catalog.Count(),
		)
	}

	loader, err := services.NewContentLoader(root)
	if err != nil {
		t.Fatal(err)
	}

	// Create separate property datasets for the two test sites.
	mockRoot := t.TempDir()

	for _, tc := range []struct {
		site string
		id   string
	}{
		{"alpha", "alpha-property"},
		{"bravo", "bravo-property"},
	} {
		folder := filepath.Join(
			mockRoot,
			"beta",
			"preview",
			tc.site,
		)

		if err := os.MkdirAll(folder, 0755); err != nil {
			t.Fatal(err)
		}

		writeFixture(
			t,
			filepath.Join(folder, "properties.json"),
			models.PropertyDataset{
				Properties: []models.Property{
					{
						ID:        tc.id,
						Name:      tc.site + " Property",
						MaxGuests: 4,
						Location: models.PropertyLocation{
							Country: "USA",
						},
						PopularityScore: 90,
					},
				},
			},
		)
	}

	propertySearch, err :=
		services.NewPropertySearchService(mockRoot)

	if err != nil {
		t.Fatal(err)
	}

	// Initialize Beego using the actual project directory.
	// TestBeegoInit loads the application configuration
	// and initializes the default views directory.
	projectRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}

	beego.TestBeegoInit(projectRoot)

	// Register our actual application controller and routes.
	routers.Register(catalog, loader)

	routers.RegisterPropertyAPI(catalog, propertySearch)

	cases := []struct {
		host        string
		path        string
		expected    string
		notExpected string
	}{
		{
			host:        "alpha.example.com",
			path:        "/",
			expected:    "<main>alpha home</main>",
			notExpected: "<main>bravo home</main>",
		},
		{
			host:        "bravo.example.com",
			path:        "/",
			expected:    "<main>bravo home</main>",
			notExpected: "<main>alpha home</main>",
		},
		{
			host:        "alpha.example.com",
			path:        "/details",
			expected:    "<main>alpha details</main>",
			notExpected: "<main>bravo details</main>",
		},
		{
			host:        "bravo.example.com",
			path:        "/details",
			expected:    "<main>bravo details</main>",
			notExpected: "<main>alpha details</main>",
		},
	}

	for _, tc := range cases {
		recorder := requestPage(tc.host, tc.path)

		if recorder.Code != http.StatusOK {
			t.Fatalf(
				"%s%s: expected 200, got %d: %s",
				tc.host, tc.path,
				recorder.Code, recorder.Body.String(),
			)
		}

		if !strings.HasPrefix(
			recorder.Header().Get("Content-Type"),
			"text/html",
		) {
			t.Errorf("%s%s: expected HTML content type",
				tc.host, tc.path)
		}

		body := recorder.Body.String()

		if !strings.Contains(body, tc.expected) {
			t.Errorf("%s%s: expected content missing",
				tc.host, tc.path)
		}

		if strings.Contains(body, tc.notExpected) {
			t.Errorf("%s%s: content mixed between sites",
				tc.host, tc.path)
		}

		site := strings.Split(tc.host, ".")[0]

		if !strings.Contains(body, site+" header") ||
			!strings.Contains(body, site+" footer") {
			t.Errorf("%s%s: incorrect shared HTML",
				tc.host, tc.path)
		}

		cdnBase := "https://cdn.123presto.com/beta/preview/" + site

		if !strings.Contains(body, cdnBase+"/global.css") ||
			!strings.Contains(body, cdnBase+"/global.js") {
			t.Errorf("%s%s: incorrect CDN asset references",
				tc.host, tc.path)
		}

		expectedCanonical := "https://" + tc.host + tc.path
		if !strings.Contains(body, expectedCanonical) {
			t.Errorf("%s%s: canonical URL missing",
				tc.host, tc.path)
		}
	}

	// Two sites, each with one header, footer,
	// homepage and details page = 8 unique files.
	if got := loader.Stats().FileReads; got != 8 {
		t.Errorf(
			"expected 8 file reads, got %d", got,
		)
	}

	// Repeating the requests should reuse cached documents.
	for _, tc := range cases {
		recorder := requestPage(tc.host, tc.path)
		if recorder.Code != http.StatusOK {
			t.Errorf("warm request failed: %s%s",
				tc.host, tc.path)
		}
	}

	if got := loader.Stats().FileReads; got != 8 {
		t.Errorf(
			"warm requests caused extra reads: %d", got,
		)
	}

	// Unknown domains and paths must not resolve.
	for _, tc := range []struct {
		host string
		path string
	}{
		{"unknown.example.com", "/"},
		{"alpha.example.com", "/missing"},
		{"bravo.example.com", "/missing"},
	} {
		recorder := requestPage(tc.host, tc.path)

		if recorder.Code != http.StatusNotFound {
			t.Errorf(
				"%s%s: expected 404, got %d",
				tc.host, tc.path, recorder.Code,
			)
		}
	}

	// Missing routes must not trigger additional content reads.
	if got := loader.Stats().FileReads; got != 8 {
		t.Errorf(
			"unknown routes triggered extra reads: %d", got,
		)
	}

	t.Run("PropertyAPI", func(t *testing.T) {
		cases := []struct {
			host       string
			query      string
			status     int
			total      int
			propertyID string
		}{
			{
				"alpha.example.com",
				"search=usa&pax=2&order=1",
				http.StatusOK,
				1,
				"alpha-property",
			},
			{
				"bravo.example.com",
				"search=usa&pax=2&order=1",
				http.StatusOK,
				1,
				"bravo-property",
			},
			{
				"alpha.example.com",
				"search=atlantis&pax=2&order=1",
				http.StatusOK,
				0,
				"",
			},
			{
				"alpha.example.com",
				"search=usa&pax=0&order=1",
				http.StatusBadRequest,
				0,
				"",
			},
			{
				"alpha.example.com",
				"search=usa&pax=abc&order=1",
				http.StatusBadRequest,
				0,
				"",
			},
			{
				"alpha.example.com",
				"search=usa&pax=2&order=9",
				http.StatusBadRequest,
				0,
				"",
			},
			{
				"unknown.example.com",
				"search=usa&pax=2&order=1",
				http.StatusNotFound,
				0,
				"",
			},
		}

		for _, tc := range cases {
			recorder := requestPage(
				tc.host,
				"/api/properties?"+tc.query,
			)

			if recorder.Code != tc.status {
				t.Errorf(
					"%s?%s: want status %d, got %d: %s",
					tc.host,
					tc.query,
					tc.status,
					recorder.Code,
					recorder.Body.String(),
				)
				continue
			}

			contentType := recorder.Header().Get(
				"Content-Type",
			)

			if !strings.HasPrefix(
				contentType,
				"application/json",
			) {
				t.Errorf(
					"expected JSON content type, got %q",
					contentType,
				)
			}

			if tc.status != http.StatusOK {
				var apiError struct {
					Error string `json:"error"`
				}

				if err := json.Unmarshal(
					recorder.Body.Bytes(),
					&apiError,
				); err != nil {
					t.Errorf(
						"invalid error JSON: %v", err,
					)
				} else if apiError.Error == "" {
					t.Error("expected API error message")
				}

				continue
			}

			var result struct {
				Search     string            `json:"search"`
				Pax        int               `json:"pax"`
				Order      int               `json:"order"`
				Total      int               `json:"total"`
				Properties []models.Property `json:"properties"`
			}

			if err := json.Unmarshal(
				recorder.Body.Bytes(),
				&result,
			); err != nil {
				t.Errorf("invalid response JSON: %v", err)
				continue
			}

			if result.Total != tc.total ||
				len(result.Properties) != tc.total {
				t.Errorf(
					"%s: want %d properties, got %d",
					tc.host,
					tc.total,
					len(result.Properties),
				)
				continue
			}

			if result.Pax != 2 && tc.status == http.StatusOK {
				t.Errorf("unexpected guest count: %d", result.Pax)
			}

			if result.Order != 1 {
				t.Errorf("unexpected sort order: %d", result.Order)
			}

			if tc.total > 0 &&
				result.Properties[0].ID != tc.propertyID {
				t.Errorf(
					"%s: want property %q, got %q",
					tc.host,
					tc.propertyID,
					result.Properties[0].ID,
				)
			}
		}
	})

}
