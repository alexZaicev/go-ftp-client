// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	context "context"

	connection "github.com/alexZaicev/go-ftp-client/internal/domain/connection"

	entities "github.com/alexZaicev/go-ftp-client/internal/domain/entities"

	mock "github.com/stretchr/testify/mock"
)

// Connection is an autogenerated mock type for the Connection type
type Connection struct {
	mock.Mock
}

// Cd provides a mock function with given fields: path
func (_m *Connection) Cd(path string) error {
	ret := _m.Called(path)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(path)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// EnableExplicitTLSMode provides a mock function with given fields:
func (_m *Connection) EnableExplicitTLSMode() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// List provides a mock function with given fields: ctx, options
func (_m *Connection) List(ctx context.Context, options *connection.ListOptions) ([]*entities.Entry, error) {
	ret := _m.Called(ctx, options)

	var r0 []*entities.Entry
	if rf, ok := ret.Get(0).(func(context.Context, *connection.ListOptions) []*entities.Entry); ok {
		r0 = rf(ctx, options)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*entities.Entry)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *connection.ListOptions) error); ok {
		r1 = rf(ctx, options)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Login provides a mock function with given fields: user, password
func (_m *Connection) Login(user string, password string) error {
	ret := _m.Called(user, password)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(user, password)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Mkdir provides a mock function with given fields: path
func (_m *Connection) Mkdir(path string) error {
	ret := _m.Called(path)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(path)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Move provides a mock function with given fields: oldPath, newPath
func (_m *Connection) Move(oldPath string, newPath string) error {
	ret := _m.Called(oldPath, newPath)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(oldPath, newPath)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Ready provides a mock function with given fields:
func (_m *Connection) Ready() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemoveDir provides a mock function with given fields: path
func (_m *Connection) RemoveDir(path string) error {
	ret := _m.Called(path)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(path)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemoveFile provides a mock function with given fields: path
func (_m *Connection) RemoveFile(path string) error {
	ret := _m.Called(path)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(path)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Size provides a mock function with given fields: path
func (_m *Connection) Size(path string) (uint64, error) {
	ret := _m.Called(path)

	var r0 uint64
	if rf, ok := ret.Get(0).(func(string) uint64); ok {
		r0 = rf(path)
	} else {
		r0 = ret.Get(0).(uint64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(path)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Status provides a mock function with given fields:
func (_m *Connection) Status() (*entities.Status, error) {
	ret := _m.Called()

	var r0 *entities.Status
	if rf, ok := ret.Get(0).(func() *entities.Status); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*entities.Status)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Stop provides a mock function with given fields:
func (_m *Connection) Stop() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Upload provides a mock function with given fields: ctx, options
func (_m *Connection) Upload(ctx context.Context, options *connection.UploadOptions) error {
	ret := _m.Called(ctx, options)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *connection.UploadOptions) error); ok {
		r0 = rf(ctx, options)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewConnection interface {
	mock.TestingT
	Cleanup(func())
}

// NewConnection creates a new instance of Connection. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewConnection(t mockConstructorTestingTNewConnection) *Connection {
	mock := &Connection{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
