package gormimpl

import (
	"context"
	"database/sql/driver"
	"testing"

	"strings"

	mocket "github.com/Selvatico/go-mocket"
	"github.com/lyft/datacatalog/pkg/common"
	"github.com/lyft/datacatalog/pkg/repositories/models"
	"github.com/lyft/datacatalog/pkg/repositories/utils"
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

// Tag cannot be joined with partitions
func TestInvalidGormJoinCondition(t *testing.T) {
	filter := NewGormJoinCondition(common.Tag, common.Partition)

	_, err := filter.GetJoinOnDBQueryExpression("tags", "partitions")
	assert.Error(t, err)
}

func TestApplyFilter(t *testing.T) {
	testDB := utils.GetDbForTest(t)
	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	validInputApply := false

	GlobalMock.NewMock().WithQuery(
		`SELECT "artifacts".* FROM "artifacts"`).WithCallback(
		func(s string, values []driver.NamedValue) {
			// separate the regex matching because the joins reorder on different test runs
			validInputApply = strings.Contains(s, `JOIN tags ON artifacts.artifact_id = tags.artifact_id`) &&
				strings.Contains(s, `JOIN partitions ON artifacts.artifact_id = partitions.artifact_id`) &&
				strings.Contains(s, `WHERE "artifacts"."deleted_at" IS NULL AND `+
					`((partitions.key1 = val1) AND (partitions.key2 = val2) AND (tags.tag_name = special)) `+
					`LIMIT 10 OFFSET 10`)
		})

	listInput := models.ListModelsInput{
		JoinEntityToConditionMap: map[common.Entity]models.ModelJoinCondition{
			common.Partition: NewGormJoinCondition(common.Artifact, common.Partition),
			common.Tag:       NewGormJoinCondition(common.Artifact, common.Tag),
		},
		Filters: []models.ModelValueFilter{
			NewGormValueFilter(common.Partition, common.Equal, "key1", "val1"),
			NewGormValueFilter(common.Partition, common.Equal, "key2", "val2"),
			NewGormValueFilter(common.Tag, common.Equal, "tag_name", "special"),
		},
		Offset: 10,
		Limit:  10,
	}

	tx, err := applyListModelsInput(context.Background(), testDB, common.Artifact, listInput)
	assert.NoError(t, err)

	tx = tx.Find(models.Artifact{})
	assert.True(t, validInputApply)
}

func TestApplyFilterEmpty(t *testing.T) {
	testDB := utils.GetDbForTest(t)
	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	validInputApply := false

	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "artifacts"  WHERE "artifacts"."deleted_at" IS NULL LIMIT 10 OFFSET 10`).WithCallback(
		func(s string, values []driver.NamedValue) {
			// separate the regex matching because the joins reorder on different test runs
			validInputApply = true
		})

	listInput := models.ListModelsInput{
		JoinEntityToConditionMap: nil,
		Filters:                  nil,
		Offset:                   10,
		Limit:                    10,
	}

	tx, err := applyListModelsInput(context.Background(), testDB, common.Artifact, listInput)
	assert.NoError(t, err)

	tx = tx.Find(models.Artifact{})
	assert.True(t, validInputApply)
}
