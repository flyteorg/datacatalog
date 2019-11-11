package gormimpl

import (
	"context"
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/lyft/datacatalog/pkg/common"
	"github.com/lyft/datacatalog/pkg/repositories/errors"
	"github.com/lyft/datacatalog/pkg/repositories/models"
)

// String formats for various GORM expression queries
const (
	equalQuery          = "%s.%s = ?"
	joinCondition       = "JOIN %s ON %s"
	joinEquals          = "%s.%s = %s.%s"
	joinAdditionalField = "AND %s"
)

// Map of the properties to join on for source table to joining table
type JoinOnMap map[string]string

// This provides the field names needed for joining a source Model to joining Model
var joinFieldNames = map[common.Entity]map[common.Entity]JoinOnMap{
	common.Artifact: {
		common.Partition: JoinOnMap{"artifact_id": "artifact_id"},
		common.Tag:       JoinOnMap{"artifact_id": "artifact_id"},
	},
}

func GetJoinOnFields(sourceEntity common.Entity, joiningEntity common.Entity) (JoinOnMap, error) {
	joiningEntityMap, ok := joinFieldNames[sourceEntity]
	if !ok {
		return nil, errors.GetInvalidEntityRelationshipError(sourceEntity.Name(), joiningEntity.Name())
	}

	fieldMap, ok := joiningEntityMap[joiningEntity]
	if !ok {
		return nil, errors.GetInvalidEntityRelationshipError(sourceEntity.Name(), joiningEntity.Name())
	}

	return fieldMap, nil
}

var entityToModel = map[common.Entity]interface{}{
	common.Artifact:  models.Artifact{},
	common.Dataset:   models.Dataset{},
	common.Partition: models.Partition{},
	common.Tag:       models.Tag{},
}

type gormValueFilterImpl struct {
	entity             common.Entity
	comparisonOperator common.ComparisonOperator
	field              string
	value              interface{}
}

func (g *gormValueFilterImpl) GetDBEntity() common.Entity {
	return g.entity
}

func (g *gormValueFilterImpl) GetDBQueryExpression(tableName string) (models.DBQueryExpr, error) {
	switch g.comparisonOperator {
	case common.Equal:
		return models.DBQueryExpr{
			Query: fmt.Sprintf(equalQuery, tableName, g.field),
			Args:  g.value,
		}, nil
	}
	return models.DBQueryExpr{}, common.GetUnsupportedFilterExpressionErr(g.comparisonOperator)
}

func NewGormValueFilter(entity common.Entity, comparisonOperator common.ComparisonOperator, field string, value interface{}) models.ModelValueFilter {
	return &gormValueFilterImpl{
		entity:             entity,
		comparisonOperator: comparisonOperator,
		field:              field,
		value:              value,
	}
}

// Contains the field details to construct GORM JOINs in the format:
// JOIN sourceTable ON sourceTable.sourceField = joiningTable.joiningField
type gormJoinConditionImpl struct {
	// The source entity type
	sourceEntity common.Entity
	// The joining entity type
	joiningEntity common.Entity
}

func (g *gormJoinConditionImpl) GetJoiningDBEntity() common.Entity {
	return g.joiningEntity
}

func (g *gormJoinConditionImpl) GetJoinOnDBQueryExpression(sourceTableName string, joiningTableName string) (string, error) {
	joinOnFieldMap, err := GetJoinOnFields(g.sourceEntity, g.joiningEntity)

	if err != nil {
		return "", err
	}

	joinFields := make([]string, 0, len(joinOnFieldMap))
	for sourceField, joiningField := range joinOnFieldMap {
		joinFieldCondition := fmt.Sprintf(joinEquals, sourceTableName, sourceField, joiningTableName, joiningField)
		if len(joinFields) > 1 {
			// append "AND" for joins on more than one column
			joinFieldCondition = fmt.Sprintf(joinAdditionalField, joinFieldCondition)
		}
		joinFields = append(joinFields, joinFieldCondition)
	}

	return fmt.Sprintf(joinCondition, joiningTableName, strings.Join(joinFields, " ")), nil
}

func NewGormJoinCondition(sourceEntity common.Entity, joiningEntity common.Entity) models.ModelJoinCondition {
	return &gormJoinConditionImpl{
		joiningEntity: joiningEntity,
		sourceEntity:  sourceEntity,
	}
}

func applyListModelsInput(ctx context.Context, db *gorm.DB, sourceEntity common.Entity, in models.ListModelsInput) (*gorm.DB, error) {
	sourceModel, ok := entityToModel[sourceEntity]

	if !ok {
		return nil, errors.GetInvalidEntityError(sourceEntity.Name()) // TODO return err
	}
	sourceTableName := db.NewScope(sourceModel).TableName()
	for joiningEntity, joinCondition := range in.JoinEntityToConditionMap {
		joiningTableName := db.NewScope(entityToModel[joiningEntity]).TableName()
		joinExpression, err := joinCondition.GetJoinOnDBQueryExpression(sourceTableName, joiningTableName)
		if err != nil {
			return nil, err
		}
		db = db.Joins(joinExpression)
	}

	for _, whereFilter := range in.Filters {
		filterEntity := whereFilter.GetDBEntity()
		entityTableName := db.NewScope(entityToModel[filterEntity]).TableName()

		dbQueryExpr, err := whereFilter.GetDBQueryExpression(entityTableName)

		if err != nil {
			return nil, err
		}
		db = db.Where(dbQueryExpr.Query, dbQueryExpr.Args)
	}

	return db, nil
}
