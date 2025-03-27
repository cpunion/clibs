package main

import (
	"fmt"
	"os"

	"github.com/cpunion/clibs/build"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list [modules...]",
	Short: "List C libraries in specified modules",
	Long:  `List C libraries in specified Go modules based on pkg.yaml configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get packages from specified modules or all modules if none specified
		pkgs, err := build.ListPkgs(args...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing C library packages: %v\n", err)
			os.Exit(1)
		}

		// Display the results
		if len(pkgs) == 0 {
			fmt.Println("No C library packages found.")
			return
		}

		fmt.Printf("Found %d C library package(s):\n", len(pkgs))
		for _, pkg := range pkgs {
			fmt.Printf("- %s\n", pkg.Mod)
			fmt.Printf("  Path: %s\n", pkg.Path)
			fmt.Printf("  Sum: %s\n", pkg.Sum)
			fmt.Printf("  Version: %s\n", pkg.Config.Version)
			if pkg.Config.Git != nil {
				fmt.Printf("  Git: %s@%s\n", pkg.Config.Git.Repo, pkg.Config.Git.Ref)
			}
			fmt.Println()
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
