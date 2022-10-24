package models

// FTP status codes, defined in RFC 959
const (
	StatusNoCheck = -1

	StatusAlreadyOpen = 125
	StatusAboutToSend = 150

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

	StatusUserOK             = 331
	StatusRequestFilePending = 350

	StatusBadCommand              = 500
	StatusBadArguments            = 501
	StatusNotImplementedParameter = 504
	StatusFileUnavailable         = 550
)
