package gormimpl

import (
	"context"

	"github.com/jinzhu/gorm"
	"github.com/lyft/datacatalog/pkg/common"
	"github.com/lyft/datacatalog/pkg/repositories/errors"
	"github.com/lyft/datacatalog/pkg/repositories/models"
)

var entityToModel = map[common.Entity]interface{}{
	common.Artifact:  models.Artifact{},
	common.Dataset:   models.Dataset{},
	common.Partition: models.Partition{},
	common.Tag:       models.Tag{},
}

// Apply the list query on the source model. This method will apply the necessary joins, filters and
// pagination on the database for the given ListModelInputs.
func applyListModelsInput(ctx context.Context, tx *gorm.DB, sourceEntity common.Entity, in models.ListModelsInput) (*gorm.DB, error) {
	sourceModel, ok := entityToModel[sourceEntity]
	if !ok {
		return nil, errors.GetInvalidEntityError(sourceEntity.Name())
	}

	sourceTableName := tx.NewScope(sourceModel).TableName()
	for joiningEntity, joinCondition := range in.JoinEntityToConditionMap {
		joiningTableName := tx.NewScope(entityToModel[joiningEntity]).TableName()
		joinExpression, err := joinCondition.GetJoinOnDBQueryExpression(sourceTableName, joiningTableName)
		if err != nil {
			return nil, err
		}
		tx = tx.Joins(joinExpression)
	}

	for _, whereFilter := range in.Filters {
		filterEntity := whereFilter.GetDBEntity()
		entityTableName := tx.NewScope(entityToModel[filterEntity]).TableName()

		dbQueryExpr, err := whereFilter.GetDBQueryExpression(entityTableName)

		if err != nil {
			return nil, err
		}
		tx = tx.Where(dbQueryExpr.Query, dbQueryExpr.Args)
	}

	tx = tx.Limit(in.Limit)
	tx = tx.Offset(in.Offset)

	return tx, nil
}
