package gormimpl

import (
	"fmt"

	"github.com/lyft/datacatalog/pkg/repositories/models"
	datacatalog "github.com/lyft/datacatalog/protos/gen"
)

// Container for the sort details
type sortParameter struct {
	sortKey   datacatalog.PaginationOptions_SortKey
	sortOrder datacatalog.PaginationOptions_SortOrder
}

// Generate the DBOrderExpression that GORM needs to order models
func (s *sortParameter) GetDBOrderExpression() string {
	var sortOrderString string
	switch s.sortOrder {
	case datacatalog.PaginationOptions_ASCENDING:
		sortOrderString = "asc"
	case datacatalog.PaginationOptions_DESCENDING:
		sortOrderString = "desc"
	default:
		sortOrderString = "desc"
	}

	var sortKeyString string
	switch s.sortKey {
	case datacatalog.PaginationOptions_CREATION_TIME:
		sortKeyString = "created_at"
	default:
		sortKeyString = "created_at"
	}
	return fmt.Sprintf(sortQuery, sortKeyString, sortOrderString)
}

// Create SortParameter for GORM
func NewGormSortParameter(sortKey datacatalog.PaginationOptions_SortKey, sortOrder datacatalog.PaginationOptions_SortOrder) models.SortParameter {
	return &sortParameter{sortKey: sortKey, sortOrder: sortOrder}
}
