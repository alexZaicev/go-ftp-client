package connection

import (
	"context"
	"io"

	"github.com/alexZaicev/go-ftp-client/internal/domain/entities"
)

type ListOptions struct {
	Path    string
	ShowAll bool
}

type UploadOptions struct {
	FileReader io.Reader
	Path       string
}

type Connection interface {
	Ready() error
	Stop() error
	Login(user, password string) error
	EnableExplicitTLSMode() error
	List(ctx context.Context, options *ListOptions) ([]*entities.Entry, error)
	Status() (*entities.Status, error)
	Mkdir(path string) error
	Upload(ctx context.Context, options *UploadOptions) error
	Cd(path string) error
	Size(path string) (uint64, error)
	RemoveFile(path string) error
	RemoveDir(path string) error
	Move(oldPath string, newPath string) error
}
