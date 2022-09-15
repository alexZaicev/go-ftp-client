package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/alexZaicev/go-ftp-client/internal/drivers/cli"
)

func main() {
	rootCMD, err := cli.NewGfcCommand()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	cobra.CheckErr(rootCMD.Execute())
}
