package tests

import (
	"bytes"
	"errors"
	"testing"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
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
	expectedOutBuffer := bytes.NewBufferString("")
	table := tablewriter.NewWriter(expectedOutBuffer)
	table.SetHeader([]string{
		"status",
		"system",
		"remote address",
		"logged in user",
		"tls enabled",
	})

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

	// extract IP
	result := outBuffer.String()
	addr := utils.ExtractRemoteAddress(result)
	table.Append([]string{"OK", "UNIX", addr, s.config.User, "NO"})
	table.Render()

	assert.Equal(s.T(), expectedOutBuffer.String(), result)
}

// nolint:dupl // similar to Test_ListTestSuite_InvalidParameters
func (s *StatusTestSuite) Test_StatusTestSuite_InvalidParameters() {
	testCases := []struct {
		name                    string
		address                 string
		user                    string
		pass                    string
		expectedErrMsg          string
		expectedUnwrappedErrMsg string
	}{
		{
			name:                    "invalid address",
			address:                 "not-valid",
			user:                    s.config.User,
			pass:                    s.config.Password,
			expectedErrMsg:          "an internal error occurred: failed to establish connection",
			expectedUnwrappedErrMsg: "an internal error occurred: failed dial server on [not-valid] address",
		},
		{
			name:                    "invalid user",
			address:                 s.config.Address,
			user:                    "not-valid",
			pass:                    s.config.Password,
			expectedErrMsg:          "an internal error occurred: failed to authenticate with provided user account",
			expectedUnwrappedErrMsg: "an internal error occurred: failed to authenticate user",
		},
		{
			name:                    "invalid pass",
			address:                 s.config.Address,
			user:                    s.config.User,
			pass:                    "not-valid",
			expectedErrMsg:          "an internal error occurred: failed to authenticate with provided user account",
			expectedUnwrappedErrMsg: "an internal error occurred: failed to authenticate user",
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			// arrange
			outBuffer := bytes.NewBufferString("")
			errBuffer := bytes.NewBufferString("")

			s.clientCMD.SetOut(outBuffer)
			s.clientCMD.SetErr(errBuffer)
			s.clientCMD.SetArgs([]string{
				"status",
				"-a",
				tc.address,
				"-u",
				tc.user,
				"-p",
				tc.pass,
			})

			// act
			err := s.clientCMD.Execute()

			// assert
			assert.NotEmpty(s.T(), outBuffer.String())
			assert.NotEmpty(s.T(), errBuffer.String())

			require.EqualError(s.T(), err, tc.expectedErrMsg)
			assert.IsType(s.T(), ftperrors.InternalErrorType, err)
			assert.EqualError(s.T(), errors.Unwrap(err), tc.expectedUnwrappedErrMsg)
		})
	}
}
