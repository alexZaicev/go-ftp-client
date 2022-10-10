package tests

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/alexZaicev/go-ftp-client/internal/drivers/cli"
	"github.com/alexZaicev/go-ftp-client/tests/utils"
)

func Test_FTPClient_Status(t *testing.T) {
	ts := NewStatusTestSuite(t)
	suite.Run(t, ts)
}

type StatusTestSuite struct {
	suite.Suite
	config    *utils.Config
	clientCMD *cobra.Command
}

func NewStatusTestSuite(t *testing.T) *StatusTestSuite {
	config, err := utils.LoadConfig()
	require.NoError(t, err, "error loading tests configuration file")

	clientCMD, err := cli.NewGfcCommand()
	require.NoError(t, err, "error creating client CMD")

	return &StatusTestSuite{
		config:    config,
		clientCMD: clientCMD,
	}
}

func (s *StatusTestSuite) Test_StatusTest_Happy() {
	// arrange
	outBuffer := bytes.NewBufferString("")
	errBuffer := bytes.NewBufferString("")

	s.clientCMD.SetOut(outBuffer)
	s.clientCMD.SetErr(errBuffer)
	s.clientCMD.SetArgs([]string{
		"status",
		"-a",
		s.config.Address,
		"-u",
		s.config.User,
		"-p",
		s.config.Password,
	})

	// act
	err := s.clientCMD.Execute()

	// assert
	require.NoError(s.T(), err)

	assert.Empty(s.T(), errBuffer.String())

	result := outBuffer.String()
	fmt.Println(result)

	for _, eh := range []string{
		"STATUS", "SYSTEM", "REMOTE ADDRESS", "LOGGED IN USER", "TLS ENABLED",
	} {
		assert.Equal(
			s.T(),
			1,
			strings.Count(result, eh),
			"expected header %q not found",
			eh,
		)
	}
}

// func (s *StatusTestSuite) Test_StatusTest_InvalidAddress() {
//	// arrange
//	outBuffer := bytes.NewBufferString("")
//	errBuffer := bytes.NewBufferString("")
//
//	s.clientCMD.SetOut(outBuffer)
//	s.clientCMD.SetErr(errBuffer)
//	s.clientCMD.SetArgs([]string{
//		"status",
//		"-a",
//		"not-valid-address:21",
//		"-u",
//		s.config.User,
//		"-p",
//		s.config.Password,
//	})
//
//	// act
//	err := s.clientCMD.Execute()
//
//	// assert
//	assert.NotEmpty(s.T(), outBuffer.String())
//	assert.NotEmpty(s.T(), errBuffer.String())
//
//	require.EqualError(s.T(), err, "an internal error occurred: failed to establish connection")
//	assert.IsType(s.T(), ftperrors.InternalErrorType, err)
//	assert.EqualError(s.T(), errors.Unwrap(err), "an internal error occurred: failed dial server on [not-valid-address:21] address")
// }
//
// func (s *StatusTestSuite) Test_StatusTest_InvalidLogin() {
//	testCases := []struct {
//		name           string
//		user           string
//		pass           string
//		expectedErrMsg string
//	}{
//		{
//			name:           "invalid user",
//			user:           "not-valid",
//			pass:           s.config.Password,
//			expectedErrMsg: "an internal error occurred: Login incorrect.",
//		},
//		{
//			name:           "invalid pass",
//			user:           s.config.User,
//			pass:           "not-valid",
//			expectedErrMsg: "an internal error occurred: failed to authenticate user",
//		},
//	}
//
//	for _, tc := range testCases {
//		s.T().Run(tc.name, func(t *testing.T) {
//			// arrange
//			outBuffer := bytes.NewBufferString("")
//			errBuffer := bytes.NewBufferString("")
//
//			s.clientCMD.SetOut(outBuffer)
//			s.clientCMD.SetErr(errBuffer)
//			s.clientCMD.SetArgs([]string{
//				"status",
//				"-a",
//				s.config.Address,
//				"-u",
//				tc.user,
//				"-p",
//				tc.pass,
//			})
//
//			server := mock_server.NewFtpServer(s.config)
//			require.NoError(s.T(), server.Start(), "error starting mock FTP server")
//
//			// act
//			err := s.clientCMD.Execute()
//
//			// assert
//			// stop mock server before reading error channel
//			server.Stop()
//			require.NoError(s.T(), server.Error(), "error recorded by the server")
//
//			assert.NotEmpty(s.T(), outBuffer.String())
//			assert.NotEmpty(s.T(), errBuffer.String())
//
//			require.EqualError(s.T(), err, "an internal error occurred: failed to authenticate with provided user account")
//			assert.IsType(s.T(), ftperrors.InternalErrorType, err)
//			assert.EqualError(s.T(), errors.Unwrap(err), tc.expectedErrMsg)
//		})
//	}
// }
