package assertlogging_test

import (
	"errors"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/alexZaicev/go-ftp-client/internal/drivers/logging/assertlogging"
)

func Test_Field_Success(t *testing.T) {
	testcases := []struct {
		name   string
		field  assertlogging.FieldValue
		actual interface{}
	}{
		{
			name:   "equal",
			field:  assertlogging.Equal("example"),
			actual: "example",
		},
		{
			name:   "equalf",
			field:  assertlogging.Equalf("%s", "example"),
			actual: "example",
		},
		{
			name:   "not equal same type",
			field:  assertlogging.NotEqual("other"),
			actual: "example",
		},
		{
			name:   "not equal different type",
			field:  assertlogging.NotEqual(100),
			actual: "example",
		},
		{
			name:   "equalerror",
			field:  assertlogging.EqualError("example"),
			actual: errors.New("example"),
		},
		{
			name:   "equalerrorf",
			field:  assertlogging.EqualErrorf("%s", "example"),
			actual: errors.New("example"),
		},
		{
			name:   "istype",
			field:  assertlogging.IsType(""),
			actual: "example",
		},
		{
			name:   "regexp",
			field:  assertlogging.Regexp(regexp.MustCompile("^example$")),
			actual: "example",
		},
		{
			name:   "true",
			field:  assertlogging.True(),
			actual: true,
		},
		{
			name:   "false",
			field:  assertlogging.False(),
			actual: false,
		},
		{
			name:   "empty (string)",
			field:  assertlogging.Empty(),
			actual: "",
		},
		{
			name:   "empty (slice)",
			field:  assertlogging.Empty(),
			actual: []string{},
		},
		{
			name:   "nil",
			field:  assertlogging.Nil(),
			actual: nil,
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// no arrange necessary

			// act
			result := assertlogging.AssertFieldValue(tc.field, t, tc.actual, "field", 0, "info")

			// assert
			assert.True(t, result)
		})
	}
}

//nolint:funlen // lots of test cases
func Test_Field_Error(t *testing.T) {
	testcases := []struct {
		name     string
		field    assertlogging.FieldValue
		actual   interface{}
		messages []string
	}{
		{
			name:     "equal same type",
			field:    assertlogging.Equal("other"),
			actual:   "example",
			messages: []string{`field "field" of log at index 0 of level info does not match`},
		},
		{
			name:     "equal different type",
			field:    assertlogging.Equal(100),
			actual:   "example",
			messages: []string{`field "field" of log at index 0 of level info does not match`},
		},
		{
			name:     "equalf",
			field:    assertlogging.Equalf("%s", "other"),
			actual:   "example",
			messages: []string{`field "field" of log at index 0 of level info does not match`},
		},
		{
			name:     "not equal",
			field:    assertlogging.NotEqual("example"),
			actual:   "example",
			messages: []string{`field "field" of log at index 0 of level info matches`},
		},
		{
			name:     "equalerror nil",
			field:    assertlogging.EqualError("example"),
			actual:   nil,
			messages: []string{`field "field" of log at index 0 of level info is not an error`},
		},
		{
			name:     "equalerror non error",
			field:    assertlogging.EqualError("example"),
			actual:   "example",
			messages: []string{`field "field" of log at index 0 of level info is not an error`},
		},
		{
			name:     "equalerror different error",
			field:    assertlogging.EqualError("other"),
			actual:   errors.New("example"),
			messages: []string{`field "field" of log at index 0 of level info does not match the expected error message`},
		},
		{
			name:     "equalerrorf",
			field:    assertlogging.EqualErrorf("%s", "other"),
			actual:   errors.New("example"),
			messages: []string{`field "field" of log at index 0 of level info does not match the expected error message`},
		},
		{
			name:     "istype nil",
			field:    assertlogging.IsType(""),
			actual:   nil,
			messages: []string{`field "field" of log at index 0 of level info is not the expected type`},
		},
		{
			name:     "istype different type",
			field:    assertlogging.IsType(""),
			actual:   100,
			messages: []string{`field "field" of log at index 0 of level info is not the expected type`},
		},
		{
			name:     "regexp not a string",
			field:    assertlogging.Regexp(regexp.MustCompile("^example$")),
			actual:   100,
			messages: []string{`field "field" of log at index 0 of level info is not a string`},
		},
		{
			name:     "regexp non-matching string",
			field:    assertlogging.Regexp(regexp.MustCompile("^other$")),
			actual:   "example",
			messages: []string{`field "field" of log at index 0 of level info does not match the regexp "^other$": example`},
		},
		{
			name:     "true same type",
			field:    assertlogging.True(),
			actual:   false,
			messages: []string{`field "field" of log at index 0 of level info is not true`},
		},
		{
			name:     "true different type",
			field:    assertlogging.True(),
			actual:   100,
			messages: []string{`field "field" of log at index 0 of level info is not a bool`},
		},
		{
			name:     "false same type",
			field:    assertlogging.False(),
			actual:   true,
			messages: []string{`field "field" of log at index 0 of level info is not false`},
		},
		{
			name:     "false different type",
			field:    assertlogging.False(),
			actual:   100,
			messages: []string{`field "field" of log at index 0 of level info is not a bool`},
		},
		{
			name:     "empty (string)",
			field:    assertlogging.Empty(),
			actual:   "hello",
			messages: []string{`field "field" of log at index 0 of level info is not empty`},
		},
		{
			name:     "empty (slice)",
			field:    assertlogging.Empty(),
			actual:   []string{"hello"},
			messages: []string{`field "field" of log at index 0 of level info is not empty`},
		},
		{
			name:     "nil",
			field:    assertlogging.Nil(),
			actual:   100,
			messages: []string{`field "field" of log at index 0 of level info is not nil`},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			mockT := new(TestOutput)

			// act
			result := assertlogging.AssertFieldValue(tc.field, mockT, tc.actual, "field", 0, "info")

			// assert
			assert.False(t, result)
			assert.Equal(t, tc.messages, mockT.Messages())
		})
	}
}

func Test_Field_String(t *testing.T) {
	testcases := []struct {
		name           string
		field          assertlogging.FieldValue
		expectedString string
	}{
		{
			name:           "equal",
			field:          assertlogging.Equal("example"),
			expectedString: "Equal:example",
		},
		{
			name:           "not equal",
			field:          assertlogging.NotEqual("other"),
			expectedString: "NotEqual:other",
		},
		{
			name:           "equalerror",
			field:          assertlogging.EqualError("example"),
			expectedString: "EqualError:example",
		},
		{
			name:           "istype",
			field:          assertlogging.IsType(""),
			expectedString: "IsType:string",
		},
		{
			name:           "regexp",
			field:          assertlogging.Regexp(regexp.MustCompile("^example$")),
			expectedString: "Regexp:^example$",
		},
		{
			name:           "true",
			field:          assertlogging.True(),
			expectedString: "True",
		},
		{
			name:           "false",
			field:          assertlogging.False(),
			expectedString: "False",
		},
		{
			name:           "empty",
			field:          assertlogging.Empty(),
			expectedString: "Empty",
		},
		{
			name:           "nil",
			field:          assertlogging.Nil(),
			expectedString: "Nil",
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// no arrange necessary

			// act
			result := tc.field.String()

			// assert
			assert.Equal(t, tc.expectedString, result)
		})
	}
}
