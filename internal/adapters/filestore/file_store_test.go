package filestore_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/filestore"
)

const (
	tmpDir = "./tmp"
)

var (
	content = []byte("File is file content")
)

func Test_FileStore_SaveFile_Success(t *testing.T) {
	// arrange
	store := filestore.FileStore{}

	// act
	err := store.SaveFile(filepath.Join(tmpDir, "file-1"), content)

	// assert
	assert.NoError(t, err)
}

func Test_FileStore_CreateDir_Success(t *testing.T) {
	// arrange
	store := filestore.FileStore{}

	// act
	err := store.CreateDir(filepath.Join(tmpDir, "tmp1"))

	// assert
	assert.NoError(t, err)
}
