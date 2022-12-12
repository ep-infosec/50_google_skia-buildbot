// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	clientcipd "go.chromium.org/luci/cipd/client/cipd"
	cipd "go.skia.org/infra/go/cipd"

	common "go.chromium.org/luci/cipd/common"

	context "context"

	deployer "go.chromium.org/luci/cipd/client/cipd/deployer"

	io "io"

	mock "github.com/stretchr/testify/mock"

	pkg "go.chromium.org/luci/cipd/client/cipd/pkg"

	regexp "regexp"

	testing "testing"

	time "time"
)

// CIPDClient is an autogenerated mock type for the CIPDClient type
type CIPDClient struct {
	mock.Mock
}

// Attach provides a mock function with given fields: ctx, pin, refs, tags, metadata
func (_m *CIPDClient) Attach(ctx context.Context, pin common.Pin, refs []string, tags []string, metadata map[string]string) error {
	ret := _m.Called(ctx, pin, refs, tags, metadata)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, common.Pin, []string, []string, map[string]string) error); ok {
		r0 = rf(ctx, pin, refs, tags, metadata)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// AttachMetadataWhenReady provides a mock function with given fields: ctx, pin, md
func (_m *CIPDClient) AttachMetadataWhenReady(ctx context.Context, pin common.Pin, md []clientcipd.Metadata) error {
	ret := _m.Called(ctx, pin, md)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, common.Pin, []clientcipd.Metadata) error); ok {
		r0 = rf(ctx, pin, md)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// AttachTagsWhenReady provides a mock function with given fields: ctx, pin, tags
func (_m *CIPDClient) AttachTagsWhenReady(ctx context.Context, pin common.Pin, tags []string) error {
	ret := _m.Called(ctx, pin, tags)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, common.Pin, []string) error); ok {
		r0 = rf(ctx, pin, tags)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// BeginBatch provides a mock function with given fields: ctx
func (_m *CIPDClient) BeginBatch(ctx context.Context) {
	_m.Called(ctx)
}

// CheckDeployment provides a mock function with given fields: ctx, paranoia
func (_m *CIPDClient) CheckDeployment(ctx context.Context, paranoia deployer.ParanoidMode) (clientcipd.ActionMap, error) {
	ret := _m.Called(ctx, paranoia)

	var r0 clientcipd.ActionMap
	if rf, ok := ret.Get(0).(func(context.Context, deployer.ParanoidMode) clientcipd.ActionMap); ok {
		r0 = rf(ctx, paranoia)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(clientcipd.ActionMap)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, deployer.ParanoidMode) error); ok {
		r1 = rf(ctx, paranoia)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Create provides a mock function with given fields: ctx, name, dir, installMode, excludeMatchingFiles, refs, tags, metadata
func (_m *CIPDClient) Create(ctx context.Context, name string, dir string, installMode pkg.InstallMode, excludeMatchingFiles []*regexp.Regexp, refs []string, tags []string, metadata map[string]string) (common.Pin, error) {
	ret := _m.Called(ctx, name, dir, installMode, excludeMatchingFiles, refs, tags, metadata)

	var r0 common.Pin
	if rf, ok := ret.Get(0).(func(context.Context, string, string, pkg.InstallMode, []*regexp.Regexp, []string, []string, map[string]string) common.Pin); ok {
		r0 = rf(ctx, name, dir, installMode, excludeMatchingFiles, refs, tags, metadata)
	} else {
		r0 = ret.Get(0).(common.Pin)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, pkg.InstallMode, []*regexp.Regexp, []string, []string, map[string]string) error); ok {
		r1 = rf(ctx, name, dir, installMode, excludeMatchingFiles, refs, tags, metadata)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Describe provides a mock function with given fields: ctx, _a1, instance
func (_m *CIPDClient) Describe(ctx context.Context, _a1 string, instance string) (*clientcipd.InstanceDescription, error) {
	ret := _m.Called(ctx, _a1, instance)

	var r0 *clientcipd.InstanceDescription
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *clientcipd.InstanceDescription); ok {
		r0 = rf(ctx, _a1, instance)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*clientcipd.InstanceDescription)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, _a1, instance)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DescribeClient provides a mock function with given fields: ctx, pin
func (_m *CIPDClient) DescribeClient(ctx context.Context, pin common.Pin) (*clientcipd.ClientDescription, error) {
	ret := _m.Called(ctx, pin)

	var r0 *clientcipd.ClientDescription
	if rf, ok := ret.Get(0).(func(context.Context, common.Pin) *clientcipd.ClientDescription); ok {
		r0 = rf(ctx, pin)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*clientcipd.ClientDescription)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, common.Pin) error); ok {
		r1 = rf(ctx, pin)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DescribeInstance provides a mock function with given fields: ctx, pin, opts
func (_m *CIPDClient) DescribeInstance(ctx context.Context, pin common.Pin, opts *clientcipd.DescribeInstanceOpts) (*clientcipd.InstanceDescription, error) {
	ret := _m.Called(ctx, pin, opts)

	var r0 *clientcipd.InstanceDescription
	if rf, ok := ret.Get(0).(func(context.Context, common.Pin, *clientcipd.DescribeInstanceOpts) *clientcipd.InstanceDescription); ok {
		r0 = rf(ctx, pin, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*clientcipd.InstanceDescription)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, common.Pin, *clientcipd.DescribeInstanceOpts) error); ok {
		r1 = rf(ctx, pin, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EndBatch provides a mock function with given fields: ctx
func (_m *CIPDClient) EndBatch(ctx context.Context) {
	_m.Called(ctx)
}

// Ensure provides a mock function with given fields: ctx, packages
func (_m *CIPDClient) Ensure(ctx context.Context, packages ...*cipd.Package) error {
	_va := make([]interface{}, len(packages))
	for _i := range packages {
		_va[_i] = packages[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, ...*cipd.Package) error); ok {
		r0 = rf(ctx, packages...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// EnsurePackages provides a mock function with given fields: ctx, pkgs, paranoia, maxThreads, dryRun
func (_m *CIPDClient) EnsurePackages(ctx context.Context, pkgs common.PinSliceBySubdir, paranoia deployer.ParanoidMode, maxThreads int, dryRun bool) (clientcipd.ActionMap, error) {
	ret := _m.Called(ctx, pkgs, paranoia, maxThreads, dryRun)

	var r0 clientcipd.ActionMap
	if rf, ok := ret.Get(0).(func(context.Context, common.PinSliceBySubdir, deployer.ParanoidMode, int, bool) clientcipd.ActionMap); ok {
		r0 = rf(ctx, pkgs, paranoia, maxThreads, dryRun)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(clientcipd.ActionMap)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, common.PinSliceBySubdir, deployer.ParanoidMode, int, bool) error); ok {
		r1 = rf(ctx, pkgs, paranoia, maxThreads, dryRun)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FetchACL provides a mock function with given fields: ctx, prefix
func (_m *CIPDClient) FetchACL(ctx context.Context, prefix string) ([]clientcipd.PackageACL, error) {
	ret := _m.Called(ctx, prefix)

	var r0 []clientcipd.PackageACL
	if rf, ok := ret.Get(0).(func(context.Context, string) []clientcipd.PackageACL); ok {
		r0 = rf(ctx, prefix)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]clientcipd.PackageACL)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, prefix)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FetchAndDeployInstance provides a mock function with given fields: ctx, subdir, pin, maxThreads
func (_m *CIPDClient) FetchAndDeployInstance(ctx context.Context, subdir string, pin common.Pin, maxThreads int) error {
	ret := _m.Called(ctx, subdir, pin, maxThreads)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, common.Pin, int) error); ok {
		r0 = rf(ctx, subdir, pin, maxThreads)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// FetchInstance provides a mock function with given fields: ctx, pin
func (_m *CIPDClient) FetchInstance(ctx context.Context, pin common.Pin) (pkg.Source, error) {
	ret := _m.Called(ctx, pin)

	var r0 pkg.Source
	if rf, ok := ret.Get(0).(func(context.Context, common.Pin) pkg.Source); ok {
		r0 = rf(ctx, pin)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(pkg.Source)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, common.Pin) error); ok {
		r1 = rf(ctx, pin)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FetchInstanceTo provides a mock function with given fields: ctx, pin, output
func (_m *CIPDClient) FetchInstanceTo(ctx context.Context, pin common.Pin, output io.WriteSeeker) error {
	ret := _m.Called(ctx, pin, output)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, common.Pin, io.WriteSeeker) error); ok {
		r0 = rf(ctx, pin, output)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// FetchPackageRefs provides a mock function with given fields: ctx, packageName
func (_m *CIPDClient) FetchPackageRefs(ctx context.Context, packageName string) ([]clientcipd.RefInfo, error) {
	ret := _m.Called(ctx, packageName)

	var r0 []clientcipd.RefInfo
	if rf, ok := ret.Get(0).(func(context.Context, string) []clientcipd.RefInfo); ok {
		r0 = rf(ctx, packageName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]clientcipd.RefInfo)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, packageName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FetchRoles provides a mock function with given fields: ctx, prefix
func (_m *CIPDClient) FetchRoles(ctx context.Context, prefix string) ([]string, error) {
	ret := _m.Called(ctx, prefix)

	var r0 []string
	if rf, ok := ret.Get(0).(func(context.Context, string) []string); ok {
		r0 = rf(ctx, prefix)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, prefix)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListInstances provides a mock function with given fields: ctx, packageName
func (_m *CIPDClient) ListInstances(ctx context.Context, packageName string) (clientcipd.InstanceEnumerator, error) {
	ret := _m.Called(ctx, packageName)

	var r0 clientcipd.InstanceEnumerator
	if rf, ok := ret.Get(0).(func(context.Context, string) clientcipd.InstanceEnumerator); ok {
		r0 = rf(ctx, packageName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(clientcipd.InstanceEnumerator)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, packageName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListPackages provides a mock function with given fields: ctx, prefix, recursive, includeHidden
func (_m *CIPDClient) ListPackages(ctx context.Context, prefix string, recursive bool, includeHidden bool) ([]string, error) {
	ret := _m.Called(ctx, prefix, recursive, includeHidden)

	var r0 []string
	if rf, ok := ret.Get(0).(func(context.Context, string, bool, bool) []string); ok {
		r0 = rf(ctx, prefix, recursive, includeHidden)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, bool, bool) error); ok {
		r1 = rf(ctx, prefix, recursive, includeHidden)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ModifyACL provides a mock function with given fields: ctx, prefix, changes
func (_m *CIPDClient) ModifyACL(ctx context.Context, prefix string, changes []clientcipd.PackageACLChange) error {
	ret := _m.Called(ctx, prefix, changes)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, []clientcipd.PackageACLChange) error); ok {
		r0 = rf(ctx, prefix, changes)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RegisterInstance provides a mock function with given fields: ctx, pin, body, timeout
func (_m *CIPDClient) RegisterInstance(ctx context.Context, pin common.Pin, body io.ReadSeeker, timeout time.Duration) error {
	ret := _m.Called(ctx, pin, body, timeout)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, common.Pin, io.ReadSeeker, time.Duration) error); ok {
		r0 = rf(ctx, pin, body, timeout)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RepairDeployment provides a mock function with given fields: ctx, paranoia, maxThreads
func (_m *CIPDClient) RepairDeployment(ctx context.Context, paranoia deployer.ParanoidMode, maxThreads int) (clientcipd.ActionMap, error) {
	ret := _m.Called(ctx, paranoia, maxThreads)

	var r0 clientcipd.ActionMap
	if rf, ok := ret.Get(0).(func(context.Context, deployer.ParanoidMode, int) clientcipd.ActionMap); ok {
		r0 = rf(ctx, paranoia, maxThreads)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(clientcipd.ActionMap)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, deployer.ParanoidMode, int) error); ok {
		r1 = rf(ctx, paranoia, maxThreads)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ResolveVersion provides a mock function with given fields: ctx, packageName, version
func (_m *CIPDClient) ResolveVersion(ctx context.Context, packageName string, version string) (common.Pin, error) {
	ret := _m.Called(ctx, packageName, version)

	var r0 common.Pin
	if rf, ok := ret.Get(0).(func(context.Context, string, string) common.Pin); ok {
		r0 = rf(ctx, packageName, version)
	} else {
		r0 = ret.Get(0).(common.Pin)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, packageName, version)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SearchInstances provides a mock function with given fields: ctx, packageName, tags
func (_m *CIPDClient) SearchInstances(ctx context.Context, packageName string, tags []string) (common.PinSlice, error) {
	ret := _m.Called(ctx, packageName, tags)

	var r0 common.PinSlice
	if rf, ok := ret.Get(0).(func(context.Context, string, []string) common.PinSlice); ok {
		r0 = rf(ctx, packageName, tags)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(common.PinSlice)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, []string) error); ok {
		r1 = rf(ctx, packageName, tags)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetRefWhenReady provides a mock function with given fields: ctx, ref, pin
func (_m *CIPDClient) SetRefWhenReady(ctx context.Context, ref string, pin common.Pin) error {
	ret := _m.Called(ctx, ref, pin)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, common.Pin) error); ok {
		r0 = rf(ctx, ref, pin)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewCIPDClient creates a new instance of CIPDClient. It also registers a cleanup function to assert the mocks expectations.
func NewCIPDClient(t testing.TB) *CIPDClient {
	mock := &CIPDClient{}

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
