package services

import "sparrow-ssr/models"

type PropertySearcher interface {
	Search(
		siteDirectory string,
		filters PropertyFilters,
	) ([]models.Property, error)
}
