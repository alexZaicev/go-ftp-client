package connection

import (
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
	List(options *ListOptions) ([]*entities.Entry, error)
	Status() (*entities.Status, error)
	Mkdir(path string) error
	Upload(options *UploadOptions) error
	Cd(path string) error
	Size(path string) (uint64, error)
}
