package models

// StatusNoCheck ignores status code validation upon command execution.
const StatusNoCheck = -1

// Positive Preliminary reply.
//
// The requested action is being initiated; expect another reply before proceeding with a new command.
// (The user-process sending another command before the completion reply would be in violation of protocol;
// but server-FTP processes should queue any commands that arrive while a preceding command is in progress.)
// This type of reply can be used to indicate that the command was accepted and the user-process may now pay
// attention to the data connections, for implementations where simultaneous monitoring is difficult. The
// server-FTP process may send at most, one 1xx reply per command.
const (
	StatusAlreadyOpen = 125
	StatusAboutToSend = 150
)

// Positive Completion reply
//
// The requested action has been successfully completed. A new request may be initiated.
const (
	StatusCommandOK             = 200
	StatusCommandNotImplemented = 202
	StatusSystem                = 211
	StatusFile                  = 213
	StatusName                  = 215
	StatusReady                 = 220
	StatusClosingDataConnection = 226
	StatusPassiveMode           = 227
	StatusExtendedPassiveMode   = 229
	StatusLoggedIn              = 230
	StatusAuthOK                = 234
	StatusRequestedFileActionOK = 250
	StatusPathCreated           = 257
)

// Positive Intermediate reply
//
// The command has been accepted, but the requested action is being held in abeyance, pending receipt of
// further information. The user should send another command specifying this information. This reply is
// used in command sequence groups.
const (
	StatusUserOK             = 331
	StatusRequestFilePending = 350
)

// Permanent Negative Completion reply
//
// The command was not accepted and the requested action did not take place. The User-process is discouraged
// from repeating the exact request (in the same sequence). Even some "permanent" error conditions can be
// corrected, so the human user may want to direct his User-process to re-initiate the command sequence by
// direct action at some point in the future (e.g., after the spelling has been changed, or the user has
// altered his directory status.)
const (
	StatusBadCommand              = 500
	StatusBadArguments            = 501
	StatusNotImplementedParameter = 504
	StatusFileUnavailable         = 550
)
