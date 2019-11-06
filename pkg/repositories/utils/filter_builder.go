package utils

import (
	"context"

	"github.com/lyft/datacatalog/pkg/common"
	datacatalog "github.com/lyft/datacatalog/protos/gen"
)

// Map of the properties to join on for source table to joining table
type JoinOnMap map[string]string

// This provides the field names needed for joining a source Model to Models others
var joinFieldNames = map[common.Entity]map[common.Entity]JoinOnMap{
	common.Artifact: {
		common.Partition: JoinOnMap{"artifact_id": "artifact_id"},
		common.Tag:       JoinOnMap{"artifact_id": "artifact_id"},
	},
}

var comparisonOperatorMap = map[datacatalog.SinglePropertyFilter_ComparisonOperator]common.ComparisonOperator{
	datacatalog.SinglePropertyFilter_EQUALS: common.Equal,
}

func ValidateListInput(ctx context.Context, filterExpression *datacatalog.FilterExpression) error {

	// Validate that the join entities provided are supported queries for the entity at hand
	// Validate for filters like Labels, they are not filtering multiple times over the same key
	return nil
}

func ConstructListInput(ctx context.Context, filterExpression *datacatalog.FilterExpression) (common.ListModelsInput, error) {
	joinModelMap := make(map[common.Entity]common.ModelJoinCondition, 0)
	sourceEntity := common.Artifact
	modelFilters := make([]common.ModelValueFilter, 0, len(filterExpression.GetFilters()))
	for _, filter := range filterExpression.GetFilters() {
		modelPropertyFilters, err := GenerateModelValueFilter(filter)

		if err != nil {
			modelFilters = append(modelFilters, modelPropertyFilters...)
			joiningEntity := modelPropertyFilters[0].GetDBEntity()
			if sourceEntity != joiningEntity {
				joinOnFields, ok := joinFieldNames[sourceEntity][joiningEntity]

				if ok {
					joinModelMap[joiningEntity] = common.NewGormJoinCondition(joiningEntity, "artifacts", "partitions", joinOnFields)
				} else {
					// err here because no joins are available
				}
			}
		}
	}

	// Need to add limit/offset/Sort
	return common.ListModelsInput{
		Filters: modelFilters,
		Joins:   joinModelMap,
	}, nil
}

func GenerateModelValueFilter(singleFilter *datacatalog.SinglePropertyFilter) ([]common.ModelValueFilter, error) {
	modelValueFilter := make([]common.ModelValueFilter, 0, 1)

	switch propertyFilter := singleFilter.GetPropertyFilter().(type) {
	case *datacatalog.SinglePropertyFilter_PartitionFilter:
		partitionPropertyFilter := singleFilter.GetPartitionFilter()
		switch partitionProperty := partitionPropertyFilter.GetProperty().(type) {
		case *datacatalog.PartitionPropertyFilter_KeyVal:
			partitionKey := partitionProperty.KeyVal.Key
			partitionValue := partitionProperty.KeyVal.Value
			modelValueFilter = append(modelValueFilter, common.NewGormValueFilter(common.Partition, comparisonOperatorMap[singleFilter.Operator], "key", partitionKey))
			modelValueFilter = append(modelValueFilter, common.NewGormValueFilter(common.Partition, comparisonOperatorMap[singleFilter.Operator], "value", partitionValue))
		}
	case *datacatalog.SinglePropertyFilter_TagFilter:
		switch tagProperty := propertyFilter.TagFilter.GetProperty().(type) {
		case *datacatalog.TagPropertyFilter_TagName:
			tagName := tagProperty.TagName
			modelValueFilter = append(modelValueFilter, common.NewGormValueFilter(common.Artifact, comparisonOperatorMap[singleFilter.Operator], "tag_name", tagName))
		}

	default:
		return nil, nil
	}

	return modelValueFilter, nil
}
