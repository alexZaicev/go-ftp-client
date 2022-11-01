package models

type Argument struct {
	Long  string
	Short string
	Help  string
}

var (
	ArgAddress         = Argument{Long: "address", Short: "a", Help: "Connection address for the FTP server (e.g. ftp.example.com:21)"}
	ArgUser            = Argument{Long: "user", Short: "u", Help: "Username for the FTP server user"}
	ArgPassword        = Argument{Long: "password", Short: "p", Help: "Password for the FTP server user"}
	ArgVerbose         = Argument{Long: "verbose", Short: "v", Help: "Verbose output"}
	ArgTLSCertFilePath = Argument{Long: "tls-cert", Help: "Path to TLS certificate file"}
	ArgTLSKeyFilePath  = Argument{Long: "tls-key", Help: "Path to TLS key file"}
	ArgTLSInsecure     = Argument{Long: "tls-insecure", Help: "Skip TLS certificate verification"}

	ArgRecursive = Argument{Long: "recursive", Short: "r"}
)
