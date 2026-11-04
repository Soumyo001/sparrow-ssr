package services

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"sparrow-ssr/models"
)

// ContentLoader reads JSON content from local storage.
type ContentLoader struct {
	root string
}

func NewContentLoader(
	contentRoot string,
) (*ContentLoader, error) {
	root, err := filepath.Abs(contentRoot)
	if err != nil {
		return nil, fmt.Errorf(
			"resolve content directory: %w", err,
		)
	}

	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf(
			"access content directory: %w", err,
		)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf(
			"content root is not a directory: %s",
			root,
		)
	}

	return &ContentLoader{root: root}, nil
}

// Load reads header, page and footer concurrently.
// Each goroutine writes to a separate result field.
func (l *ContentLoader) Load(
	location models.PageLocation,
) (models.PageContent, error) {
	var content models.PageContent
	var wg sync.WaitGroup

	var headerErr error
	var pageErr error
	var footerErr error

	wg.Add(3)

	go func() {
		defer wg.Done()

		headerErr = readJSON(
			l.root,
			location.HeaderJSON,
			&content.Header,
		)
	}()

	go func() {
		defer wg.Done()

		pageErr = readJSON(
			l.root,
			location.PageJSON,
			&content.Page,
		)
	}()

	go func() {
		defer wg.Done()

		footerErr = readJSON(
			l.root,
			location.FooterJSON,
			&content.Footer,
		)
	}()

	// all three reads must finish before using the results.
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
