package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	baseBranch string
	verbose    bool
)

// changesCmd represents the changes command
var changesCmd = &cobra.Command{
	Use:   "changes",
	Short: "Detect changed packages that need rebuilding",
	Long:  `Detect which packages have changes in their pkg.yaml files or other files in their directories.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get the repository root directory
		repoRoot, err := getRepoRoot()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Set default base branch if not specified
		fromRef := baseBranch
		if fromRef == "" {
			// Default to HEAD~1 if no base branch specified
			fromRef = "HEAD~1"
		}
		toRef := "HEAD"

		// Path to the detection script
		scriptPath := filepath.Join(repoRoot, ".github", "scripts", "detect-changes.sh")

		// Check if the script exists
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: Detection script not found at %s\n", scriptPath)
			fmt.Fprintf(os.Stderr, "Please ensure the repository is properly set up with CI scripts\n")
			os.Exit(1)
		}

		// Make sure the script is executable
		if err := os.Chmod(scriptPath, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not make script executable: %v\n", err)
		}

		// Create a pipe to capture the script output
		cmd2 := exec.Command(scriptPath, fromRef, toRef)
		cmd2.Dir = repoRoot

		// If verbose, show all output
		if verbose {
			cmd2.Stdout = os.Stdout
			cmd2.Stderr = os.Stderr

			if err := cmd2.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Error running detection script: %v\n", err)
				os.Exit(1)
			}
		} else {
			// Otherwise, capture and parse the output
			stdout, err := cmd2.StdoutPipe()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating stdout pipe: %v\n", err)
				os.Exit(1)
			}

			cmd2.Stderr = os.Stderr

			if err := cmd2.Start(); err != nil {
				fmt.Fprintf(os.Stderr, "Error starting detection script: %v\n", err)
				os.Exit(1)
			}

			// Parse the output to get changed directories
			var changedDirs []string
			hasChanges := false

			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				line := scanner.Text()

				if strings.HasPrefix(line, "CHANGED_DIRS=") {
					dirStr := strings.TrimPrefix(line, "CHANGED_DIRS=")
					if dirStr != "" {
						changedDirs = strings.Fields(dirStr)
					}
				} else if line == "has_changes=true" {
					hasChanges = true
				}
			}

			if err := scanner.Err(); err != nil {
				fmt.Fprintf(os.Stderr, "Error reading script output: %v\n", err)
			}

			if err := cmd2.Wait(); err != nil {
				fmt.Fprintf(os.Stderr, "Error running detection script: %v\n", err)
				os.Exit(1)
			}

			// Display results
			if hasChanges {
				fmt.Println("Changed packages that need rebuilding:")
				for _, dir := range changedDirs {
					fmt.Printf("github.com/goplus/clibs/%s\n", dir)
				}
			} else {
				fmt.Println("No package changes detected.")
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(changesCmd)
	changesCmd.Flags().StringVarP(&baseBranch, "base", "b", "", "Base branch or commit to compare against (default: HEAD~1)")
	changesCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose output from the detection script")
}

// getRepoRoot returns the root directory of the git repository
func getRepoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get repository root: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}
