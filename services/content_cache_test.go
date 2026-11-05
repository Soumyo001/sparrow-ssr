package services

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"sparrow-ssr/models"
)

const testSiteHost = "test-travel-templates.beta.123presto.com"

func TestContentCacheReusesSharedFiles(t *testing.T) {
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

	home, ok := catalog.Resolve(testSiteHost, "/")
	if !ok {
		t.Fatal("home route not found")
	}

	privacy, ok := catalog.Resolve(
		testSiteHost, "/privacy-policy",
	)
	if !ok {
		t.Fatal("privacy route not found")
	}

	// First page: header, home, footer.
	if _, err := loader.Load(home); err != nil {
		t.Fatal(err)
	}

	if got := loader.Stats().FileReads; got != 3 {
		t.Fatalf(
			"first page: expected 3 reads, got %d", got,
		)
	}

	// Privacy page should reuse header and footer.
	if _, err := loader.Load(privacy); err != nil {
		t.Fatal(err)
	}

	if got := loader.Stats().FileReads; got != 4 {
		t.Fatalf(
			"second page: expected 4 total reads, got %d",
			got,
		)
	}

	// Repeated home request should require zero new reads.
	if _, err := loader.Load(home); err != nil {
		t.Fatal(err)
	}

	if got := loader.Stats().FileReads; got != 4 {
		t.Fatalf(
			"warm page performed extra read: %d", got,
		)
	}
}

func TestContentCacheConcurrentFirstAccess(t *testing.T) {
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

	home, ok := catalog.Resolve(testSiteHost, "/")
	if !ok {
		t.Fatal("home route not found")
	}

	const requests = 50

	errors := make(chan error, requests)
	var wg sync.WaitGroup

	for i := 0; i < requests; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			content, err := loader.Load(home)
			if err != nil {
				errors <- err
				return
			}

			if content.Page.Body.Content == "" {
				errors <- fmt.Errorf("empty page body")
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Error(err)
	}

	// All 50 requests should share just 3 file loads.
	if got := loader.Stats().FileReads; got != 3 {
		t.Fatalf(
			"expected 3 file reads for 50 requests, got %d",
			got,
		)
	}
}

func TestContentCacheInvalidatePage(t *testing.T) {
	root := t.TempDir()

	folder := filepath.Join(
		root, "beta", "preview", "example",
	)

	if err := os.MkdirAll(folder, 0755); err != nil {
		t.Fatal(err)
	}

	write := func(name, data string) {
		t.Helper()

		err := os.WriteFile(
			filepath.Join(folder, name),
			[]byte(data),
			0644,
		)
		if err != nil {
			t.Fatal(err)
		}
	}

	write("header.json",
		`{"header":"<header>Header</header>"}`)

	write("footer.json",
		`{"footer":"<footer>Footer</footer>"}`)

	write("home.json",
		`{"head":{"canonical_url":"https://example.com"},"body":{"content":"before"}}`)

	location := models.PageLocation{
		HeaderJSON: "beta/preview/example/header.json",
		PageJSON:   "beta/preview/example/home.json",
		FooterJSON: "beta/preview/example/footer.json",
	}

	loader, err := NewContentLoader(root)
	if err != nil {
		t.Fatal(err)
	}

	first, err := loader.Load(location)
	if err != nil {
		t.Fatal(err)
	}

	if first.Page.Body.Content != "before" {
		t.Fatal("incorrect initial content")
	}

	// Simulate content being updated after deployment.
	write("home.json",
		`{"head":{"canonical_url":"https://example.com"},"body":{"content":"after"}}`)

	// Without invalidation, cached content remains unchanged.
	stale, err := loader.Load(location)
	if err != nil {
		t.Fatal(err)
	}

	if stale.Page.Body.Content != "before" {
		t.Fatal("unexpected cache refresh")
	}

	// Invalidate only the modified page.
	if err := loader.Invalidate(
		location.PageJSON,
	); err != nil {
		t.Fatal(err)
	}

	updated, err := loader.Load(location)
	if err != nil {
		t.Fatal(err)
	}

	if updated.Page.Body.Content != "after" {
		t.Fatal("invalidation did not reload content")
	}

	if got := loader.Stats().FileReads; got != 4 {
		t.Fatalf(
			"expected 4 total reads, got %d", got,
		)
	}
}

func TestContentLoaderRejectsTraversal(t *testing.T) {
	loader, err := NewContentLoader("../content")
	if err != nil {
		t.Fatal(err)
	}

	_, err = loader.Load(models.PageLocation{
		HeaderJSON: "../outside.json",
		PageJSON: "beta/preview/" +
			"test-travel-templates/home.json",
		FooterJSON: "beta/preview/" +
			"test-travel-templates/footer.json",
	})

	if err == nil {
		t.Fatal("expected path validation error")
	}

	if err := loader.Invalidate(
		"../outside.json",
	); err == nil {
		t.Fatal("expected invalidation to reject traversal")
	}
}
