// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	context "context"
	testing "testing"

	mock "github.com/stretchr/testify/mock"

	time "time"

	types "go.skia.org/infra/bugs-central/go/types"
)

// BugsDB is an autogenerated mock type for the BugsDB type
type BugsDB struct {
	mock.Mock
}

// GenerateRunId provides a mock function with given fields: ts
func (_m *BugsDB) GenerateRunId(ts time.Time) string {
	ret := _m.Called(ts)

	var r0 string
	if rf, ok := ret.Get(0).(func(time.Time) string); ok {
		r0 = rf(ts)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetAllRecognizedRunIds provides a mock function with given fields: ctx
func (_m *BugsDB) GetAllRecognizedRunIds(ctx context.Context) (map[string]bool, error) {
	ret := _m.Called(ctx)

	var r0 map[string]bool
	if rf, ok := ret.Get(0).(func(context.Context) map[string]bool); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]bool)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetClientsFromDB provides a mock function with given fields: ctx
func (_m *BugsDB) GetClientsFromDB(ctx context.Context) (map[types.RecognizedClient]map[types.IssueSource]map[string]bool, error) {
	ret := _m.Called(ctx)

	var r0 map[types.RecognizedClient]map[types.IssueSource]map[string]bool
	if rf, ok := ret.Get(0).(func(context.Context) map[types.RecognizedClient]map[types.IssueSource]map[string]bool); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[types.RecognizedClient]map[types.IssueSource]map[string]bool)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetCountsFromDB provides a mock function with given fields: ctx, client, source, query
func (_m *BugsDB) GetCountsFromDB(ctx context.Context, client types.RecognizedClient, source types.IssueSource, query string) (*types.IssueCountsData, error) {
	ret := _m.Called(ctx, client, source, query)

	var r0 *types.IssueCountsData
	if rf, ok := ret.Get(0).(func(context.Context, types.RecognizedClient, types.IssueSource, string) *types.IssueCountsData); ok {
		r0 = rf(ctx, client, source, query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.IssueCountsData)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, types.RecognizedClient, types.IssueSource, string) error); ok {
		r1 = rf(ctx, client, source, query)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetQueryDataFromDB provides a mock function with given fields: ctx, client, source, query
func (_m *BugsDB) GetQueryDataFromDB(ctx context.Context, client types.RecognizedClient, source types.IssueSource, query string) ([]*types.QueryData, error) {
	ret := _m.Called(ctx, client, source, query)

	var r0 []*types.QueryData
	if rf, ok := ret.Get(0).(func(context.Context, types.RecognizedClient, types.IssueSource, string) []*types.QueryData); ok {
		r0 = rf(ctx, client, source, query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*types.QueryData)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, types.RecognizedClient, types.IssueSource, string) error); ok {
		r1 = rf(ctx, client, source, query)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PutInDB provides a mock function with given fields: ctx, client, source, query, runId, countsData
func (_m *BugsDB) PutInDB(ctx context.Context, client types.RecognizedClient, source types.IssueSource, query string, runId string, countsData *types.IssueCountsData) error {
	ret := _m.Called(ctx, client, source, query, runId, countsData)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, types.RecognizedClient, types.IssueSource, string, string, *types.IssueCountsData) error); ok {
		r0 = rf(ctx, client, source, query, runId, countsData)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StoreRunId provides a mock function with given fields: ctx, runId
func (_m *BugsDB) StoreRunId(ctx context.Context, runId string) error {
	ret := _m.Called(ctx, runId)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, runId)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewBugsDB creates a new instance of BugsDB. It also registers a cleanup function to assert the mocks expectations.
func NewBugsDB(t testing.TB) *BugsDB {
	mock := &BugsDB{}

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
