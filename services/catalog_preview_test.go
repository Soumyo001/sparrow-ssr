package services

import "testing"

func TestRegisterPreviewAlias(t *testing.T) {
	root := t.TempDir()

	alphaFile := testDeployment(
		t, root, "alpha", "alpha.example.com",
	)
	bravoFile := testDeployment(
		t, root, "bravo", "bravo.example.com",
	)

	catalog, err := NewMultiCatalog(
		root,
		[]string{alphaFile, bravoFile},
	)
	if err != nil {
		t.Fatal(err)
	}

	before := catalog.Count()

	err = catalog.RegisterPreviewAlias(
		"alpha.example.com",
		"sparrow-demo.onrender.com",
	)
	if err != nil {
		t.Fatal(err)
	}

	if catalog.Count() != before+1 {
		t.Fatalf(
			"expected one additional route, got %d",
			catalog.Count(),
		)
	}

	original, found := catalog.Resolve(
		"alpha.example.com", "/",
	)
	if !found {
		t.Fatal("original publisher route disappeared")
	}

	preview, found := catalog.Resolve(
		"sparrow-demo.onrender.com", "/",
	)
	if !found {
		t.Fatal("preview route was not registered")
	}

	if preview != original {
		t.Fatalf(
			"preview location changed publisher identity: %+v",
			preview,
		)
	}

	if _, found := catalog.Resolve(
		"bravo.example.com", "/",
	); !found {
		t.Fatal("second publisher route disappeared")
	}

	for _, tc := range []struct {
		host string
		path string
	}{
		{"unknown.example.com", "/"},
		{"sparrow-demo.onrender.com", "/missing"},
	} {
		if _, found := catalog.Resolve(
			tc.host, tc.path,
		); found {
			t.Fatalf(
				"unexpected route for %s%s",
				tc.host, tc.path,
			)
		}
	}

	if err := catalog.RegisterPreviewAlias(
		"bravo.example.com",
		"sparrow-demo.onrender.com",
	); err == nil {
		t.Fatal("duplicate preview Host was accepted")
	}

	if err := catalog.RegisterPreviewAlias(
		"unknown.example.com",
		"another-demo.onrender.com",
	); err == nil {
		t.Fatal("unregistered source Host was accepted")
	}

	if err := catalog.RegisterPreviewAlias(
		"alpha.example.com",
		"https://invalid.example.com",
	); err == nil {
		t.Fatal("invalid preview hostname was accepted")
	}
}
