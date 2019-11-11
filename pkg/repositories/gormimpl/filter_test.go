package gormimpl

import (
	"testing"

	"github.com/lyft/datacatalog/pkg/common"
	"github.com/stretchr/testify/assert"
)

func TestGormValueFilter(t *testing.T) {
	filter := NewGormValueFilter(common.Partition, common.Equal, "key", "region")
	expression, err := filter.GetDBQueryExpression("partitions")
	assert.NoError(t, err)
	assert.Equal(t, filter.GetDBEntity(), common.Partition)
	assert.Equal(t, expression.Query, "partitions.key = ?")
	assert.Equal(t, expression.Args, "region")
}

func TestGormJoinCondition(t *testing.T) {
	filter := NewGormJoinCondition(common.Artifact, common.Partition)
	assert.Equal(t, filter.GetJoiningDBEntity(), common.Partition)

	joinQuery, err := filter.GetJoinOnDBQueryExpression("artifacts", "partitions")
	assert.NoError(t, err)
	assert.Equal(t, joinQuery, "JOIN partitions ON artifacts.artifact_id = partitions.artifact_id")
}

func TestInvalidGormJoinCondition(t *testing.T) {
	filter := NewGormJoinCondition(common.Tag, common.Partition)

	_, err := filter.GetJoinOnDBQueryExpression("tags", "partitions")
	assert.Error(t, err)
}
