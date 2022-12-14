// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	http "net/http"
	testing "testing"

	mock "github.com/stretchr/testify/mock"
)

// Auth is an autogenerated mock type for the Auth type
type Auth struct {
	mock.Mock
}

// Init provides a mock function with given fields: port, local
func (_m *Auth) Init(port string, local bool) error {
	ret := _m.Called(port, local)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, bool) error); ok {
		r0 = rf(port, local)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// LoggedInAs provides a mock function with given fields: r
func (_m *Auth) LoggedInAs(r *http.Request) string {
	ret := _m.Called(r)

	var r0 string
	if rf, ok := ret.Get(0).(func(*http.Request) string); ok {
		r0 = rf(r)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// LoginURL provides a mock function with given fields: w, r
func (_m *Auth) LoginURL(w http.ResponseWriter, r *http.Request) string {
	ret := _m.Called(w, r)

	var r0 string
	if rf, ok := ret.Get(0).(func(http.ResponseWriter, *http.Request) string); ok {
		r0 = rf(w, r)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// NewAuth creates a new instance of Auth. It also registers a cleanup function to assert the mocks expectations.
func NewAuth(t testing.TB) *Auth {
	mock := &Auth{}

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
