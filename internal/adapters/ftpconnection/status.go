package ftpconnection

import (
	"strings"

	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpconnection/models"
	"github.com/alexZaicev/go-ftp-client/internal/domain/entities"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
)

//nolint:gocyclo // parsing status lines can be quite long
func (c *ServerConnection) Status() (*entities.Status, error) {
	_, msg, err := c.cmd(models.StatusSystem, models.CommandStatus)
	if err != nil {
		return nil, ftperrors.NewInternalError("failed to fetch server status", err)
	}

	status := &entities.Status{}

	var connEncrypted bool
	var dataConnEncrypted bool

	lines := strings.Split(msg, "\n")
	for _, line := range lines {
		line = strings.ToLower(line)
		if strings.Contains(line, "server status:") || strings.Contains(line, "end of status") {
			continue
		}
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "connected to") {
			const tokenSize = 3
			tokens := strings.SplitN(line, " ", tokenSize)
			if len(tokens) >= tokenSize {
				status.RemoteAddress = tokens[2]
			}
			continue
		}

		if strings.Contains(line, "logged in") {
			const tokenSize = 4
			tokens := strings.SplitN(line, " ", tokenSize)
			if len(tokens) >= tokenSize {
				status.LoggedInUser = tokens[3]
			}
			continue
		}

		if strings.Contains(line, "control connection") && strings.Contains(line, "encrypted") {
			connEncrypted = true
			continue
		}

		if strings.Contains(line, "data connection") && strings.Contains(line, "encrypted") {
			dataConnEncrypted = true
			continue
		}
	}

	if connEncrypted && dataConnEncrypted {
		status.TLSEnabled = c.features.AuthTLS
	}

	_, msg, err = c.cmd(models.StatusName, models.CommandSystem)
	if err != nil {
		return nil, ftperrors.NewInternalError("failed to fetch system type", err)
	}

	msg = strings.TrimSpace(msg)
	const tokenSize = 2
	tokens := strings.SplitN(msg, " ", tokenSize)
	status.System = tokens[0]

	return status, nil
}
