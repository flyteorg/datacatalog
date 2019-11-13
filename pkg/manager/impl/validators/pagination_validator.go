package validators

import (
	"strconv"

	"github.com/lyft/datacatalog/pkg/errors"
	datacatalog "github.com/lyft/datacatalog/protos/gen"
	"google.golang.org/grpc/codes"
)

// The token is a string that should be opaque to the client
// It represents the offset as an integer encoded as a string,
// but in the future it can be a string that encodes anything
func ValidateToken(token string) (int, error) {
	if token == "" {
		return 0, nil
	}
	offset, err := strconv.Atoi(token)
	if err != nil {
		return 0, err
	}
	if offset < 0 {
		return 0, errors.NewDataCatalogErrorf(codes.InvalidArgument, "Invalid token value: %s", token)
	}
	return offset, nil
}

func ValidatePaginationOptions(options *datacatalog.PaginationOptions) error {
	_, err := ValidateToken(options.Token)
	if err != nil {
		return err
	}

	if options.SortKey != datacatalog.PaginationOptions_CREATION_TIME {
		return errors.NewDataCatalogErrorf(codes.InvalidArgument, "Invalid sort key %v", options.SortKey)
	}

	return nil
}
