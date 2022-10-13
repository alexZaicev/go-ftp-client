package parsers_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/parsers"
	"github.com/alexZaicev/go-ftp-client/internal/domain/entities"
)

func Test_unixListParser_Parse_Success(t *testing.T) {
	// TODO: add below test cases:
	// drwxr-xr-x    3 110      1002            3 Dec 02  2009 pub
	// drwxr-xr-x    3 110      1002            3 Dec 02  2009 p u b
	// -rw-r--r--   1 marketwired marketwired    12016 Mar 16  2016 2016031611G087802-001.newsml
	// drwxr-xr-x    3 110      1002            3 Dec 02  2009 spaces   dir   name
	// -rwxr-xr-x    3 110      1002            1234567 Dec 02  2009 file   name
	// -rwxr-xr-x    3 110      1002            1234567 Dec 02  2009  foo bar
	// lrwxrwxrwx    1 0        1001           27 Jul 07  2017 R-3.4.0.pkg -> el-capitan/base/R-3.4.0.pkg

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
		{
			name:  "hidden . directory",
			input: "lr--------   23 root      root           2 Sep 08 15:15 logs -> /var/logs",
			expectedEntry: &entities.Entry{
				Type:                 entities.EntryTypeLink,
				Permissions:          "r--------",
				NumHardLinks:         23,
				OwnerUser:            "root",
				OwnerGroup:           "root",
				SizeInBytes:          2,
				LastModificationDate: time.Date(0, 9, 8, 15, 15, 0, 0, time.UTC),
				Name:                 "logs",
				LinkName:             "/var/logs",
			},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			p := parsers.NewGenericListParser()
			actual, err := p.Parse(tc.input, &parsers.Options{})
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedEntry, actual)
		})
	}
}

func Test_unixListParser_Errors(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "failed to parse number of hard links",
			input: "-rwxr-xr-x   not-valid ftp      ftpg           672 Sep 08 15:15 docker-compose.yaml",
		},
		{
			name:  "failed to parse size in bytes",
			input: "-rwxr-xr-x   1 ftp      ftpg           not-valid Sep 08 15:15 docker-compose.yaml",
		},
		{
			name:  "failed to parse last modification date",
			input: "-rwxr-xr-x   1 ftp      ftpg           672 not-valid docker-compose.yaml",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := parsers.NewGenericListParser()
			actual, err := p.Parse(tc.input, &parsers.Options{
				Location: time.UTC,
			})
			assert.Nil(t, actual)
			assert.EqualError(t, err, "an internal error occurred: unsupported entry format")
		})
	}
}
