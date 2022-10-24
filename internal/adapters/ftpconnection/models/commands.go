package models

type Command string

const (
	CommandQuit                 = "QUIT"
	CommandAuthTLS              = "AUTH TLS"
	CommandUser                 = "USER %s"
	CommandPass                 = "PASS %s"
	CommandFeat                 = "FEAT"
	CommandProtectionBufferSize = "PBSZ 0"
	CommandProtocol             = "PROT P"
	CommandType                 = "TYPE %s"
	CommandOptions              = "OPTS %s %s"
	CommandStatus               = "STAT"
	CommandSystem               = "SYST"
	CommandList                 = "LIST %s"
	CommandListHidden           = "LIST -a %s"
	CommandPreTransfer          = "PRET %s"
	CommandPassive              = "PASV"
	CommandExtendedPassiveMode  = "EPSV"
	CommandRestartTransfer      = "REST %d"
	CommandListMachineReadable  = "MLSD %s"
	CommandStore                = "STOR %s"
	CommandMakeDir              = "MKD %s"
	CommandChangeWorkDir        = "CWD %s"
	CommandSize                 = "SIZE %s"
	CommandRemoveFile           = "DELE %s"
	CommandRemoveDir            = "RMD %s"
	CommandRenameFrom           = "RNFR %s"
	CommandRenameTo             = "RNTO %s"
	CommandRetrieve             = "RETR %s"
)
