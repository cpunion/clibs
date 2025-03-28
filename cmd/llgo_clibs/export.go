package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/cpunion/clibs"
	"github.com/spf13/cobra"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export [libs...]",
	Short: "Export environment variables from C libraries",
	Long:  `Export environment variables from C libraries for specified Go libs based on lib.yaml configuration.`,
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
		prebuilt, _ := cmd.Flags().GetBool("prebuilt")

		fmt.Printf("Export: GOOS: %s, GOARCH: %s, Prebuilt: %v\n",
			goos, goarch, prebuilt)

		libs, err := clibs.ListLibs(args...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting C library libs: %v\n", err)
			os.Exit(1)
		}

		buildConfig := clibs.Config{
			Goos:     goos,
			Goarch:   goarch,
			Prebuilt: prebuilt,
		}
		exports, err := clibs.Export(buildConfig, libs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(exports) == 0 {
			fmt.Println("No exports found.")
			return
		}

		// Print the exported variables
		for _, export := range exports {
			fmt.Printf("%s\n", export)
		}
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)

	// Add flags to the export command
	exportCmd.Flags().BoolP("prebuilt", "p", false, "Export from prebuilt directory")
}
