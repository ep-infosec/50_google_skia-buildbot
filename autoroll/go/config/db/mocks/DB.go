// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	context "context"

	config "go.skia.org/infra/autoroll/go/config"

	mock "github.com/stretchr/testify/mock"

	testing "testing"
)

// DB is an autogenerated mock type for the DB type
type DB struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *DB) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Delete provides a mock function with given fields: ctx, rollerID
func (_m *DB) Delete(ctx context.Context, rollerID string) error {
	ret := _m.Called(ctx, rollerID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, rollerID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Get provides a mock function with given fields: ctx, rollerID
func (_m *DB) Get(ctx context.Context, rollerID string) (*config.Config, error) {
	ret := _m.Called(ctx, rollerID)

	var r0 *config.Config
	if rf, ok := ret.Get(0).(func(context.Context, string) *config.Config); ok {
		r0 = rf(ctx, rollerID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*config.Config)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, rollerID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAll provides a mock function with given fields: ctx
func (_m *DB) GetAll(ctx context.Context) ([]*config.Config, error) {
	ret := _m.Called(ctx)

	var r0 []*config.Config
	if rf, ok := ret.Get(0).(func(context.Context) []*config.Config); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*config.Config)
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

// Put provides a mock function with given fields: ctx, rollerID, cfg
func (_m *DB) Put(ctx context.Context, rollerID string, cfg *config.Config) error {
	ret := _m.Called(ctx, rollerID, cfg)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *config.Config) error); ok {
		r0 = rf(ctx, rollerID, cfg)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewDB creates a new instance of DB. It also registers a cleanup function to assert the mocks expectations.
func NewDB(t testing.TB) *DB {
	mock := &DB{}

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
