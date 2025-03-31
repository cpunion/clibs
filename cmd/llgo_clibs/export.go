package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/cpunion/clibs"
)

// runExport 执行 export 命令
func runExport(prebuilt bool, tags string, args []string) {
	goos := os.Getenv("GOOS")
	goarch := os.Getenv("GOARCH")
	if goos == "" {
		goos = runtime.GOOS
	}
	if goarch == "" {
		goarch = runtime.GOARCH
	}

	fmt.Printf("Export: GOOS: %s, GOARCH: %s, Prebuilt: %v, Tags: %v\n",
		goos, goarch, prebuilt, tags)

	// 准备 tags 参数，使用 Go 标准格式
	var tagArgs []string
	if tags != "" {
		tagArgs = []string{"-tags", tags}
	}

	libs, err := clibs.ListLibs(tagArgs, args...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting C library libs: %v\n", err)
		os.Exit(1)
	}

	buildConfig := clibs.Config{
		Goos:     goos,
		Goarch:   goarch,
		Prebuilt: prebuilt,
		Tags:     tagArgs,
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
}
