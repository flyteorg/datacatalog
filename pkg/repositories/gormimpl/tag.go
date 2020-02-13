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

	// There are several steps that need to be done in a transaction in order for tag stealing to occur
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Find the set of partitions this artifact belongs to
	var artifactToTag models.Artifact
	tx.Preload("Partitions").Find(&artifactToTag, models.Artifact{
		ArtifactKey: models.ArtifactKey{ArtifactID: tag.ArtifactID},
	})

	// 2. List artifacts in the partitions that are currently tagged
	modelFilters := make([]models.ModelFilter, 0, len(artifactToTag.Partitions)+2)
	for _, partition := range artifactToTag.Partitions {
		modelFilters = append(modelFilters, models.ModelFilter{
			Entity: common.Partition,
			ValueFilters: []models.ModelValueFilter{
				NewGormValueFilter(common.Equal, "key", partition.Key),
				NewGormValueFilter(common.Equal, "value", partition.Value),
			},
			JoinCondition: NewGormJoinCondition(common.Artifact, common.Partition),
		})
	}

	modelFilters = append(modelFilters, models.ModelFilter{
		Entity: common.Tag,
		ValueFilters: []models.ModelValueFilter{
			NewGormValueFilter(common.Equal, "tag_name", tag.TagName),
			NewGormValueFilter(common.Equal, "deleted_at", gorm.Expr("NULL")), // AC: this may not work, may have to specially handle nil
		},
		JoinCondition: NewGormJoinCondition(common.Artifact, common.Tag),
	})

	listTaggedInput := models.ListModelsInput{
		ModelFilters: modelFilters,
		Limit:        100,
	}

	listArtifactsScope, err := applyListModelsInput(tx, common.Artifact, listTaggedInput)
	if err != nil {
		logger.Errorf(ctx, "Unable to construct artiact list, rolling back, tag: [%v], err [%v]", tag, tx.Error)
		tx.Rollback()
		return h.errorTransformer.ToDataCatalogError(err)

	}

	var artifacts []models.Artifact
	if listArtifactsScope.Find(&artifacts).Error != nil {
		logger.Errorf(ctx, "Unable to find previously tagged artifacts, rolling back, tag: [%v], err [%v]", tag, listArtifactsScope.Error)
		tx.Rollback()
		return h.errorTransformer.ToDataCatalogError(listArtifactsScope.Error)
	}

	// 3. Remove the tags from the currently tagged artifacts
	if len(artifacts) != 0 {
		// Soft-delete the existing tags on the artifacts that are currently tagged
		for _, artifact := range artifacts {

			// if the artifact to tag is already tagged, no need to remove it
			if artifactToTag.ArtifactID != artifact.ArtifactID {
				oldTag := models.Tag{
					TagKey:      models.TagKey{TagName: tag.TagName},
					ArtifactID:  artifact.ArtifactID,
					DatasetUUID: artifact.DatasetUUID,
				}
				deleteScope := tx.NewScope(&models.Tag{}).DB().Delete(&models.Tag{}, oldTag)
				if deleteScope.Error != nil {
					logger.Errorf(ctx, "Unable to delete previously tagged artifacts, rolling back, tag: [%v], err [%v]", tag, deleteScope.Error)
					tx.Rollback()
					return h.errorTransformer.ToDataCatalogError(deleteScope.Error)
				}
			}
		}
	}

	// 4. If the artifact was ever previously tagged with this tag, we need to
	// un-delete the record because we cannot tag the artifact again since
	// the primary keys are the same.
	undeleteScope := tx.Unscoped().Model(&tag).Update("deleted_at", gorm.Expr("NULL")) // unscope will ignore deletedAt
	if undeleteScope.Error != nil {
		logger.Errorf(ctx, "Unable to undelete tag tag, rolling back, tag: [%v], err [%v]", tag, tx.Error)
		tx.Rollback()
		return h.errorTransformer.ToDataCatalogError(tx.Error)
	}

	// 5. Tag the new artifact
	if undeleteScope.RowsAffected == 0 {
		if err := tx.Create(&tag).Error; err != nil {
			logger.Errorf(ctx, "Unable to create tag, rolling back, tag: [%v], err [%v]", tag, err)
			tx.Rollback()
			return h.errorTransformer.ToDataCatalogError(err)
		}
	}

	tx = tx.Commit()
	if tx.Error != nil {
		logger.Errorf(ctx, "Unable to commit transaction, rolling back, tag: [%v], err [%v]", tag, tx.Error)
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
