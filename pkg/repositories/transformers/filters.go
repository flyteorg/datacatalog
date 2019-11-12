package transformers

import (
	"context"

	"github.com/lyft/datacatalog/pkg/common"

	"github.com/lyft/datacatalog/pkg/manager/impl/validators"
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
		modelPropertyFilters, err := ToModelValueFilter(ctx, filter)
		if err != nil {
			return models.ListModelsInput{}, err
		}

		modelFilters = append(modelFilters, modelPropertyFilters...)
		joiningEntity := modelPropertyFilters[0].GetDBEntity()
		if sourceEntity != joiningEntity {
			joinModelMap[joiningEntity] = gormimpl.NewGormJoinCondition(sourceEntity, joiningEntity)
		}
	}

	// Need to add limit/offset/Sort
	return models.ListModelsInput{
		Filters:                  modelFilters,
		JoinEntityToConditionMap: joinModelMap,
	}, nil
}

func ToModelValueFilter(ctx context.Context, singleFilter *datacatalog.SinglePropertyFilter) ([]models.ModelValueFilter, error) {
	modelValueFilters := make([]models.ModelValueFilter, 0, 1)

	switch propertyFilter := singleFilter.GetPropertyFilter().(type) {
	case *datacatalog.SinglePropertyFilter_PartitionFilter:
		partitionPropertyFilter := singleFilter.GetPartitionFilter()

		switch partitionProperty := partitionPropertyFilter.GetProperty().(type) {
		case *datacatalog.PartitionPropertyFilter_KeyVal:
			if err := validators.ValidateEmptyStringField(partitionProperty.KeyVal.Key, "PartitionKey"); err != nil {
				return nil, err
			}
			if err := validators.ValidateEmptyStringField(partitionProperty.KeyVal.Value, "PartitionValue"); err != nil {
				return nil, err
			}
			partitionKeyFilter := gormimpl.NewGormValueFilter(common.Partition, comparisonOperatorMap[singleFilter.Operator], partitionKeyFieldName, partitionProperty.KeyVal.Key)
			partitionValueFilter := gormimpl.NewGormValueFilter(common.Partition, comparisonOperatorMap[singleFilter.Operator], partitionValueFieldName, partitionProperty.KeyVal.Value)
			modelValueFilters = append(modelValueFilters, partitionKeyFilter, partitionValueFilter)
		}
	case *datacatalog.SinglePropertyFilter_TagFilter:
		switch tagProperty := propertyFilter.TagFilter.GetProperty().(type) {
		case *datacatalog.TagPropertyFilter_TagName:
			if err := validators.ValidateEmptyStringField(tagProperty.TagName, "TagName"); err != nil {
				return nil, err
			}
			tagNameFilter := gormimpl.NewGormValueFilter(common.Tag, comparisonOperatorMap[singleFilter.Operator], tagNameFieldName, tagProperty.TagName)
			modelValueFilters = append(modelValueFilters, tagNameFilter)
		}

	default:
		return nil, nil
	}
	return modelValueFilters, nil
}
