package main

import (
	"fmt"
	"os"

	"github.com/cpunion/clibs"
)

// runList 执行 list 命令
func runList(tags string, args []string) {
	// 准备 tags 参数，使用 Go 标准格式
	var tagArgs []string
	if tags != "" {
		tagArgs = []string{"-tags", tags}
	}

	// Get libs from specified libs or all libs if none specified
	libs, err := clibs.ListLibs(tagArgs, args...)
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
}
