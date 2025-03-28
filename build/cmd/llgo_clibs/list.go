package main

import (
	"fmt"
	"os"

	"github.com/cpunion/clibs/build"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list [libs...]",
	Short: "List C libraries in specified libs",
	Long:  `List C libraries in specified Go libs based on lib.yaml configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get libs from specified libs or all libs if none specified
		libs, err := build.ListLibs(args...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing C libraries: %v\n", err)
			os.Exit(1)
		}

		// Display the results
		if len(libs) == 0 {
			fmt.Println("No C libraries found.")
			return
		}

		fmt.Printf("Found %d C libraries:\n", len(libs))
		for _, lib := range libs {
			fmt.Printf("- %s\n", lib.ModName)
			fmt.Printf("  Path: %s\n", lib.Path)
			fmt.Printf("  Sum: %s\n", lib.Sum)
			fmt.Printf("  Version: %s\n", lib.Config.Version)
			if lib.Config.Git != nil {
				fmt.Printf("  Git: %s@%s\n", lib.Config.Git.Repo, lib.Config.Git.Ref)
			}
			fmt.Println()
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
