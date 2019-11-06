package common

import (
	"fmt"
	"strings"

	"github.com/lyft/datacatalog/pkg/errors"
	"github.com/lyft/datacatalog/pkg/repositories/utils"
	"google.golang.org/grpc/codes"
)

// String formats for various GORM expression queries
const (
	joinFormat          = "%s.%s"
	equalQuery          = "%s = ?"
	joinCondition       = "JOIN %s ON %s"
	joinEquals          = "%s.%s = %s.%s"
	joinAdditionalField = "AND %s"
)

type Entity int

const (
	Artifact Entity = iota
	Dataset
	Partition
	Tag
)

type ComparisonOperator int

const (
	Equal ComparisonOperator = iota
	// TODO: Add more operators as needed
)

func GetUnsupportedFilterExpressionErr(operator ComparisonOperator) error {
	return errors.NewDataCatalogErrorf(codes.InvalidArgument, "unsupported filter expression operator: %s",
		operator)
}

type ListModelsInput struct {
	// Joins for the list query. It is represented as a 1:1 mapping between a joiningEntity and a joinCondition
	Joins map[Entity]ModelJoinCondition
	// Value filters for the list query
	Filters []ModelValueFilter
	// The number of models to list
	Limit int
	// The token to offset results by
	Offset int
	// Parameter to sort by
	SortParameter
}

type SortParameter interface {
	GetDBOrderExpression() string
}

// Generates db filter expressions for model values
type ModelValueFilter interface {
	GetDBEntity() Entity
	GetDBQueryExpression() (DBQueryExpr, error)
	GetJoinedDBQueryExpression(tableName string) (DBQueryExpr, error)
}

// Encapsulates the query and necessary arguments to issue a DB query.
type DBQueryExpr struct {
	Query string
	Args  interface{}
}

type gormValueFilterImpl struct {
	entity             Entity
	comparisonOperator ComparisonOperator
	field              string
	value              interface{}
}

func (g *gormValueFilterImpl) GetDBEntity() Entity {
	return g.entity
}

func (g *gormValueFilterImpl) GetJoinedDBQueryExpression(tableName string) (DBQueryExpr, error) {
	g.field = fmt.Sprintf(joinFormat, tableName, g.field)
	return g.GetDBQueryExpression()
}

func (g *gormValueFilterImpl) GetDBQueryExpression() (DBQueryExpr, error) {
	switch g.comparisonOperator {
	case Equal:
		return DBQueryExpr{
			Query: fmt.Sprintf(equalQuery, g.field),
			Args:  g.value,
		}, nil
	}
	return DBQueryExpr{}, GetUnsupportedFilterExpressionErr(g.comparisonOperator)
}

func NewGormValueFilter(entity Entity, comparisonOperator ComparisonOperator, field string, value interface{}) ModelValueFilter {
	return &gormValueFilterImpl{
		entity:             entity,
		comparisonOperator: comparisonOperator,
		field:              field,
		value:              value,
	}
}

type ModelJoinCondition interface {
	GetJoiningDBEntity() Entity
	GetJoinOnDBQueryExpression() (string, error)
}

// Contains the field details to construct GORM JOINs in the format:
// JOIN sourceTable ON sourceTable.sourceField = joiningTable.joiningField
type gormJoinConditionImpl struct {
	// The joining entity type
	joiningEntity Entity
	// A map of the originating tables field names to the joining tables field names
	joinOnFieldMap utils.JoinOnMap
	// The source table's name
	sourceTableName string
	// The joining table's name
	joiningTableName string
}

func (g *gormJoinConditionImpl) GetJoiningDBEntity() Entity {
	return g.joiningEntity
}

func (g *gormJoinConditionImpl) GetJoinOnDBQueryExpression() (string, error) {
	joinFields := make([]string, 0, len(g.joinOnFieldMap))
	for sourceField, joiningField := range g.joinOnFieldMap {
		joinFieldCondition := fmt.Sprintf(joinEquals, g.sourceTableName, sourceField, g.joiningTableName, joiningField)
		if len(joinFields) > 1 {
			// append "AND" for joins on more than one column
			joinFieldCondition = fmt.Sprintf(joinAdditionalField, joinFieldCondition)
		}
		joinFields = append(joinFields)
	}

	return fmt.Sprintf(joinCondition, strings.Join(joinFields, " ")), nil
}

func NewGormJoinCondition(joiningEntity Entity, sourceTableName string, joiningTableName string, joinOnFieldMap utils.JoinOnMap) ModelJoinCondition {
	return &gormJoinConditionImpl{
		joiningEntity:    joiningEntity,
		sourceTableName:  sourceTableName,
		joiningTableName: joiningTableName,
		joinOnFieldMap:   joinOnFieldMap,
	}
}
