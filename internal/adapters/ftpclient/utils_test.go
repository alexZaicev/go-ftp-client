package ftpclient_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient"
)

func Test_FormatSizeInBytes_Success(t *testing.T) {
	testCases := []struct {
		name           string
		input          uint64
		expectedOutput string
	}{
		{
			name:           "simple KiB",
			input:          1504,
			expectedOutput: "1.47 KB",
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			actual := ftpclient.FormatSizeInBytes(tc.input)
			assert.Equal(t, tc.expectedOutput, actual)
		})
	}
}
