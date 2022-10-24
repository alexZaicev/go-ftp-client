package repositories

type FileStore interface {
	SaveFile(path string, data []byte) error
	CreateDir(path string) error
}
