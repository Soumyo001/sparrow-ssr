package services

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"sparrow-ssr/models"
)

// ContentLoader reads and caches decoded JSON documents.
type ContentLoader struct {
	root      string
	headers   *fileCache[models.SiteHeader]
	pages     *fileCache[models.PageDocument]
	footers   *fileCache[models.SiteFooter]
	fileReads atomic.Uint64
}

// LoaderStats provides basic cache instrumentation.
// FileReads counts attempted file reads and JSON decodes.
type LoaderStats struct {
	FileReads     uint64
	CachedHeaders int
	CachedPages   int
	CachedFooters int
}

func NewContentLoader(
	contentRoot string,
) (*ContentLoader, error) {
	root, err := filepath.Abs(contentRoot)
	if err != nil {
		return nil, fmt.Errorf(
			"resolve content root: %w", err,
		)
	}

	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf(
			"stat content root: %w", err,
		)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf(
			"content root is not a directory: %s", root,
		)
	}

	return &ContentLoader{
		root: root,

		headers: newFileCache[models.SiteHeader](),
		pages:   newFileCache[models.PageDocument](),
		footers: newFileCache[models.SiteFooter](),
	}, nil
}

// Load returns the decoded documents required for one page.
func (l *ContentLoader) Load(
	location models.PageLocation,
) (models.PageContent, error) {

	// Fast path: try the existing decoded cache first.
	header, headerReady := l.headers.getReady(
		location.HeaderJSON,
	)

	page, pageReady := l.pages.getReady(
		location.PageJSON,
	)

	footer, footerReady := l.footers.getReady(
		location.FooterJSON,
	)

	if headerReady && pageReady && footerReady {
		return models.PageContent{
			Header: header,
			Page:   page,
			Footer: footer,
		}, nil
	}

	// At least one file is missing or still loading.
	return l.loadConcurrent(location, true)
}

// loadUncached retains the original loading behavior
// for performance comparison in benchmarks.
func (l *ContentLoader) loadUncached(
	location models.PageLocation,
) (models.PageContent, error) {
	return l.loadConcurrent(location, false)
}

// loadConcurrent loads the independent documents in parallel.
func (l *ContentLoader) loadConcurrent(
	location models.PageLocation,
	useCache bool,
) (models.PageContent, error) {
	var content models.PageContent
	var headerErr, pageErr, footerErr error
	var wg sync.WaitGroup

	wg.Add(3)

	go func() {
		defer wg.Done()

		content.Header, headerErr = readContent(
			l,
			l.headers,
			location.HeaderJSON,
			useCache,
		)
	}()

	go func() {
		defer wg.Done()

		content.Page, pageErr = readContent(
			l,
			l.pages,
			location.PageJSON,
			useCache,
		)
	}()

	go func() {
		defer wg.Done()

		content.Footer, footerErr = readContent(
			l,
			l.footers,
			location.FooterJSON,
			useCache,
		)
	}()

	wg.Wait()

	if headerErr != nil {
		return models.PageContent{},
			fmt.Errorf("load header: %w", headerErr)
	}

	if pageErr != nil {
		return models.PageContent{},
			fmt.Errorf("load page: %w", pageErr)
	}

	if footerErr != nil {
		return models.PageContent{},
			fmt.Errorf("load footer: %w", footerErr)
	}

	return content, nil
}

// readContent supports all three JSON model types.
func readContent[T any](
	l *ContentLoader,
	cache *fileCache[T],
	relative string,
	useCache bool,
) (T, error) {
	var zero T

	local := filepath.FromSlash(relative)

	// Reject absolute paths and traversal outside content root.
	if !filepath.IsLocal(local) ||
		filepath.Clean(local) != local {
		return zero, fmt.Errorf(
			"invalid relative content path %q", relative,
		)
	}

	fetch := func() (T, error) {
		var value T

		l.fileReads.Add(1)

		// Reuse readJSON from services/catalog.go.
		err := readJSON(l.root, local, &value)

		return value, err
	}

	if !useCache {
		return fetch()
	}

	return cache.getOrLoad(local, fetch)
}

// Invalidate removes selected cached files.
// In-flight requests may still finish with their previous values.
func (l *ContentLoader) Invalidate(
	relativePaths ...string,
) error {
	validated := make([]string, 0, len(relativePaths))

	for _, relative := range relativePaths {
		local := filepath.FromSlash(relative)

		if !filepath.IsLocal(local) ||
			filepath.Clean(local) != local {
			return fmt.Errorf(
				"invalid relative content path %q",
				relative,
			)
		}

		validated = append(validated, local)
	}

	for _, local := range validated {
		l.headers.invalidate(local)
		l.pages.invalidate(local)
		l.footers.invalidate(local)
	}

	return nil
}

// Stats returns a snapshot of the loader's cache statistics.
func (l *ContentLoader) Stats() LoaderStats {
	return LoaderStats{
		FileReads:     l.fileReads.Load(),
		CachedHeaders: l.headers.size(),
		CachedPages:   l.pages.size(),
		CachedFooters: l.footers.size(),
	}
}
