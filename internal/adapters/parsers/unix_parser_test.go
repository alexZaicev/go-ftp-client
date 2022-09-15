package parsers_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/parsers"
	"github.com/alexZaicev/go-ftp-client/internal/domain/entities"
)

func Test_UnixListParser_Parse_Success(t *testing.T) {
	testCases := []struct {
		name          string
		input         string
		expectedEntry *entities.Entry
	}{
		{
			name:  "file entry",
			input: "-rwxr-xr-x   1 ftp      ftpg           672 Sep 08 15:15 docker-compose.yaml",
			expectedEntry: &entities.Entry{
				Type:                 entities.EntryTypeFile,
				Permissions:          "rwxr-xr-x",
				NumHardLinks:         1,
				OwnerUser:            "ftp",
				OwnerGroup:           "ftpg",
				SizeInBytes:          672,
				LastModificationDate: time.Date(0, 9, 8, 15, 15, 0, 0, time.UTC),
				Name:                 "docker-compose.yaml",
			},
		},
		{
			name:  "hidden . directory",
			input: "dr--------   23 root      root           2 Sep 08 15:15 .",
			expectedEntry: &entities.Entry{
				Type:                 entities.EntryTypeDir,
				Permissions:          "r--------",
				NumHardLinks:         23,
				OwnerUser:            "root",
				OwnerGroup:           "root",
				SizeInBytes:          2,
				LastModificationDate: time.Date(0, 9, 8, 15, 15, 0, 0, time.UTC),
				Name:                 ".",
			},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			p := parsers.NewUnixListParser()
			actual, err := p.Parse(tc.input)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedEntry, actual)
		})
	}
}
