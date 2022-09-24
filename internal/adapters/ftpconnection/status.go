package ftpconnection

import (
	"strings"

	"github.com/alexZaicev/go-ftp-client/internal/domain/entities"
	ftperrors "github.com/alexZaicev/go-ftp-client/internal/domain/errors"
)

func (c *ServerConnection) Status() (*entities.Status, error) {
	_, msg, err := c.cmd(StatusSystem, CommandStatus)
	if err != nil {
		return nil, ftperrors.NewInternalError("failed to fetch server status", err)
	}

	status := &entities.Status{}

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
		// TODO: add status check TLS
	}

	_, msg, err = c.cmd(StatusName, CommandSystem)
	if err != nil {
		return nil, ftperrors.NewInternalError("failed to fetch system type", err)
	}

	msg = strings.TrimSpace(msg)
	const tokenSize = 2
	tokens := strings.SplitN(msg, " ", tokenSize)
	status.System = tokens[0]

	return status, nil
}