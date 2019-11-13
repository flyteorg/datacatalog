package validators

import (
	"strconv"
	"strings"

	"fmt"

	"github.com/lyft/datacatalog/pkg/errors"
	datacatalog "github.com/lyft/datacatalog/protos/gen"
	"google.golang.org/grpc/codes"
)

const (
	defaultOffset = 0
	maxLimit      = 50
)

// The token is a string that should be opaque to the client
// It represents the offset as an integer encoded as a string,
// but in the future it can be a string that encodes anything
func ValidateToken(token string) (int, error) {
	if len(strings.Trim(token, " ")) == 0 {
		return defaultOffset, nil
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

// Validate the pagination options and return a new copy with missing values set to default
func ValidateAndGetPaginationOptions(options *datacatalog.PaginationOptions) (datacatalog.PaginationOptions, error) {
	if options == nil {
		options = &datacatalog.PaginationOptions{
			Limit: maxLimit,
			Token: "",
		}
	}
	offset, err := ValidateToken(options.Token)
	if err != nil {
		return datacatalog.PaginationOptions{}, err
	}
	options.Token = fmt.Sprintf("%d", offset)

	if options.Limit < 0 {
		return datacatalog.PaginationOptions{}, errors.NewDataCatalogErrorf(codes.InvalidArgument, "Invalid page limit %v", options.SortKey)
	} else if options.Limit > maxLimit {
		options.Limit = maxLimit
	}

	if options.SortKey != datacatalog.PaginationOptions_CREATION_TIME {
		return datacatalog.PaginationOptions{}, errors.NewDataCatalogErrorf(codes.InvalidArgument, "Invalid sort key %v", options.SortKey)
	}

	return *options, nil
}
