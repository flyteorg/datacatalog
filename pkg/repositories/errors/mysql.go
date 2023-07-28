package errors

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/flyteorg/flytestdlib/logger"
	"github.com/go-sql-driver/mysql"

	catalogErrors "github.com/flyteorg/datacatalog/pkg/errors"
	"google.golang.org/grpc/codes"
	"gorm.io/gorm"
)

// MySql error codes
const (
	duplicateKeyError             = 1062
	// uniqueConstraintViolationCode = "23505"
	// undefinedTable                = "42P01"
)

type mysqlErrorTransformer struct {
}

const (
	duplicateKeyErrorFormat = "duplicate key value violates unique constraint %s"
	defaultMysqlError            = "failed database operation with code [%s] and msg [%s]"
	// unexpectedType            = "unexpected error type for: %v"
	// uniqueConstraintViolation = "value with matching already exists (%s)"
	// defaultPgError            = "failed database operation with code [%s] and msg [%s]"
	// unsupportedTableOperation = "cannot query with specified table attributes: %s"
)

func (p *mysqlErrorTransformer) fromGormError(err error) error {
	switch err.Error() {
	case gorm.ErrRecordNotFound.Error():
		return catalogErrors.NewDataCatalogErrorf(codes.NotFound, "entry not found")
	default:
		return catalogErrors.NewDataCatalogErrorf(codes.Internal, unexpectedType, err)
	}
}

func (p *mysqlErrorTransformer) ToDataCatalogError(err error) error {
	if unwrappedErr := errors.Unwrap(err); unwrappedErr != nil {
		err = unwrappedErr
	}

	mysqlError, ok := err.(*mysql.MySQLError)
	if !ok {
		logger.InfofNoCtx("Unable to cast to mysql.MySQLError. Error type: [%v]",
			reflect.TypeOf(err))
		return p.fromGormError(err)
	}

	switch mysqlError.Number {
	case 1062:
		return catalogErrors.NewDataCatalogErrorf(codes.AlreadyExists, duplicateKeyErrorFormat, mysqlError.Message)
	// case uniqueConstraintViolationCode:
	// 	return catalogErrors.NewDataCatalogErrorf(codes.AlreadyExists, uniqueConstraintViolation, pqError.Message)
	// case undefinedTable:
	// 	return catalogErrors.NewDataCatalogErrorf(codes.InvalidArgument, unsupportedTableOperation, pqError.Message)
	default:
		return catalogErrors.NewDataCatalogErrorf(codes.Unknown, fmt.Sprintf(defaultMysqlError, mysqlError.Number, mysqlError.Message))
	}
}

func NewMySqlErrorTransformer() ErrorTransformer {
	return &mysqlErrorTransformer{}
}
