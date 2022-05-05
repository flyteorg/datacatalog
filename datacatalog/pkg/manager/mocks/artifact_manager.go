// Code generated by mockery v1.0.1. DO NOT EDIT.

package mocks

import (
	context "context"

	datacatalog "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"

	mock "github.com/stretchr/testify/mock"
)

// ArtifactManager is an autogenerated mock type for the ArtifactManager type
type ArtifactManager struct {
	mock.Mock
}

type ArtifactManager_CreateArtifact struct {
	*mock.Call
}

func (_m ArtifactManager_CreateArtifact) Return(_a0 *datacatalog.CreateArtifactResponse, _a1 error) *ArtifactManager_CreateArtifact {
	return &ArtifactManager_CreateArtifact{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *ArtifactManager) OnCreateArtifact(ctx context.Context, request *datacatalog.CreateArtifactRequest) *ArtifactManager_CreateArtifact {
	c := _m.On("CreateArtifact", ctx, request)
	return &ArtifactManager_CreateArtifact{Call: c}
}

func (_m *ArtifactManager) OnCreateArtifactMatch(matchers ...interface{}) *ArtifactManager_CreateArtifact {
	c := _m.On("CreateArtifact", matchers...)
	return &ArtifactManager_CreateArtifact{Call: c}
}

// CreateArtifact provides a mock function with given fields: ctx, request
func (_m *ArtifactManager) CreateArtifact(ctx context.Context, request *datacatalog.CreateArtifactRequest) (*datacatalog.CreateArtifactResponse, error) {
	ret := _m.Called(ctx, request)

	var r0 *datacatalog.CreateArtifactResponse
	if rf, ok := ret.Get(0).(func(context.Context, *datacatalog.CreateArtifactRequest) *datacatalog.CreateArtifactResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*datacatalog.CreateArtifactResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *datacatalog.CreateArtifactRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type ArtifactManager_GetArtifact struct {
	*mock.Call
}

func (_m ArtifactManager_GetArtifact) Return(_a0 *datacatalog.GetArtifactResponse, _a1 error) *ArtifactManager_GetArtifact {
	return &ArtifactManager_GetArtifact{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *ArtifactManager) OnGetArtifact(ctx context.Context, request *datacatalog.GetArtifactRequest) *ArtifactManager_GetArtifact {
	c := _m.On("GetArtifact", ctx, request)
	return &ArtifactManager_GetArtifact{Call: c}
}

func (_m *ArtifactManager) OnGetArtifactMatch(matchers ...interface{}) *ArtifactManager_GetArtifact {
	c := _m.On("GetArtifact", matchers...)
	return &ArtifactManager_GetArtifact{Call: c}
}

// GetArtifact provides a mock function with given fields: ctx, request
func (_m *ArtifactManager) GetArtifact(ctx context.Context, request *datacatalog.GetArtifactRequest) (*datacatalog.GetArtifactResponse, error) {
	ret := _m.Called(ctx, request)

	var r0 *datacatalog.GetArtifactResponse
	if rf, ok := ret.Get(0).(func(context.Context, *datacatalog.GetArtifactRequest) *datacatalog.GetArtifactResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*datacatalog.GetArtifactResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *datacatalog.GetArtifactRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type ArtifactManager_ListArtifacts struct {
	*mock.Call
}

func (_m ArtifactManager_ListArtifacts) Return(_a0 *datacatalog.ListArtifactsResponse, _a1 error) *ArtifactManager_ListArtifacts {
	return &ArtifactManager_ListArtifacts{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *ArtifactManager) OnListArtifacts(ctx context.Context, request *datacatalog.ListArtifactsRequest) *ArtifactManager_ListArtifacts {
	c := _m.On("ListArtifacts", ctx, request)
	return &ArtifactManager_ListArtifacts{Call: c}
}

func (_m *ArtifactManager) OnListArtifactsMatch(matchers ...interface{}) *ArtifactManager_ListArtifacts {
	c := _m.On("ListArtifacts", matchers...)
	return &ArtifactManager_ListArtifacts{Call: c}
}

// ListArtifacts provides a mock function with given fields: ctx, request
func (_m *ArtifactManager) ListArtifacts(ctx context.Context, request *datacatalog.ListArtifactsRequest) (*datacatalog.ListArtifactsResponse, error) {
	ret := _m.Called(ctx, request)

	var r0 *datacatalog.ListArtifactsResponse
	if rf, ok := ret.Get(0).(func(context.Context, *datacatalog.ListArtifactsRequest) *datacatalog.ListArtifactsResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*datacatalog.ListArtifactsResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *datacatalog.ListArtifactsRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}