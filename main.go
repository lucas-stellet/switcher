package main

import (
	"fmt"
	"os"

	"github.com/lucas/switch/internal/cli"
)

func main() {
	if err := cli.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "switch: %s\n", err)
		os.Exit(1)
	}
}
