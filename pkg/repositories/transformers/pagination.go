package transformers

import (
	"strconv"

	"github.com/lyft/datacatalog/pkg/errors"
	"github.com/lyft/datacatalog/pkg/repositories/gormimpl"
	"github.com/lyft/datacatalog/pkg/repositories/models"
	datacatalog "github.com/lyft/datacatalog/protos/gen"
	"google.golang.org/grpc/codes"
)

func ApplyPagination(paginationOpts *datacatalog.PaginationOptions, input *models.ListModelsInput) error {
	offset, err := strconv.Atoi(paginationOpts.Token)
	if err != nil {
		return errors.NewDataCatalogErrorf(codes.InvalidArgument, "Invalid token %v", offset)
	}
	input.Offset = offset
	input.Limit = int(paginationOpts.Limit)
	input.SortParameter = gormimpl.NewGormSortParameter(paginationOpts.SortKey, paginationOpts.Order)
	return nil
}
