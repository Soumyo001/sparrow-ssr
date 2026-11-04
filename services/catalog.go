package services

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"sparrow-ssr/models"
)

type routeKey struct {
	host string
	path string
}

// Catalog is read-only after initialization.
type Catalog struct {
	routes map[routeKey]models.PageLocation
}

func NewCatalog(
	contentRoot, deploymentFile string,
) (*Catalog, error) {
	metadataBytes, err := os.ReadFile(deploymentFile)
	if err != nil {
		return nil, fmt.Errorf(
			"read deployment metadata: %w", err,
		)
	}

	var deployment models.DeploymentResponse

	if err := json.Unmarshal(metadataBytes, &deployment); err != nil {
		return nil, fmt.Errorf(
			"decode deployment metadata: %w", err,
		)
	}

	published, err := url.Parse(
		deployment.Data.PublishedDomain,
	)

	if err != nil || published.Scheme != "https" || published.Hostname() == "" {
		return nil, fmt.Errorf(
			"invalid published_domain: %q", deployment.Data.PublishedDomain,
		)
	}

	publishedHost := strings.ToLower(
		published.Hostname(),
	)

	root, err := filepath.Abs(contentRoot)
	if err != nil {
		return nil, err
	}

	fileNames := make(map[string]string)
	var directory string

	// Resolve deployment CDN paths into local file paths.
	for _, remotePath := range deployment.Data.UploadedFilePaths {
		cdnURL, err := url.Parse(remotePath)

		if err != nil ||
			cdnURL.Scheme != "https" ||
			!strings.EqualFold(
				cdnURL.Hostname(), "cdn.123presto.com",
			) ||
			cdnURL.User != nil ||
			cdnURL.RawQuery != "" ||
			cdnURL.Fragment != "" {
			return nil, fmt.Errorf(
				"invalid CDN file URL: %q", remotePath,
			)
		}

		// Only normalized JSON object paths are accepted.
		if !strings.HasPrefix(cdnURL.Path, "/") ||
			path.Clean(cdnURL.Path) != cdnURL.Path ||
			path.Ext(cdnURL.Path) != ".json" {
			return nil, fmt.Errorf(
				"invalid JSON storage path: %q",
				cdnURL.Path,
			)
		}

		relative := strings.TrimPrefix(
			cdnURL.Path, "/",
		)

		folder := path.Dir(relative) // beta/preview/test-travel-templates

		if directory == "" {
			directory = folder
		} else if folder != directory {
			return nil, fmt.Errorf(
				"deployment references multiple site directories",
			)
		}

		filename := filepath.Join(
			root, filepath.FromSlash(relative),
		)

		info, err := os.Stat(filename)
		if err != nil || info.IsDir() {
			return nil, fmt.Errorf(
				"local content file unavailable: %s",
				filename,
			)
		}

		base := path.Base(relative)

		if _, exists := fileNames[base]; exists {
			return nil, fmt.Errorf(
				"duplicate uploaded filename %q", base,
			)
		}

		fileNames[base] = relative
	}

	// All pages in this supplied site use these two files.
	for _, shared := range []string{
		"header.json",
		"footer.json",
	} {
		if _, found := fileNames[shared]; !found {
			return nil, fmt.Errorf(
				"missing %s in deployment metadata",
				shared,
			)
		}
	}

	var header models.SiteHeader
	if err := readJSON(
		root, fileNames["header.json"], &header,
	); err != nil {
		return nil, err
	}
	if header.HeaderHTML == "" {
		return nil, fmt.Errorf(
			"header.json has an empty header",
		)
	}

	var footer models.SiteFooter
	if err := readJSON(
		root, fileNames["footer.json"], &footer,
	); err != nil {
		return nil, err
	}
	if footer.FooterHTML == "" {
		return nil, fmt.Errorf(
			"footer.json has an empty footer",
		)
	}

	catalog := &Catalog{
		routes: make(map[routeKey]models.PageLocation),
	}

	// Index actual page JSON files.
	for name, relative := range fileNames {
		if name == "header.json" ||
			name == "footer.json" ||
			name == "links.json" {
			continue
		}

		var page models.PageDocument

		if err := readJSON(
			root, relative, &page,
		); err != nil {
			return nil, err
		}

		canonical, err := url.Parse(
			page.Head.CanonicalURL,
		)

		if err != nil ||
			canonical.Scheme != "https" ||
			!strings.EqualFold(
				canonical.Hostname(), publishedHost,
			) ||
			canonical.User != nil ||
			canonical.RawQuery != "" ||
			canonical.Fragment != "" ||
			page.Body.Content == "" {
			return nil, fmt.Errorf(
				"invalid canonical page: %s",
				relative,
			)
		}

		publicPath := normalizePath(canonical.Path)

		key := routeKey{
			host: publishedHost,
			path: publicPath,
		}

		if _, exists := catalog.routes[key]; exists {
			return nil, fmt.Errorf(
				"duplicate canonical URL %s",
				page.Head.CanonicalURL,
			)
		}

		catalog.routes[key] = models.PageLocation{
			Domain:           publishedHost,
			URLPath:          publicPath,
			ContentDirectory: directory,
			PageJSON:         relative,
			HeaderJSON:       fileNames["header.json"],
			FooterJSON:       fileNames["footer.json"],
		}
	}

	if len(catalog.routes) == 0 {
		return nil, fmt.Errorf(
			"deployment contains no routable pages",
		)
	}

	return catalog, nil
}

func readJSON(
	root, relative string,
	destination any,
) error {
	filename := filepath.Join(
		root, filepath.FromSlash(relative),
	)

	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf(
			"read %s: %w", relative, err,
		)
	}

	if err := json.Unmarshal(data, destination); err != nil {
		return fmt.Errorf(
			"decode %s: %w", relative, err,
		)
	}

	return nil
}

// Resolve is safe for concurrent reads after initialization.
func (c *Catalog) Resolve(host, requestPath string) (models.PageLocation, bool) {
	parsed := url.URL{Host: host}
	hostname := parsed.Hostname()

	location, found := c.routes[routeKey{
		host: strings.ToLower(hostname),
		path: normalizePath(requestPath),
	}]

	return location, found
}

func (c *Catalog) Count() int {
	return len(c.routes)
}

func normalizePath(p string) string {
	if p == "" {
		return "/"
	}

	return path.Clean(
		"/" + strings.TrimPrefix(p, "/"),
	)
}
