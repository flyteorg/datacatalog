package gormimpl

import (
	"testing"

	datacatalog "github.com/lyft/datacatalog/protos/gen"
	"github.com/stretchr/testify/assert"
)

func TestSortAsc(t *testing.T) {
	dbSortExpression := NewGormSortParameter(
		datacatalog.PaginationOptions_CREATION_TIME,
		datacatalog.PaginationOptions_ASCENDING).GetDBOrderExpression()

	assert.Equal(t, dbSortExpression, "created_at asc")
}

func TestSortDesc(t *testing.T) {
	dbSortExpression := NewGormSortParameter(
		datacatalog.PaginationOptions_CREATION_TIME,
		datacatalog.PaginationOptions_DESCENDING).GetDBOrderExpression()

	assert.Equal(t, dbSortExpression, "created_at desc")
}
