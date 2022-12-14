// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	context "context"

	ftp "github.com/alexZaicev/go-ftp-client/internal/usecases/ftp"
	mock "github.com/stretchr/testify/mock"
)

// UploadFileUseCase is an autogenerated mock type for the UploadFileUseCase type
type UploadFileUseCase struct {
	mock.Mock
}

// Execute provides a mock function with given fields: _a0, _a1, _a2
func (_m *UploadFileUseCase) Execute(_a0 context.Context, _a1 *ftp.UploadFileRepos, _a2 *ftp.UploadFileInput) error {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *ftp.UploadFileRepos, *ftp.UploadFileInput) error); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewUploadFileUseCase interface {
	mock.TestingT
	Cleanup(func())
}

// NewUploadFileUseCase creates a new instance of UploadFileUseCase. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewUploadFileUseCase(t mockConstructorTestingTNewUploadFileUseCase) *UploadFileUseCase {
	mock := &UploadFileUseCase{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
