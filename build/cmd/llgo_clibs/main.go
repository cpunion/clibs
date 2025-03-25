package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	// Set program name to match the binary name
	_, progName := filepath.Split(os.Args[0])
	rootCmd.Use = progName

	if err := Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}
