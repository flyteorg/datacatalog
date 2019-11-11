package transformers

import (
	"context"

	"github.com/lyft/datacatalog/pkg/common"

	"github.com/lyft/datacatalog/pkg/repositories/gormimpl"
	"github.com/lyft/datacatalog/pkg/repositories/models"
	datacatalog "github.com/lyft/datacatalog/protos/gen"
)

const (
	partitionKeyFieldName   = "key"
	partitionValueFieldName = "value"
	tagNameFieldName        = "tag_name"
)

var comparisonOperatorMap = map[datacatalog.SinglePropertyFilter_ComparisonOperator]common.ComparisonOperator{
	datacatalog.SinglePropertyFilter_EQUALS: common.Equal,
}

func ValidateListInput(ctx context.Context, filterExpression *datacatalog.FilterExpression) error {

	// Validate that the join entities provided are supported queries for the entity at hand
	// Validate for filters like Labels, they are not filtering multiple times over the same key
	return nil
}

func ToListInput(ctx context.Context, sourceEntity common.Entity, filterExpression *datacatalog.FilterExpression) (models.ListModelsInput, error) {
	// ListInput is composed of ModelFilters and ModelJoins. Let's construct those filters and joins.
	modelFilters := make([]models.ModelValueFilter, 0, len(filterExpression.GetFilters()))
	joinModelMap := make(map[common.Entity]models.ModelJoinCondition, 0)
	for _, filter := range filterExpression.GetFilters() {
		modelPropertyFilters, err := ToModelValueFilter(filter)

		if err != nil {
			modelFilters = append(modelFilters, modelPropertyFilters...)
			joiningEntity := modelPropertyFilters[0].GetDBEntity()
			if sourceEntity != joiningEntity {
				joinModelMap[joiningEntity] = gormimpl.NewGormJoinCondition(sourceEntity, joiningEntity)
			}
		}
	}

	// Need to add limit/offset/Sort
	return models.ListModelsInput{
		Filters:                  modelFilters,
		JoinEntityToConditionMap: joinModelMap,
	}, nil
}

func ToModelValueFilter(singleFilter *datacatalog.SinglePropertyFilter) ([]models.ModelValueFilter, error) {
	modelValueFilters := make([]models.ModelValueFilter, 0, 1)

	switch propertyFilter := singleFilter.GetPropertyFilter().(type) {
	case *datacatalog.SinglePropertyFilter_PartitionFilter:
		partitionPropertyFilter := singleFilter.GetPartitionFilter()
		switch partitionProperty := partitionPropertyFilter.GetProperty().(type) {
		case *datacatalog.PartitionPropertyFilter_KeyVal:
			partitionKey := partitionProperty.KeyVal.Key
			partitionValue := partitionProperty.KeyVal.Value
			modelValueFilters = append(modelValueFilters, gormimpl.NewGormValueFilter(common.Partition, comparisonOperatorMap[singleFilter.Operator], partitionKeyFieldName, partitionKey))
			modelValueFilters = append(modelValueFilters, gormimpl.NewGormValueFilter(common.Partition, comparisonOperatorMap[singleFilter.Operator], partitionValueFieldName, partitionValue))
		}
	case *datacatalog.SinglePropertyFilter_TagFilter:
		switch tagProperty := propertyFilter.TagFilter.GetProperty().(type) {
		case *datacatalog.TagPropertyFilter_TagName:
			tagName := tagProperty.TagName
			modelValueFilters = append(modelValueFilters, gormimpl.NewGormValueFilter(common.Artifact, comparisonOperatorMap[singleFilter.Operator], tagNameFieldName, tagName))
		}

	default:
		return nil, nil
	}

	return modelValueFilters, nil
}
