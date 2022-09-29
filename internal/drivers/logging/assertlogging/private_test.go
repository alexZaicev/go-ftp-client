// This file is for exposing private functions and properties to support unit testing.

package assertlogging

import (
	"github.com/stretchr/testify/assert"
)

// AssertExpectations proxies the given parameters to Logger.assertExpectations for use in unit
// tests.
func (l *Logger) AssertExpectations(t assert.TestingT) bool {
	return l.assertExpectations(t)
}

// assertFieldValue runs the assert method on the given assertion with the provided parameters.
//
// The assert method on the FieldValue interface is currently unexported to reduce the public API of the package, but we still
// need a way to call it directly within this package's tests.
func AssertFieldValue(assertion FieldValue, t assert.TestingT, actual interface{}, field string, index int, level string) bool {
	return assertion.assert(t, actual, field, index, level)
}
