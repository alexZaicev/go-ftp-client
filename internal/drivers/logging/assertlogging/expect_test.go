package assertlogging_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/alexZaicev/go-ftp-client/internal/drivers/logging/assertlogging"
)

func Test_Logger_Expect_Happy(t *testing.T) {
	// arrange
	logger := assertlogging.NewLogger(t, assertlogging.SkipAssert())

	logger.ExpectError("example").WithField("key", assertlogging.Equal("value"))
	logger.ExpectError("example 2").WithError(assertlogging.Equal(errors.New("value")))
	logger.
		ExpectError("example 3").
		WithField("key", assertlogging.Equal("value")).
		WithField("key 2", assertlogging.Equal("value 2"))

	logger.ExpectWarn("example").WithField("key", assertlogging.Equal("value"))
	logger.ExpectInfo("example").WithField("key", assertlogging.Equal("value"))

	logger.WithField("key", "value").Error("example")
	logger.WithError(errors.New("value")).Error("example 2")
	logger.WithField("key", "value").WithField("key 2", "value 2").Error("example 3")
	logger.WithField("key", "value").Warn("example")
	logger.WithField("key", "value").Info("example")
	logger.WithField("key", "value").Debug("example")

	// act
	result := logger.AssertExpectations(t)

	// assert
	assert.True(t, result)
}

func Test_Logger_Expect_Happy_AssertWithoutOrder(t *testing.T) {
	// arrange
	logger := assertlogging.NewLogger(t, assertlogging.SkipAssert(), assertlogging.AssertWithoutOrder())

	logger.ExpectError("example 2").WithError(assertlogging.Equal(errors.New("value")))
	logger.
		ExpectError("example 3").
		WithField("key", assertlogging.Equal("value")).
		WithField("key 2", assertlogging.Equal("value 2"))
	logger.ExpectError("example").WithField("key", assertlogging.Equal("value"))

	logger.ExpectWarn("example").WithField("key", assertlogging.Equal("value"))
	logger.ExpectInfo("example").WithField("key", assertlogging.Equal("value"))

	logger.WithField("key", "value").Error("example")
	logger.WithError(errors.New("value")).Error("example 2")
	logger.WithField("key", "value").WithField("key 2", "value 2").Error("example 3")
	logger.WithField("key", "value").Warn("example")
	logger.WithField("key", "value").Info("example")
	logger.WithField("key", "value").Debug("example")

	// act
	result := logger.AssertExpectations(t)

	// assert
	assert.True(t, result)
}

func Test_Logger_Expect_Happy_AssertDebugLogs(t *testing.T) {
	// arrange
	logger := assertlogging.NewLogger(
		t,
		assertlogging.SkipAssert(),
		assertlogging.AssertDebugLogs(),
	)
	logger.ExpectDebug("example").WithField("key", assertlogging.Equal("value"))
	logger.WithField("key", "value").Debug("example")

	// act
	result := logger.AssertExpectations(t)

	// assert
	assert.True(t, result)
}

func Test_Logger_Expect_Unexpected_AssertWithoutOrder(t *testing.T) {
	// arrange
	logger := assertlogging.NewLogger(t, assertlogging.SkipAssert(), assertlogging.AssertWithoutOrder())
	logger.Error("example")

	mockT := new(TestOutput)

	// act
	result := logger.AssertExpectations(mockT)

	// assert
	assert.False(t, result)
	assert.Equal(t, []string{"no expectation matched log at index 0 of level error: message=\"example\", fields=map[]"}, mockT.Messages())
}

func Test_Logger_Expect_Unfulfilled_AssertWithoutOrder(t *testing.T) {
	// arrange
	logger := assertlogging.NewLogger(t, assertlogging.SkipAssert(), assertlogging.AssertWithoutOrder())
	logger.ExpectError("example")

	mockT := new(TestOutput)

	// act
	result := logger.AssertExpectations(mockT)

	// assert
	assert.False(t, result)
	assert.Equal(t, []string{"no log matched expectation at index 0 of level error: message=\"example\", fields=map[]"}, mockT.Messages())
}

func Test_Logger_Expect_Unmatched_AssertWithoutOrder(t *testing.T) {
	// arrange
	logger := assertlogging.NewLogger(t, assertlogging.SkipAssert(), assertlogging.AssertWithoutOrder())
	logger.ExpectError("example")
	logger.ExpectError("example2")
	logger.ExpectError("example3 unfulfilled").WithField("foo", assertlogging.Equal("bar")).WithField("baz", assertlogging.True())
	logger.Error("example2")
	logger.Error("example3 unmatched")
	logger.Error("example")

	mockT := new(TestOutput)

	// act
	result := logger.AssertExpectations(mockT)

	// assert
	assert.False(t, result)
	assert.Equal(
		t,
		[]string{
			"no expectation matched log at index 1 of level error: message=\"example3 unmatched\", fields=map[]",
			"no log matched expectation at index 2 of level error: message=\"example3 unfulfilled\", fields=map[baz:[True] foo:[Equal:bar]]",
		},
		mockT.Messages(),
	)
}

func Test_Logger_Expect_Unexpected_AssertDebugLogs(t *testing.T) {
	// arrange
	logger := assertlogging.NewLogger(t, assertlogging.SkipAssert())
	logger.ExpectDebug("example").WithField("key", assertlogging.Equal("value"))

	mockT := new(TestOutput)

	// act
	result := logger.AssertExpectations(mockT)

	// assert
	assert.False(t, result)
	assert.Equal(t, []string{"unexpected expectations on debug logs present"}, mockT.Messages())
}

func Test_Logger_Expect_Unexpected(t *testing.T) {
	// arrange
	logger := assertlogging.NewLogger(t, assertlogging.SkipAssert())
	logger.Error("example")

	mockT := new(TestOutput)

	// act
	result := logger.AssertExpectations(mockT)

	// assert
	assert.False(t, result)
	assert.Equal(t, []string{"unexpected error log at index 0"}, mockT.Messages())
}

func Test_Logger_Expect_Unfulfilled(t *testing.T) {
	// arrange
	logger := assertlogging.NewLogger(t, assertlogging.SkipAssert())
	logger.ExpectError("example")

	mockT := new(TestOutput)

	// act
	result := logger.AssertExpectations(mockT)

	// assert
	assert.False(t, result)
	assert.Equal(t, []string{"missing error log at index 0"}, mockT.Messages())
}

func Test_Logger_Expect_MismatchedMessage(t *testing.T) {
	// arrange
	logger := assertlogging.NewLogger(t, assertlogging.SkipAssert())
	logger.Error("something")
	logger.ExpectError("other")

	mockT := new(TestOutput)

	// act
	result := logger.AssertExpectations(mockT)

	// assert
	assert.False(t, result)
	assert.Equal(
		t,
		[]string{`message of log at index 0 of level error does not match`},
		mockT.Messages(),
	)
}

func Test_Logger_Expect_UnexpectedField(t *testing.T) {
	// arrange
	logger := assertlogging.NewLogger(t, assertlogging.SkipAssert())
	logger.WithField("key", "value").Error("example")
	logger.ExpectError("example")

	mockT := new(TestOutput)

	// act
	result := logger.AssertExpectations(mockT)

	// assert
	assert.False(t, result)
	assert.Equal(
		t,
		[]string{`unexpected field "key" found in log at index 0 of level error`},
		mockT.Messages(),
	)
}

func Test_Logger_Expect_UnfulfilledField(t *testing.T) {
	// arrange
	logger := assertlogging.NewLogger(t, assertlogging.SkipAssert())
	logger.Error("example")
	logger.ExpectError("example").WithField("key", assertlogging.Equal("value"))

	mockT := new(TestOutput)

	// act
	result := logger.AssertExpectations(mockT)

	// assert
	assert.False(t, result)
	assert.Equal(
		t,
		[]string{`missing field "key" found in log at index 0 of level error`},
		mockT.Messages(),
	)
}

func Test_Logger_Expect_NoFieldExpectations(t *testing.T) {
	// arrange
	logger := assertlogging.NewLogger(t, assertlogging.SkipAssert())
	logger.WithField("key", "value").Error("example")
	logger.ExpectError("example").WithField("key")

	mockT := new(TestOutput)

	// act
	result := logger.AssertExpectations(mockT)

	// assert
	assert.False(t, result)
	assert.Equal(
		t,
		[]string{`no expectations found for field "key" of log at index 0 of level error`},
		mockT.Messages(),
	)
}
