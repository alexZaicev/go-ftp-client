package ftpconnection

import (
	"errors"
	"strings"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpconnection/models"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
)

// Login function authenticate user with provided account username and password. Upon successful authentication,
// server is then queried to list supported features to update connection settings at runtime.
func (c *ServerConnection) Login(user, password string) error {
	code, msg, err := c.cmd(models.StatusNoCheck, models.CommandUser, user)
	if err != nil {
		return ftperrors.NewInternalError("failed to start username authentication", err)
	}

	switch code {
	case models.StatusLoggedIn:
	case models.StatusUserOK:
		if _, _, pwdErr := c.cmd(models.StatusLoggedIn, models.CommandPass, password); pwdErr != nil {
			return ftperrors.NewInternalError("failed to authenticate user", pwdErr)
		}
	default:
		return ftperrors.NewInternalError(msg, nil)
	}

	if updateErr := c.updateFeatures(); updateErr != nil {
		return updateErr
	}

	return nil
}

// updateFeatures function queries FTP server for supported features and adjusts connection settings
// based on user and received settings.
func (c *ServerConnection) updateFeatures() error {
	code, msg, err := c.cmd(models.StatusNoCheck, models.CommandFeat)
	if err != nil {
		return ftperrors.NewInternalError("failed to list supported features", err)
	}

	if code != models.StatusSystem || msg == "" {
		// The server does not support the FEAT command. This is not an
		// error, as we consider that there is no additional features.
		return nil
	}

	features := c.getFeaturesMap(msg)
	c.features = models.NewServerFeatures(features)

	// switch to binary mode
	if _, _, cmdErr := c.cmd(models.StatusCommandOK, models.CommandType, models.TransferTypeBinary); cmdErr != nil {
		return ftperrors.NewInternalError("failed to set binary transfer mode", cmdErr)
	}

	if c.features.SupportUTF8 && !c.disableUTF8 {
		if utfErr := c.setUTF8(); utfErr != nil {
			return ftperrors.NewInternalError("failed to turn UTF-8 option on", utfErr)
		}
	}

	// If using implicit TLS, make data connections also use TLS
	if c.tlsConfig != nil {
		if _, _, err = c.cmd(models.StatusCommandOK, models.CommandProtectionBufferSize); err != nil {
			return ftperrors.NewInternalError("failed to set protocol buffer size", err)
		}
		if _, _, err = c.cmd(models.StatusCommandOK, models.CommandProtocol); err != nil {
			return ftperrors.NewInternalError("failed to enable TLS protocol", err)
		}
	}

	return nil
}

// getFeaturesMap function processes value parameter returned by the FEAT command and
// composes a map[COMMAND]COMMAND_DESC of supporter server features.
func (c *ServerConnection) getFeaturesMap(value string) map[string]string {
	features := make(map[string]string, 0)
	for _, line := range strings.Split(value, "\n") {
		loweredLine := strings.ToLower(line)
		if strings.Contains(loweredLine, "features") || strings.Contains(loweredLine, "end") {
			continue
		}
		line = strings.TrimSpace(line)
		const tokenSize = 2
		tokens := strings.SplitN(line, " ", tokenSize)

		var cmdDesc string
		if len(tokens) == tokenSize {
			cmdDesc = tokens[1]
		}
		features[tokens[0]] = cmdDesc
	}
	return features
}

// setUTF8 function sets UTF-8 format on connected server. If server does not support this option,
// it's ignored.
func (c *ServerConnection) setUTF8() error {
	code, msg, err := c.cmd(models.StatusNoCheck, models.CommandOptions, models.FeatureUTF8, "ON")
	if err != nil {
		return err
	}

	// Workaround for FTP servers, that does not support this option.
	if code == models.StatusBadArguments || code == models.StatusNotImplementedParameter {
		return nil
	}

	// The ftpd "filezilla-server" has FEAT support for UTF8, but always returns
	// "202 UTF8 mode is always enabled. No need to send this command." when
	// trying to use it. That's OK
	if code == models.StatusCommandNotImplemented {
		return nil
	}

	if code != models.StatusCommandOK {
		return errors.New(msg)
	}
	return nil
}
