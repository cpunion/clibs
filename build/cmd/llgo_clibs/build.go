package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/goplus/clibs/build"
	"github.com/spf13/cobra"
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build [modules...]",
	Short: "Build C libraries for specified modules",
	Long:  `Build C libraries for specified Go modules based on pkg.yaml configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := build.Build(runtime.GOOS, runtime.GOARCH, args...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
