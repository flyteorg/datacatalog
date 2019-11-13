package transformers

import (
	"strconv"

	"github.com/lyft/datacatalog/pkg/common"
	"github.com/lyft/datacatalog/pkg/errors"
	"github.com/lyft/datacatalog/pkg/repositories/gormimpl"
	"github.com/lyft/datacatalog/pkg/repositories/models"
	datacatalog "github.com/lyft/datacatalog/protos/gen"
	"google.golang.org/grpc/codes"
)

func ApplyPagination(paginationOpts *datacatalog.PaginationOptions, input *models.ListModelsInput) error {
	var (
		offset    = common.DefaultOffset
		limit     = common.MaxLimit
		sortKey   = datacatalog.PaginationOptions_CREATION_TIME
		sortOrder = datacatalog.PaginationOptions_DESCENDING
	)

	if paginationOpts != nil {
		var err error
		offset, err = strconv.Atoi(paginationOpts.Token)
		if err != nil {
			return errors.NewDataCatalogErrorf(codes.InvalidArgument, "Invalid token %v", offset)
		}
		limit = int(paginationOpts.Limit)
		sortKey = paginationOpts.SortKey
		sortOrder = paginationOpts.SortOrder
	}

	input.Offset = offset
	input.Limit = limit
	input.SortParameter = gormimpl.NewGormSortParameter(sortKey, sortOrder)
	return nil
}
