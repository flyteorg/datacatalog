package transformers

import (
	"strconv"
	"strings"

	"github.com/flyteorg/datacatalog/pkg/common"
	"github.com/flyteorg/datacatalog/pkg/errors"
	"github.com/flyteorg/datacatalog/pkg/repositories/gormimpl"
	"github.com/flyteorg/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"google.golang.org/grpc/codes"
)

func ApplyPagination(paginationOpts *datacatalog.PaginationOptions, input *models.ListModelsInput) error {
	var (
		offset    = common.DefaultPageOffset
		limit     = common.MaxPageLimit
		sortKey   = datacatalog.PaginationOptions_CREATION_TIME
		sortOrder = datacatalog.PaginationOptions_DESCENDING
	)

	if paginationOpts != nil {
		// if the token is empty, that is still valid input since it is optional
		if len(strings.Trim(paginationOpts.Token, " ")) == 0 {
			offset = common.DefaultPageOffset
		} else {
			parsedOffset, err := strconv.ParseUint(paginationOpts.Token, 10, 32)
			if err != nil {
				return errors.NewDataCatalogErrorf(codes.InvalidArgument, "Invalid token %v", offset)
			}
			offset = uint32(parsedOffset)
		}
		limit = paginationOpts.Limit
		sortKey = paginationOpts.SortKey
		sortOrder = paginationOpts.SortOrder
	}

	input.Offset = offset
	input.Limit = limit
	input.SortParameter = gormimpl.NewGormSortParameter(sortKey, sortOrder)
	return nil
}
