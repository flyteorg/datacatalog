package gormimpl

import (
	"context"

	"github.com/jinzhu/gorm"
	"github.com/lyft/datacatalog/pkg/common"
	"github.com/lyft/datacatalog/pkg/repositories/errors"
	"github.com/lyft/datacatalog/pkg/repositories/interfaces"
	"github.com/lyft/datacatalog/pkg/repositories/models"
	idl_datacatalog "github.com/lyft/datacatalog/protos/gen"
	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/promutils"
)

type tagRepo struct {
	db               *gorm.DB
	errorTransformer errors.ErrorTransformer
	repoMetrics      gormMetrics
}

func NewTagRepo(db *gorm.DB, errorTransformer errors.ErrorTransformer, scope promutils.Scope) interfaces.TagRepo {
	return &tagRepo{
		db:               db,
		errorTransformer: errorTransformer,
		repoMetrics:      newGormMetrics(scope),
	}
}

// A tag is associated with a single artifact for each partition combination
// When creating a tag, we remove the tag from any artifacts of the same partition
// Then add the tag to the new artifact
func (h *tagRepo) Create(ctx context.Context, tag models.Tag) error {
	timer := h.repoMetrics.CreateDuration.Start(ctx)
	defer timer.Stop()

	tx := h.db.Begin()

	var artifactToTag models.Artifact
	tx = tx.Preload("Partitions").Find(&artifactToTag, models.Artifact{
		ArtifactKey: models.ArtifactKey{ArtifactID: tag.ArtifactID},
	})

	// List artifacts with the same partitions and tag
	filters := make([]models.ModelValueFilter, 0, len(artifactToTag.Partitions)*2+1)
	for _, partition := range artifactToTag.Partitions {
		filters = append(filters, NewGormValueFilter(common.Partition, common.Equal, "key", partition.Key))
		filters = append(filters, NewGormValueFilter(common.Partition, common.Equal, "value", partition.Value))
	}

	filters = append(filters, NewGormValueFilter(common.Artifact, common.Equal, "tag_name", tag.TagName))

	listTaggedInput := models.ListModelsInput{
		JoinEntityToConditionMap: map[common.Entity]models.ModelJoinCondition{
			common.Tag:       NewGormJoinCondition(common.Artifact, common.Tag),
			common.Partition: NewGormJoinCondition(common.Artifact, common.Partition),
		},
		Filters: filters,
	}

	tx, err := applyListModelsInput(tx, common.Artifact, listTaggedInput)
	if err != nil {
		tx.Rollback()
		return err
	}

	var artifacts []models.Artifact
	tx = tx.Find(&artifacts)
	if tx.Error != nil {
		logger.Errorf(ctx, "Unable to find previously tagged artifacts, rolling back, tag: [%v], err [%v]", tag, tx.Error)
		tx.Rollback()
		return h.errorTransformer.ToDataCatalogError(tx.Error)
	}

	// if len(artifacts) != 0 {
	// 	// Soft-delete the existing tags on the artifacts that are tagged by this tag in the partition
	// 	for _, artifact := range artifacts {
	// 		oldTag := models.Tag{
	// 			TagKey:      models.TagKey{TagName: tag.TagName},
	// 			ArtifactID:  artifact.ArtifactID,
	// 			DatasetUUID: artifact.DatasetUUID,
	// 		}
	// 		tx = tx.Where(oldTag).Delete(&models.Tag{})
	// 	}
	// }

	// If the artifact was ever previously tagged with this tag, we need to
	// undelete the record because we cannot tag the artifact again since
	// the primary keys are the same.
	// var previouslyTagged *models.Artifact
	// tx = tx.Unscoped().Find(previouslyTagged, tag) // unscope will ignore deletedAt
	// if previouslyTagged != nil {
	// 	previouslyTagged.DeletedAt = nil
	// 	tx = tx.Update(previouslyTagged)
	// } else {
	// 	// Tag the new artifact
	// 	tx = tx.Create(&tag)
	// }

	tx = tx.Commit()
	if tx.Error != nil {
		logger.Errorf(ctx, "Unable to create tag, rolling back, tag: [%v], err [%v]", tag, tx.Error)
		tx.Rollback()
		return h.errorTransformer.ToDataCatalogError(tx.Error)
	}
	return nil
}

func (h *tagRepo) Get(ctx context.Context, in models.TagKey) (models.Tag, error) {
	timer := h.repoMetrics.GetDuration.Start(ctx)
	defer timer.Stop()

	var tag models.Tag
	result := h.db.Preload("Artifact").
		Preload("Artifact.ArtifactData").
		Preload("Artifact.Partitions", func(db *gorm.DB) *gorm.DB {
			return db.Order("partitions.created_at ASC") // preserve the order in which the partitions were created
		}).
		Preload("Artifact.Tags").
		Order("tags.created_at DESC").
		First(&tag, &models.Tag{
			TagKey: in,
		})

	if result.Error != nil {
		return models.Tag{}, h.errorTransformer.ToDataCatalogError(result.Error)
	}
	if result.RecordNotFound() {
		return models.Tag{}, errors.GetMissingEntityError("Tag", &idl_datacatalog.Tag{
			Name: tag.TagName,
		})
	}

	return tag, nil
}
