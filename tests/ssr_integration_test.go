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

	// Exercise site selection and API handling without a database connection.
	// The database search itself is tested separately against Neon.
	propertySearch := fixturePropertySearcher{
		"beta/preview/alpha": {
			property: models.Property{
				ID:              "alpha-property",
				Name:            "alpha Property",
				MaxGuests:       4,
				Location:        models.PropertyLocation{Country: "USA"},
				PopularityScore: 90,
			},
			availableFrom: "2026-11-01",
		},
		"beta/preview/bravo": {
			property: models.Property{
				ID:              "bravo-property",
				Name:            "bravo Property",
				MaxGuests:       4,
				Location:        models.PropertyLocation{Country: "USA"},
				PopularityScore: 90,
			},
			availableFrom: "2026-11-15",
		},
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

	t.Run("Health", func(t *testing.T) {
		// Render may call /health using an unregistered
		// onrender.com hostname, so the health endpoint
		// must not require a published-site Host.
		for _, host := range []string{
			"alpha.example.com",
			"unknown.example.com",
		} {
			recorder := requestPage(host, "/health")

			if recorder.Code != http.StatusOK {
				t.Fatalf(
					"%s/health: expected 200, got %d: %s",
					host,
					recorder.Code,
					recorder.Body.String(),
				)
			}

			if !strings.HasPrefix(
				recorder.Header().Get("Content-Type"),
				"application/json",
			) {
				t.Errorf(
					"%s: expected JSON content type",
					host,
				)
			}

			if recorder.Header().Get("Cache-Control") != "no-store" {
				t.Errorf(
					"%s: expected Cache-Control: no-store",
					host,
				)
			}

			var health struct {
				Status   string `json:"status"`
				API      string `json:"api"`
				Frontend string `json:"frontend"`
			}

			if err := json.Unmarshal(
				recorder.Body.Bytes(),
				&health,
			); err != nil {
				t.Fatal(err)
			}

			if health.Status != "ok" ||
				health.API != "ready" ||
				health.Frontend != "ready" {
				t.Errorf(
					"%s: unexpected health response: %+v",
					host,
					health,
				)
			}
		}
	})

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

		for _, asset := range []string{
			"/static/css/property_search_popup.css",
			"/static/js/property_search_popup.js",
		} {
			if !strings.Contains(body, asset) {
				t.Errorf(
					"%s%s: missing search enhancement asset %q",
					tc.host,
					tc.path,
					asset,
				)
			}
		}

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

	t.Run("PropertyAPIDates", func(t *testing.T) {
		cases := []struct {
			host   string
			query  string
			status int
			total  int
		}{
			{
				"alpha.example.com",
				"search=usa&pax=2&order=1&dateStart=2026-11-06&dateEnd=2026-11-13",
				http.StatusOK,
				1,
			},
			{
				"bravo.example.com",
				"search=usa&pax=2&order=1&dateStart=2026-11-06&dateEnd=2026-11-13",
				http.StatusOK,
				0,
			},
			{
				"bravo.example.com",
				"search=usa&pax=2&order=1&dateStart=2026-11-16&dateEnd=2026-11-20",
				http.StatusOK,
				1,
			},
			{
				"alpha.example.com",
				"search=usa&pax=2&order=1&dateStart=2026-11-13&dateEnd=2026-11-06",
				http.StatusBadRequest,
				0,
			},
			{
				"alpha.example.com",
				"search=usa&pax=2&order=1&dateStart=2026-11-06",
				http.StatusBadRequest,
				0,
			},
		}

		for _, tc := range cases {
			recorder := requestPage(
				tc.host,
				"/api/properties?"+tc.query,
			)

			if recorder.Code != tc.status {
				t.Errorf(
					"%s: expected %d, got %d: %s",
					tc.host,
					tc.status,
					recorder.Code,
					recorder.Body.String(),
				)
				continue
			}

			if !strings.HasPrefix(
				recorder.Header().Get("Content-Type"),
				"application/json",
			) {
				t.Error("expected application/json")
			}

			if tc.status != http.StatusOK {
				var apiError struct {
					Error string `json:"error"`
				}

				if err := json.Unmarshal(
					recorder.Body.Bytes(),
					&apiError,
				); err != nil || apiError.Error == "" {
					t.Errorf(
						"invalid API error response: %v",
						err,
					)
				}
				continue
			}

			var result struct {
				Total      int               `json:"total"`
				Properties []models.Property `json:"properties"`
			}

			if err := json.Unmarshal(
				recorder.Body.Bytes(),
				&result,
			); err != nil {
				t.Fatal(err)
			}

			if result.Total != tc.total ||
				len(result.Properties) != tc.total {
				t.Errorf(
					"%s: expected %d properties, got %d",
					tc.host,
					tc.total,
					len(result.Properties),
				)
			}
		}
	})

	t.Run("RefineClientPage", func(t *testing.T) {
		cases := []struct {
			host        string
			expected    string
			notExpected string
		}{
			{
				host:        "alpha.example.com",
				expected:    "alpha header",
				notExpected: "bravo header",
			},
			{
				host:        "bravo.example.com",
				expected:    "bravo header",
				notExpected: "alpha header",
			},
		}

		for _, tc := range cases {
			recorder := requestPage(
				tc.host,
				"/refine?search=usa&pax=2&order=1",
			)

			if recorder.Code != http.StatusOK {
				t.Fatalf(
					"%s/refine: expected 200, got %d: %s",
					tc.host,
					recorder.Code,
					recorder.Body.String(),
				)
			}

			if !strings.HasPrefix(
				recorder.Header().Get("Content-Type"),
				"text/html",
			) {
				t.Errorf(
					"%s: expected HTML response",
					tc.host,
				)
			}

			body := recorder.Body.String()

			required := []string{
				`id="refine-app"`,
				`id="refine-grid"`,
				`/static/css/refine.css`,
				`/static/js/refine.js`,
				`/static/css/property_search_popup.css`,
				`/static/js/property_search_popup.js`,
				"https://" + tc.host + "/refine",
				tc.expected,
			}

			for _, fragment := range required {
				if !strings.Contains(body, fragment) {
					t.Errorf(
						"%s: missing %q",
						tc.host,
						fragment,
					)
				}
			}

			if strings.Contains(body, tc.notExpected) {
				t.Errorf(
					"%s: another site's header leaked",
					tc.host,
				)
			}
		}

		// The virtual route must not bypass hostname checks.
		unknown := requestPage(
			"unknown.example.com",
			"/refine?search=usa&pax=2&order=1",
		)

		if unknown.Code != http.StatusNotFound {
			t.Errorf(
				"unknown host: expected 404, got %d",
				unknown.Code,
			)
		}
	})

	t.Run("CountrySSR", func(t *testing.T) {
		for _, tc := range []struct {
			host, path   string
			wantStatus   int
			want, reject string
		}{
			{
				"alpha.example.com", "/all/usa",
				http.StatusOK,
				"alpha Property", "bravo Property",
			},
			{
				"bravo.example.com", "/all/usa",
				http.StatusOK,
				"bravo Property", "alpha Property",
			},
			{
				"alpha.example.com", "/all/canada",
				http.StatusOK,
				"No properties are currently listed",
				"alpha Property",
			},
			{
				"unknown.example.com", "/all/usa",
				http.StatusNotFound,
				"Site not found", "alpha Property",
			},
			{
				"alpha.example.com", "/all/USA",
				http.StatusNotFound,
				"Country not found", "alpha Property",
			},
		} {
			response := requestPage(tc.host, tc.path)

			if response.Code != tc.wantStatus {
				t.Errorf(
					"%s%s: HTTP %d, want %d: %s",
					tc.host, tc.path,
					response.Code, tc.wantStatus,
					response.Body.String(),
				)
				continue
			}

			body := response.Body.String()

			if !strings.Contains(body, tc.want) ||
				strings.Contains(body, tc.reject) {
				t.Errorf(
					"%s%s: wrong country/site content",
					tc.host, tc.path,
				)
			}

			if tc.wantStatus == http.StatusOK {
				if !strings.Contains(
					body, `id="country-results"`,
				) ||
					!strings.Contains(
						body,
						"https://"+tc.host+tc.path,
					) ||
					!strings.Contains(
						body, `/static/js/refine.js`,
					) ||
					!strings.Contains(
						body, `id="refine-pagination"`,
					) ||
					!strings.Contains(
						body, `data-filter-open="dates"`,
					) {
					t.Errorf(
						"%s%s: expected SSR country page with shared filters",
						tc.host, tc.path,
					)
				}
			}
		}
	})

}
