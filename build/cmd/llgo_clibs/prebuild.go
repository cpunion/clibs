package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/goplus/clibs/build"
	"github.com/spf13/cobra"
)

// prebuildCmd represents the prebuild command
var prebuildCmd = &cobra.Command{
	Use:   "prebuild [modules...]",
	Short: "Prebuild C libraries for specified modules",
	Long:  `Prebuild C libraries for specified Go modules based on pkg.yaml configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := build.Prebuild(runtime.GOOS, runtime.GOARCH, args...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(prebuildCmd)
}
