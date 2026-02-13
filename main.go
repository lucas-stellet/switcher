package main

import (
	"fmt"
	"os"

	"github.com/lucas-stellet/switcher/internal/cli"
)

var version = "dev"

func main() {
	if err := cli.Run(os.Args[1:], version); err != nil {
		fmt.Fprintf(os.Stderr, "switcher: %s\n", err)
		os.Exit(1)
	}
}
