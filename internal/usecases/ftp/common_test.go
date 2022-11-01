package ftp_test

import (
	"fmt"
	"math/rand"
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

	localPathWithDir = fmt.Sprintf("%s/%s", dirPath, fileName)

	fileContent = randStringRunes(sizeInBytes)

	letterRunes = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	rootDir1 = &entities.Entry{
		Type:                 entities.EntryTypeDir,
		Permissions:          "rwxrwxrwx",
		Name:                 ".",
		OwnerUser:            "user01",
		OwnerGroup:           "group01",
		SizeInBytes:          sizeInBytes,
		NumHardLinks:         2,
		LastModificationDate: time.Date(2022, 1, 12, 16, 23, 0, 0, time.UTC),
	}
	rootDir2 = &entities.Entry{
		Type:                 entities.EntryTypeDir,
		Permissions:          "rwxrwxrwx",
		Name:                 "..",
		OwnerUser:            "user01",
		OwnerGroup:           "group01",
		SizeInBytes:          sizeInBytes,
		NumHardLinks:         2,
		LastModificationDate: time.Date(2022, 1, 12, 16, 23, 0, 0, time.UTC),
	}
)

//nolint:gosec // math/random is used, but linter thinks it's crypto/rand
func randStringRunes(n uint64) []byte {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return b
}

func getEntries(t *testing.T) []*entities.Entry {
	return []*entities.Entry{
		newEntry(t, entities.EntryTypeFile, "file5", 167, "2022-01-12 16:23"),
		newEntry(t, entities.EntryTypeFile, "file3", 40032, "2022-01-24 16:23"),
		newEntry(t, entities.EntryTypeFile, "file9", 2, "2022-01-02 11:23"),
		newEntry(t, entities.EntryTypeFile, "file2", 102, "2021-05-02 17:23"),
		newEntry(t, entities.EntryTypeFile, "file6", 9635, "2022-01-02 13:23"),
		newEntry(t, entities.EntryTypeFile, "file7", 4352, "2020-04-02 14:23"),
		newEntry(t, entities.EntryTypeFile, "file1", 100, "2022-01-02 15:23"),
		newEntry(t, entities.EntryTypeFile, "file8", 1034, "2022-09-02 10:23"),
		newEntry(t, entities.EntryTypeFile, "file4", 5043, "2022-01-12 19:23"),
	}
}

func newEntry(
	t *testing.T,
	entryType entities.EntryType,
	name string,
	sizeInBytes uint64,
	dateStr string,
) *entities.Entry {
	date, err := time.Parse(dateFormat, dateStr)
	require.NoError(t, err, "Failed to parse test case date")

	return &entities.Entry{
		Type:                 entryType,
		Permissions:          "rwxrwxrwx",
		Name:                 name,
		OwnerUser:            "user01",
		OwnerGroup:           "group01",
		SizeInBytes:          sizeInBytes,
		NumHardLinks:         2,
		LastModificationDate: date,
	}
}
