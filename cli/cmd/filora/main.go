package main

import (
	"fmt"
	"os"

	"github.com/rizqynugroho9/filora-dam/cli/internal/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
