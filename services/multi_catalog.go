package services

import (
	"fmt"
	"strings"

	"sparrow-ssr/models"
)

// NewMultiCatalog combines deployment catalogs into one read-only route index.
// Every route is identified by both its public hostname and URL path.
func NewMultiCatalog(contentRoot string, deploymentFiles []string) (*Catalog, error) {
	if len(deploymentFiles) == 0 {
		return nil, fmt.Errorf("no deployment metadata files provided")
	}

	combined := &Catalog{
		routes: make(map[routeKey]models.PageLocation),
	}

	for _, file := range deploymentFiles {
		if strings.TrimSpace(file) == "" {
			return nil, fmt.Errorf("deployment metadata path cannot be empty")
		}

		siteCatalog, err := NewCatalog(contentRoot, file)
		if err != nil {
			return nil, fmt.Errorf("load deployment %q: %w", file, err)
		}

		for key, location := range siteCatalog.routes {
			if _, exists := combined.routes[key]; exists {
				return nil, fmt.Errorf(
					"duplicate route across deployments: host %q, path %q",
					key.host, key.path,
				)
			}
			combined.routes[key] = location
		}
	}

	return combined, nil
}
