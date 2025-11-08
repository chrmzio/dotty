/*
Copyright © 2025 dotty <chrmzio@pm.me>
*/
package main

import (
	"fmt"
	"os"

	"github.com/chrmzio/dotty/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
