package ftpclient_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient"
	"github.com/alexZaicev/go-ftp-client/internal/domain/entities"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
)

func Test_CallbackWriter_Write_Success(t *testing.T) {
	bytesReadActual := int64(0)

	cw := ftpclient.CallbackWriter{
		Callback: func(bytesRead int64) {
			bytesReadActual = bytesRead
		},
	}

	s := "hello world"

	n, err := cw.Write([]byte(s))
	assert.NoError(t, err)
	assert.Equal(t, int64(n), bytesReadActual)
}

func Test_FormatSizeInBytes_Success(t *testing.T) {
	testCases := []struct {
		name           string
		input          uint64
		expectedOutput string
	}{
		{
			name:           "simple B",
			input:          150,
			expectedOutput: "150 B",
		},
		{
			name:           "simple KiB",
			input:          150 * uint64(ftpclient.KiB),
			expectedOutput: "150.00 KB",
		},
		{
			name:           "simple MiB",
			input:          150 * uint64(ftpclient.MiB),
			expectedOutput: "150.00 MB",
		},
		{
			name:           "simple GiB",
			input:          150 * uint64(ftpclient.GiB),
			expectedOutput: "150.00 GB",
		},
		{
			name:           "simple TiB",
			input:          150 * uint64(ftpclient.TiB),
			expectedOutput: "150.00 TB",
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

func Test_EntryTypeToStr_Success(t *testing.T) {
	testCases := []struct {
		name      string
		entryType entities.EntryType
		expected  string
	}{
		{
			name:      "file",
			entryType: entities.EntryTypeFile,
			expected:  "F",
		},
		{
			name:      "directory",
			entryType: entities.EntryTypeDir,
			expected:  "D",
		},
		{
			name:      "link",
			entryType: entities.EntryTypeLink,
			expected:  "L",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := ftpclient.EntryTypeToStr(tc.entryType)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func Test_EntryTypeToStr_Error(t *testing.T) {
	actual, err := ftpclient.EntryTypeToStr(entities.EntryType(0))
	assert.Empty(t, actual)
	require.EqualError(t, err, "an unknown error occurred: unexpected entry type: 0")
	assert.IsType(t, ftperrors.UnknownErrorType, err)
	assert.NoError(t, errors.Unwrap(err))
}
