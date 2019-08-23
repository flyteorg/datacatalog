package gormimpl

import (
	"context"

	"github.com/jinzhu/gorm"
	"github.com/lyft/datacatalog/pkg/repositories/errors"
	"github.com/lyft/datacatalog/pkg/repositories/interfaces"
	"github.com/lyft/datacatalog/pkg/repositories/models"
	"github.com/lyft/flytestdlib/logger"

	datacatalog "github.com/lyft/flyteidl/gen/pb-go/flyteidl/datacatalog"
)

type dataSetRepo struct {
	db               *gorm.DB
	errorTransformer errors.ErrorTransformer

	// TODO: add metrics
}

func NewDatasetRepo(db *gorm.DB, errorTransformer errors.ErrorTransformer) interfaces.DatasetRepo {
	return &dataSetRepo{
		db:               db,
		errorTransformer: errorTransformer,
	}
}

// Create a Dataset model
func (h *dataSetRepo) Create(ctx context.Context, in models.Dataset) error {
	result := h.db.Create(&in)
	if result.Error != nil {
		return h.errorTransformer.ToDataCatalogError(result.Error)
	}
	return nil
}

// Get Dataset model
func (h *dataSetRepo) Get(ctx context.Context, in models.DatasetKey) (models.Dataset, error) {
	var ds models.Dataset
	result := h.db.Where(&models.Dataset{DatasetKey: in}).First(&ds)

	if result.Error != nil {
		logger.Debugf(ctx, "Unable to find Dataset: [%+v], err: %v", in, result.Error)
		return models.Dataset{}, h.errorTransformer.ToDataCatalogError(result.Error)
	}
	if result.RecordNotFound() {
		return models.Dataset{}, errors.GetMissingEntityError("Dataset", &datacatalog.DatasetID{
			Project: in.Project,
			Domain:  in.Domain,
			Name:    in.Name,
			Version: in.Version,
		})
	}

	return ds, nil
}
