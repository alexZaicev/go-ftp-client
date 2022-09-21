package assertlogging

// Option implementations can be used to configure a new Logger instance.
type Option = func(*loggerOptions)

// loggerOptions is a consolidation of the possible Logger configuration options.
type loggerOptions struct {
	assertDebugLogs    bool
	assertWithoutOrder bool
	skipAssert         bool
}

// AssertDebugLogs enables assertions against DEBUG level logs. If this is not used, DEBUG level
// logs are ignored by Logger.AssertExpectations.
func AssertDebugLogs() Option {
	return func(opts *loggerOptions) {
		opts.assertDebugLogs = true
	}
}

// AssertWithoutOrder enables assertions in any order for each level of logs. This allows assertions on logs
// pushed to the same underlying logger from multiple goroutines (as the order the logs are generated in may
// be nondeterministic).
//
// Each generated log will be compared with each expectation until one matches, at which point that expectation
// is removed from the available list, as such the expectations must be specific enough to match only the single
// log they are intended to match.
func AssertWithoutOrder() Option {
	return func(opts *loggerOptions) {
		opts.assertWithoutOrder = true
	}
}

// SkipAssert disables all assertions such that any or no logs can be produced without being
// checked.
//
// This is generally inappropriate for unit testing, but may be useful for contract testing where
// only the API inputs and outputs are of interest.
func SkipAssert() Option {
	return func(opts *loggerOptions) {
		opts.skipAssert = true
	}
}
