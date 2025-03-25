package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	baseBranch string
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

		// Get the base commit to compare against
		baseCommit := baseBranch
		if baseCommit == "" {
			// Default to HEAD^ if no base branch specified
			baseCommit = "HEAD^"
		}

		// Get changed files
		changedFiles, err := getChangedFiles(repoRoot, baseCommit)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Find changed directories with pkg.yaml
		changedDirs := findChangedDirs(repoRoot, changedFiles)

		// Output the changed packages
		if len(changedDirs) > 0 {
			fmt.Println("Changed packages that need rebuilding:")
			for _, dir := range changedDirs {
				fmt.Printf("github.com/goplus/clibs/%s\n", dir)
			}
		} else {
			fmt.Println("No package changes detected.")
		}
	},
}

func init() {
	rootCmd.AddCommand(changesCmd)
	changesCmd.Flags().StringVarP(&baseBranch, "base", "b", "", "Base branch or commit to compare against (default: HEAD^)")
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

// getChangedFiles returns a list of files that have changed compared to the base commit
func getChangedFiles(repoRoot, baseCommit string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", baseCommit)
	cmd.Dir = repoRoot
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get changed files: %v", err)
	}
	
	if len(output) == 0 {
		return []string{}, nil
	}
	
	return strings.Split(strings.TrimSpace(string(output)), "\n"), nil
}

// findChangedDirs identifies directories that have pkg.yaml files and have changes
func findChangedDirs(repoRoot string, changedFiles []string) []string {
	dirMap := make(map[string]bool)
	
	for _, file := range changedFiles {
		// Get the top-level directory
		parts := strings.SplitN(file, "/", 2)
		if len(parts) < 2 {
			continue
		}
		
		dir := parts[0]
		
		// Skip dot directories and build directory
		if strings.HasPrefix(dir, ".") || dir == "build" {
			continue
		}
		
		// Check if this directory has a pkg.yaml file
		pkgYamlPath := filepath.Join(repoRoot, dir, "pkg.yaml")
		if _, err := os.Stat(pkgYamlPath); err == nil {
			// Either pkg.yaml itself changed or another file in the directory changed
			if file == filepath.Join(dir, "pkg.yaml") || strings.HasPrefix(file, dir+"/") {
				dirMap[dir] = true
			}
		}
	}
	
	// Convert map to slice
	var result []string
	for dir := range dirMap {
		result = append(result, dir)
	}
	
	return result
}
