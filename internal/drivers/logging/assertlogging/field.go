package assertlogging

import (
	"fmt"
	"regexp"

	"github.com/stretchr/testify/assert"
)

const (
	fieldBaseMsg       = "field %q of log at index %d of level %s"
	fieldEqualMsg      = fieldBaseMsg + " does not match"
	fieldNotEqualMsg   = fieldBaseMsg + " matches"
	fieldErrorTypeMsg  = fieldBaseMsg + " is not an error"
	fieldEqualErrorMsg = fieldBaseMsg + " does not match the expected error message"
	fieldIsTypeMsg     = fieldBaseMsg + " is not the expected type"
	fieldRegexpTypeMsg = fieldBaseMsg + " is not a string"
	fieldRegexpMsg     = fieldBaseMsg + " does not match the regexp %q: %s"
	fieldBoolTypeMsg   = fieldBaseMsg + " is not a bool"
	fieldTrueMsg       = fieldBaseMsg + " is not true"
	fieldFalseMsg      = fieldBaseMsg + " is not false"
	fieldEmptyMsg      = fieldBaseMsg + " is not empty"
	fieldNilMsg        = fieldBaseMsg + " is not nil"
)

// FieldValue is used to assert that a log field value matches a given expectation.
//
// There is currently no "Any" implementation to allow any value to be present in a log field,
// partly to encourage good behavior when asserting against log fields, but also to help identify
// new Value implementations to cover additional use cases.
type FieldValue interface {
	fmt.Stringer
	// assert method is unexported to reduce the public API of the package
	assert(t assert.TestingT, actual interface{}, field string, index int, level string) bool
}

// fieldEqual asserts that a log field value is equal to the given expected value.
type fieldEqual struct {
	expected interface{}
}

func (f fieldEqual) assert(t assert.TestingT, actual interface{}, field string, index int, level string) bool {
	return assert.Equal(t, f.expected, actual, fieldEqualMsg, field, index, level)
}

func (f fieldEqual) String() string {
	return fmt.Sprintf("Equal:%+v", f.expected)
}

// Equal asserts that a log field value is equal to the given expected value.
//
//nolint:revive // private type is passed back into public functions so is safe to ignore
func Equal(expected interface{}) fieldEqual {
	return fieldEqual{expected: expected}
}

// Equalf asserts that a log field value is equal to the given expected string value with arguments
// substituted in.
//
//nolint:revive // private type is passed back into public functions so is safe to ignore
func Equalf(expected string, args ...interface{}) fieldEqual {
	expected = fmt.Sprintf(expected, args...)
	return Equal(expected)
}

// fieldNotEqual asserts that a log field value is not equal to the given expected value.
type fieldNotEqual struct {
	expected interface{}
}

func (f fieldNotEqual) assert(t assert.TestingT, actual interface{}, field string, index int, level string) bool {
	return assert.NotEqual(t, f.expected, actual, fieldNotEqualMsg, field, index, level)
}

func (f fieldNotEqual) String() string {
	return fmt.Sprintf("NotEqual:%+v", f.expected)
}

// NotEqual asserts that a log field value is not equal to the given expected value.
//
//nolint:revive // private type is passed back into public functions so is safe to ignore
func NotEqual(expected interface{}) fieldNotEqual {
	return fieldNotEqual{expected: expected}
}

// fieldEqualError asserts that a log field value is an error with a message equal to the given expected
// value.
type fieldEqualError struct {
	expected string
}

func (f fieldEqualError) assert(t assert.TestingT, actual interface{}, field string, index int, level string) bool {
	a, ok := actual.(error)
	if assert.True(t, ok, fieldErrorTypeMsg, field, index, level) {
		return assert.EqualError(t, a, f.expected, fieldEqualErrorMsg, field, index, level)
	}

	return false
}

func (f fieldEqualError) String() string {
	return fmt.Sprintf("EqualError:%+v", f.expected)
}

// EqualError asserts that a log field value is an error with a message equal to the given expected
// value.
//
//nolint:revive // private type is passed back into public functions so is safe to ignore
func EqualError(expected string) fieldEqualError {
	return fieldEqualError{expected: expected}
}

// EqualErrorf asserts that a log field value is an error with a message equal to the given expected
// value with arguments substituted in.
//
//nolint:revive // private type is passed back into public functions so is safe to ignore
func EqualErrorf(expected string, args ...interface{}) fieldEqualError {
	expected = fmt.Sprintf(expected, args...)
	return EqualError(expected)
}

// fieldIsType asserts that a log field value matched the type of the given expected value.
type fieldIsType struct {
	expected interface{}
}

func (f fieldIsType) assert(t assert.TestingT, actual interface{}, field string, index int, level string) bool {
	return assert.IsType(t, f.expected, actual, fieldIsTypeMsg, field, index, level)
}

func (f fieldIsType) String() string {
	return fmt.Sprintf("IsType:%T", f.expected)
}

// IsType asserts that a log field value matched the type of the given expected value.
//
//nolint:revive // private type is passed back into public functions so is safe to ignore
func IsType(expected interface{}) fieldIsType {
	return fieldIsType{expected: expected}
}

// fieldRegexp asserts that a log field value is a string that matches the given regular expression.
type fieldRegexp struct {
	expected *regexp.Regexp
}

func (f fieldRegexp) assert(t assert.TestingT, actual interface{}, field string, index int, level string) bool {
	a, ok := actual.(string)
	if assert.True(t, ok, fieldRegexpTypeMsg, field, index, level) {
		return assert.True(
			t,
			f.expected.MatchString(a),
			fieldRegexpMsg,
			field,
			index,
			level,
			f.expected.String(),
			a,
		)
	}

	return false
}

func (f fieldRegexp) String() string {
	return fmt.Sprintf("Regexp:%s", f.expected.String())
}

// Regexp asserts that a log field value is a string that matches the given regular expression.
//
//nolint:revive // private type is passed back into public functions so is safe to ignore
func Regexp(expected *regexp.Regexp) fieldRegexp {
	return fieldRegexp{expected: expected}
}

// fieldTrue asserts that a log field value is true.
type fieldTrue struct{}

func (f fieldTrue) assert(t assert.TestingT, actual interface{}, field string, index int, level string) bool {
	a, ok := actual.(bool)
	if assert.True(t, ok, fieldBoolTypeMsg, field, index, level) {
		return assert.True(t, a, fieldTrueMsg, field, index, level)
	}

	return false
}

func (f fieldTrue) String() string {
	return "True"
}

// True asserts that a log field value is true.
//
//nolint:revive // private type is passed back into public functions so is safe to ignore
func True() fieldTrue {
	return fieldTrue{}
}

// fieldFalse asserts that a log field value is false.
type fieldFalse struct{}

func (f fieldFalse) assert(t assert.TestingT, actual interface{}, field string, index int, level string) bool {
	a, ok := actual.(bool)
	if assert.True(t, ok, fieldBoolTypeMsg, field, index, level) {
		return assert.False(t, a, fieldFalseMsg, field, index, level)
	}

	return false
}

func (f fieldFalse) String() string {
	return "False"
}

// False asserts that a log field value is false.
//
//nolint:revive // private type is passed back into public functions so is safe to ignore
func False() fieldFalse {
	return fieldFalse{}
}

// fieldEmpty asserts that a log field value is empty.
type fieldEmpty struct{}

func (f fieldEmpty) assert(t assert.TestingT, actual interface{}, field string, index int, level string) bool {
	return assert.Empty(t, actual, fieldEmptyMsg, field, index, level)
}

func (f fieldEmpty) String() string {
	return "Empty"
}

// Empty asserts that a log field value is empty.
//
//nolint:revive // private type is passed back into public functions so is safe to ignore
func Empty() fieldEmpty {
	return fieldEmpty{}
}

// fieldNil asserts that a log field value is nil.
type fieldNil struct{}

func (f fieldNil) assert(t assert.TestingT, actual interface{}, field string, index int, level string) bool {
	return assert.Nil(t, actual, fieldNilMsg, field, index, level)
}

func (f fieldNil) String() string {
	return "Nil"
}

// Nil asserts that a log field value is nil.
//
//nolint:revive // private type is passed back into public functions so is safe to ignore
func Nil() fieldNil {
	return fieldNil{}
}
