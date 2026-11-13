package services

import (
	"fmt"
	"net/url"
	"strings"
)

// RegisterPreviewAlias registers an explicitly configured preview
// hostname for one existing publisher.
//
// Call this only during startup, before the HTTP server begins
// handling requests. The catalog remains read-only afterward.
//
// Existing publisher page locations are reused unchanged, so
// their canonical domain, storage paths and site key are preserved.
func (c *Catalog) RegisterPreviewAlias(
	sourceHost, previewHost string,
) error {
	if c == nil {
		return fmt.Errorf("catalog is nil")
	}

	source, err := previewHostname(sourceHost)
	if err != nil {
		return fmt.Errorf("invalid preview source host: %w", err)
	}

	preview, err := previewHostname(previewHost)
	if err != nil {
		return fmt.Errorf("invalid preview host: %w", err)
	}

	if source == preview {
		return fmt.Errorf(
			"preview host must differ from publisher host",
		)
	}

	paths := make([]string, 0)

	// Check every existing route before mutating the catalog.
	for key := range c.routes {
		if key.host == preview {
			return fmt.Errorf(
				"preview host %q is already registered",
				preview,
			)
		}

		if key.host == source {
			paths = append(paths, key.path)
		}
	}

	if len(paths) == 0 {
		return fmt.Errorf(
			"publisher host %q has no registered routes",
			source,
		)
	}

	for _, pagePath := range paths {
		originalKey := routeKey{
			host: source,
			path: pagePath,
		}

		previewKey := routeKey{
			host: preview,
			path: pagePath,
		}

		// Do not alter PageLocation.Domain.
		// The original published canonical identity remains intact.
		c.routes[previewKey] = c.routes[originalKey]
	}

	return nil
}

// Accept a hostname only, not a URL, port or URL path.
func previewHostname(raw string) (string, error) {
	host := strings.ToLower(strings.TrimSpace(raw))

	parsed, err := url.Parse("https://" + host)
	if err != nil ||
		host == "" ||
		!strings.Contains(host, ".") ||
		parsed.Host != host ||
		parsed.Hostname() != host ||
		parsed.Port() != "" ||
		parsed.User != nil ||
		parsed.Path != "" ||
		parsed.RawQuery != "" ||
		parsed.Fragment != "" {
		return "", fmt.Errorf(
			"expected a DNS hostname without scheme, port or path",
		)
	}

	return host, nil
}
