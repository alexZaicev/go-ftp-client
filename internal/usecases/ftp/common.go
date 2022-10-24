package ftp

func isRootDir(name string) bool {
	return name == "." || name == ".."
}
