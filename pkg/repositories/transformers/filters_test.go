package transformers

import (
	"context"
	"testing"

	"github.com/lyft/datacatalog/pkg/common"
	"github.com/lyft/datacatalog/pkg/repositories/models"
	datacatalog "github.com/lyft/datacatalog/protos/gen"
	"github.com/stretchr/testify/assert"
)

func assertJoinExpression(t *testing.T, listInput models.ListModelsInput, joiningEntity common.Entity, sourceTableName string, joiningTableName string, expectedJoinStatement string) {
	joinCondition, ok := listInput.JoinEntityToConditionMap[joiningEntity]
	assert.True(t, ok)
	assert.Equal(t, joinCondition.GetJoiningDBEntity(), joiningEntity)
	expr, err := joinCondition.GetJoinOnDBQueryExpression(sourceTableName, joiningTableName)
	assert.NoError(t, err)
	assert.Equal(t, expr, expectedJoinStatement)
}

func TestListInputWithPartitionsAndTags(t *testing.T) {
	filter := &datacatalog.FilterExpression{
		Filters: []*datacatalog.SinglePropertyFilter{
			{
				PropertyFilter: &datacatalog.SinglePropertyFilter_PartitionFilter{
					PartitionFilter: &datacatalog.PartitionPropertyFilter{
						Property: &datacatalog.PartitionPropertyFilter_KeyVal{
							KeyVal: &datacatalog.KeyValuePair{Key: "key1", Value: "val1"},
						},
					},
				},
			},
			{
				PropertyFilter: &datacatalog.SinglePropertyFilter_PartitionFilter{
					PartitionFilter: &datacatalog.PartitionPropertyFilter{
						Property: &datacatalog.PartitionPropertyFilter_KeyVal{
							KeyVal: &datacatalog.KeyValuePair{Key: "key2", Value: "val2"},
						},
					},
				},
			},
			{
				PropertyFilter: &datacatalog.SinglePropertyFilter_TagFilter{
					TagFilter: &datacatalog.TagPropertyFilter{
						Property: &datacatalog.TagPropertyFilter_TagName{
							TagName: "special",
						},
					},
				},
			},
		},
	}
	listInput, err := FilterToListInput(context.Background(), common.Artifact, filter)
	assert.NoError(t, err)
	assert.Len(t, listInput.Filters, 5)                  // 2 for each partition filter, 1 for tag filter
	assert.Len(t, listInput.JoinEntityToConditionMap, 2) // join on partition and tag

	// even though there are 5 filters, there should only have entries for 2 joins
	assertJoinExpression(t, listInput, common.Partition, "artifacts", "partitions", "JOIN partitions ON artifacts.artifact_id = partitions.artifact_id")
	assertJoinExpression(t, listInput, common.Tag, "artifacts", "tags", "JOIN tags ON artifacts.artifact_id = tags.artifact_id")
}

func TestEmptyFiledListInput(t *testing.T) {
	filter := &datacatalog.FilterExpression{
		Filters: []*datacatalog.SinglePropertyFilter{
			{
				PropertyFilter: &datacatalog.SinglePropertyFilter_PartitionFilter{
					PartitionFilter: &datacatalog.PartitionPropertyFilter{
						Property: &datacatalog.PartitionPropertyFilter_KeyVal{
							KeyVal: &datacatalog.KeyValuePair{Key: "", Value: ""},
						},
					},
				},
			},
		},
	}
	_, err := FilterToListInput(context.Background(), common.Artifact, filter)
	assert.Error(t, err)
}
