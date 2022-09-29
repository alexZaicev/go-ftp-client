package assertlogging_test

import (
	"fmt"
	"strings"

	"github.com/stretchr/testify/assert"
)

const (
	// Prefixes added to test output by testify
	messagesPrefix   = "Messages:"
	errorTracePrefix = "Error Trace:"
)

// TestOutput is a mock implementation of the assert.TestingT interface to support validation of
// assert messages during unit tests.
type TestOutput struct {
	output string
}

// Non-allocating compile time check to ensure the assert.TestingT interface is implemented
// correctly.
var _ assert.TestingT = (*TestOutput)(nil)

// Errorf formats the given error messages and adds it to a buffer
func (t *TestOutput) Errorf(format string, args ...interface{}) {
	t.output += fmt.Sprintf(format, args...)
}

// Messages outputs the messages added by Errorf as a slice.
func (t *TestOutput) Messages() []string {
	var messages []string

	rawMessages := strings.Split(t.output, messagesPrefix)
	if len(rawMessages) > 0 {
		// The very first entry will be the test error trace, etc. of the first set of messages
		rawMessages = rawMessages[1:]
	}

	for _, rawMessage := range rawMessages {
		if index := strings.Index(rawMessage, errorTracePrefix); index != -1 {
			// Remove the Error Trace (from the next test error)
			rawMessage = rawMessage[:index]
		}
		// Make result one entry per message
		msgs := strings.Split(rawMessage, "\n")
		for _, msg := range msgs {
			if msg = strings.TrimSpace(msg); msg != "" {
				messages = append(messages, msg)
			}
		}
	}

	return messages
}
