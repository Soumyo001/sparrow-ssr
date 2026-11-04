package services

import "testing"

func TestCatalogFromUploadedData(t *testing.T) {
	catalog, err := NewCatalog(
		"../content",
		"../conf/deployment.json",
	)
	if err != nil {
		t.Fatal(err)
	}

	if catalog.Count() != 4 {
		t.Fatalf(
			"expected 4 pages, got %d",
			catalog.Count(),
		)
	}

	host := "test-travel-templates.beta.123presto.com"

	tests := []struct {
		urlPath  string
		jsonFile string
	}{
		{"/", "home.json"},
		{"/privacy-policy", "privacy-policy.json"},
		{"/site-terms", "site-terms.json"},
		{
			"/blog/untitled-post-wrurmp",
			"untitled-post-wrurmp.json",
		},
	}

	for _, test := range tests {
		got, found := catalog.Resolve(
			host, test.urlPath,
		)

		expected := "beta/preview/test-travel-templates/" +
			test.jsonFile

		if !found || got.PageJSON != expected {
			t.Errorf(
				"incorrect resolution for %q: %+v, found %v",
				test.urlPath, got, found,
			)
		}
	}

	for _, request := range []struct {
		host    string
		urlPath string
	}{
		{"other.example.com", "/"},
		{host, "/missing"},
		{host, "/blog"},
	} {
		if _, found := catalog.Resolve(
			request.host, request.urlPath,
		); found {
			t.Errorf(
				"unexpected resolution for %s%s",
				request.host,
				request.urlPath,
			)
		}
	}
}
