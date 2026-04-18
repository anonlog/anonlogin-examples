// anonlogin is the command-line interface for anonlog.in.
package main

import (
	"fmt"
	"os"

	"anonlogin-cli/cmd"
)

func main() {
	if err := cmd.Root().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
