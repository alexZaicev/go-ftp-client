package ftp_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/alexZaicev/go-ftp-client/internal/domain/entities"
)

const (
	dirPath              = "/foo/bar/baz"
	remoteDirPath        = "/doo/dee/daa"
	fileName             = "foobarbaz.txt"
	sizeInBytes   uint64 = 587

	dateFormat = "2006-01-02 15:04"
)

var (
	remotePathNoDir   = fileName
	remotePathWithDir = fmt.Sprintf("%s/%s", remoteDirPath, fileName)
)

func getEntries(t *testing.T) []*entities.Entry {
	return []*entities.Entry{
		newEntry(t, "file5", 167, "2022-01-12 16:23"),
		newEntry(t, "file3", 40032, "2022-01-24 16:23"),
		newEntry(t, "file9", 2, "2022-01-02 11:23"),
		newEntry(t, "file2", 102, "2021-05-02 17:23"),
		newEntry(t, "file6", 9635, "2022-01-02 13:23"),
		newEntry(t, "file7", 4352, "2020-04-02 14:23"),
		newEntry(t, "file1", 100, "2022-01-02 15:23"),
		newEntry(t, "file8", 1034, "2022-09-02 10:23"),
		newEntry(t, "file4", 5043, "2022-01-12 19:23"),
	}
}

func newEntry(t *testing.T, name string, sizeInBytes uint64, dateStr string) *entities.Entry {
	date, err := time.Parse(dateFormat, dateStr)
	require.NoError(t, err, "Failed to parse test case date")

	return &entities.Entry{
		Type:                 entities.EntryTypeFile,
		Permissions:          "rwxrwxrwx",
		Name:                 name,
		OwnerUser:            "user01",
		OwnerGroup:           "group01",
		SizeInBytes:          sizeInBytes,
		NumHardLinks:         2,
		LastModificationDate: date,
	}
}
