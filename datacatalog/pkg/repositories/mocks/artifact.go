// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import context "context"

import mock "github.com/stretchr/testify/mock"
import models "github.com/lyft/datacatalog/pkg/repositories/models"

// ArtifactRepo is an autogenerated mock type for the ArtifactRepo type
type ArtifactRepo struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, in
func (_m *ArtifactRepo) Create(ctx context.Context, in models.Artifact) error {
	ret := _m.Called(ctx, in)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, models.Artifact) error); ok {
		r0 = rf(ctx, in)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Get provides a mock function with given fields: ctx, in
func (_m *ArtifactRepo) Get(ctx context.Context, in models.ArtifactKey) (models.Artifact, error) {
	ret := _m.Called(ctx, in)

	var r0 models.Artifact
	if rf, ok := ret.Get(0).(func(context.Context, models.ArtifactKey) models.Artifact); ok {
		r0 = rf(ctx, in)
	} else {
		r0 = ret.Get(0).(models.Artifact)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, models.ArtifactKey) error); ok {
		r1 = rf(ctx, in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: ctx, datasetKey, in
func (_m *ArtifactRepo) List(ctx context.Context, datasetKey models.DatasetKey, in models.ListModelsInput) ([]models.Artifact, error) {
	ret := _m.Called(ctx, datasetKey, in)

	var r0 []models.Artifact
	if rf, ok := ret.Get(0).(func(context.Context, models.DatasetKey, models.ListModelsInput) []models.Artifact); ok {
		r0 = rf(ctx, datasetKey, in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]models.Artifact)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, models.DatasetKey, models.ListModelsInput) error); ok {
		r1 = rf(ctx, datasetKey, in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}