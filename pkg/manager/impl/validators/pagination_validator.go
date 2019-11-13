package validators

import (
	"strconv"
	"strings"

	"fmt"

	"github.com/lyft/datacatalog/pkg/common"
	"github.com/lyft/datacatalog/pkg/errors"
	datacatalog "github.com/lyft/datacatalog/protos/gen"
	"google.golang.org/grpc/codes"
)

// The token is a string that should be opaque to the client
// It represents the offset as an integer encoded as a string,
// but in the future it can be a string that encodes anything
func ValidateToken(token string) (int, error) {
	if len(strings.Trim(token, " ")) == 0 {
		return common.DefaultOffset, nil
	}
	offset, err := strconv.Atoi(token)
	if err != nil {
		return 0, errors.NewDataCatalogErrorf(codes.InvalidArgument, "Invalid token value: %s", token)
	}
	if offset < 0 {
		return 0, errors.NewDataCatalogErrorf(codes.InvalidArgument, "Token needs to be a positive value: %s", token)
	}
	return offset, nil
}

// Validate the pagination options and set default limits
func ValidatePagination(options *datacatalog.PaginationOptions) error {
	offset, err := ValidateToken(options.Token)
	if err != nil {
		return err
	}
	options.Token = fmt.Sprintf("%d", offset)

	if options.Limit <= 0 {
		return errors.NewDataCatalogErrorf(codes.InvalidArgument, "Invalid page limit %v", options.Limit)
	} else if options.Limit > common.MaxLimit {
		options.Limit = common.MaxLimit
	}

	if options.SortKey != datacatalog.PaginationOptions_CREATION_TIME {
		return errors.NewDataCatalogErrorf(codes.InvalidArgument, "Invalid sort key %v", options.SortKey)
	}

	if options.SortOrder != datacatalog.PaginationOptions_ASCENDING &&
		options.SortOrder != datacatalog.PaginationOptions_DESCENDING {
		return errors.NewDataCatalogErrorf(codes.InvalidArgument, "Invalid sort order %v", options.SortOrder)
	}

	return nil
}
