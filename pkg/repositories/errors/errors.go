// Generic errors used in the repos layer
package errors

import (
	"github.com/golang/protobuf/proto"
	"github.com/lyft/datacatalog/pkg/errors"
	"google.golang.org/grpc/codes"
)

const (
	notFound      = "missing entity of type %s with identifier %v"
	invalidJoin   = "cannot relate entity %s with entity %s"
	invalidEntity = "no such entity %s"
)

func GetMissingEntityError(entityType string, identifier proto.Message) error {
	return errors.NewDataCatalogErrorf(codes.NotFound, notFound, entityType, identifier)
}

func GetInvalidEntityRelationshipError(entityType string, otherEntityType string) error {
	return errors.NewDataCatalogErrorf(codes.InvalidArgument, invalidJoin, entityType, otherEntityType)
}

func GetInvalidEntityError(entityType string) error {
	return errors.NewDataCatalogErrorf(codes.InvalidArgument, invalidEntity, entityType)
}
