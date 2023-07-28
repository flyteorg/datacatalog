package errors

import (
	"errors"
	"reflect"

	"github.com/flyteorg/flytestdlib/logger"

	catalogErrors "github.com/flyteorg/datacatalog/pkg/errors"
	"google.golang.org/grpc/codes"
	"gorm.io/gorm"
)

type genericErrorTransformer struct {
}

func (p *genericErrorTransformer) fromGormError(err error) error {
	switch err.Error() {
	case gorm.ErrRecordNotFound.Error():
		return catalogErrors.NewDataCatalogErrorf(codes.NotFound, "entry not found")
	default:
		logger.InfofNoCtx("Generic error detected. Error type: [%v]", reflect.TypeOf(err))
		return catalogErrors.NewDataCatalogErrorf(codes.Internal, unexpectedType, err)
	}
}

func (p *genericErrorTransformer) ToDataCatalogError(err error) error {
	if unwrappedErr := errors.Unwrap(err); unwrappedErr != nil {
		err = unwrappedErr
	}

	return p.fromGormError(err)
}

func NewGenericErrorTransformer() ErrorTransformer {
	return &genericErrorTransformer{}
}
