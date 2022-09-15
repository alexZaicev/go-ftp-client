package ftpconnection

import "io"

type verboseConnectionWrapper struct {
	io.Reader
	io.Writer
	conn io.ReadWriteCloser
}

func newVerboseConnectionWrapper(conn io.ReadWriteCloser, w io.Writer) *verboseConnectionWrapper {
	return &verboseConnectionWrapper{
		conn:   conn,
		Reader: io.TeeReader(conn, w),
		Writer: io.MultiWriter(conn, w),
	}
}

func (c *verboseConnectionWrapper) Close() error {
	return c.conn.Close()
}
