package ftpclient_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpclient"
)

func Test_NewConnector_Success(t *testing.T) {
	assert.NotNil(t, ftpclient.NewConnector())
}
