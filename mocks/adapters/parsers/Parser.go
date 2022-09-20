// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	entities "github.com/alexZaicev/go-ftp-client/internal/domain/entities"
	mock "github.com/stretchr/testify/mock"
)

// Parser is an autogenerated mock type for the Parser type
type Parser struct {
	mock.Mock
}

// Parse provides a mock function with given fields: data
func (_m *Parser) Parse(data string) (*entities.Entry, error) {
	ret := _m.Called(data)

	var r0 *entities.Entry
	if rf, ok := ret.Get(0).(func(string) *entities.Entry); ok {
		r0 = rf(data)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*entities.Entry)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(data)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewParser interface {
	mock.TestingT
	Cleanup(func())
}

// NewParser creates a new instance of Parser. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewParser(t mockConstructorTestingTNewParser) *Parser {
	mock := &Parser{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}