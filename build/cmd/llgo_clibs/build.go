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
		goos := os.Getenv("GOOS")
		goarch := os.Getenv("GOARCH")
		if goos == "" {
			goos = runtime.GOOS
		}
		if goarch == "" {
			goarch = runtime.GOARCH
		}

		// Get flag values
		force, _ := cmd.Flags().GetBool("force")
		prebuilt, _ := cmd.Flags().GetBool("prebuilt")

		fmt.Printf("Build: GOOS: %s, GOARCH: %s, Force: %v, Prebuilt: %v\n",
			goos, goarch, force, prebuilt)

		pkgs, err := build.ListPkgs(args...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting C library packages: %v\n", err)
			os.Exit(1)
		}
		buildConfig := build.BuildConfig{
			Goos:     goos,
			Goarch:   goarch,
			Force:    force,
			Prebuilt: prebuilt,
		}
		err = build.Build(buildConfig, pkgs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

	// Add flags to the build command
	buildCmd.Flags().BoolP("force", "f", false, "Force rebuild even if already built")
	buildCmd.Flags().BoolP("prebuilt", "p", false, "Build to prebuilt directory")
}
