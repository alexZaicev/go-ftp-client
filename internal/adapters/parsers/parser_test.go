package parsers_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/parsers"
	"github.com/alexZaicev/go-ftp-client/internal/domain/entities"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
)

func Test_NewGenericListParser_Success(t *testing.T) {
	assert.NotNil(t, parsers.NewGenericListParser())
}

func Test_GenericListParser_Parse_Success(t *testing.T) {
	expectedEntry := &entities.Entry{
		Type:                 entities.EntryTypeFile,
		Permissions:          "rwxr-xr-x",
		NumHardLinks:         1,
		OwnerUser:            "ftp",
		OwnerGroup:           "ftpg",
		SizeInBytes:          672,
		LastModificationDate: time.Date(0, 9, 8, 15, 15, 0, 0, time.UTC),
		Name:                 "docker-compose.yaml",
	}

	p := parsers.NewGenericListParser()
	input := "-rwxr-xr-x   1 ftp      ftpg           672 Sep 08 15:15 docker-compose.yaml"
	entry, err := p.Parse(input, &parsers.Options{})
	assert.Equal(t, expectedEntry, entry)
	assert.NoError(t, err)
}

func Test_NewGenericListParser_InvalidArgumentErrors(t *testing.T) {
	testCases := []struct {
		name           string
		input          string
		options        *parsers.Options
		expectedErrMsg string
	}{
		{
			name:           "blank input",
			options:        &parsers.Options{},
			expectedErrMsg: "an invalid argument error occurred: argument data cannot be blank",
		},
		{
			name:           "nil options",
			input:          "-rwxr-xr-x   1 ftp      ftpg           672 Sep 08 15:15 docker-compose.yaml",
			expectedErrMsg: "an invalid argument error occurred: argument options cannot be nil",
		},
	}

	for _, tc := range testCases {
		p := parsers.NewGenericListParser()
		entry, err := p.Parse(tc.input, tc.options)
		assert.Nil(t, entry)
		require.EqualError(t, err, tc.expectedErrMsg)
		assert.IsType(t, ftperrors.InvalidArgumentErrorType, err)
		assert.NoError(t, errors.Unwrap(err), nil)
	}
}

func Test_GenericListParser_Parse_Error(t *testing.T) {
	p := parsers.NewGenericListParser()
	input := "not-valid"

	entry, err := p.Parse(input, &parsers.Options{})
	assert.Nil(t, entry)
	require.EqualError(t, err, "an internal error occurred: unsupported entry format")
	assert.IsType(t, ftperrors.InternalErrorType, err)
	assert.NoError(t, errors.Unwrap(err), nil)
}
