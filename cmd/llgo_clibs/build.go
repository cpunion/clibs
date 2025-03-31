package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/cpunion/clibs"
)

// runBuild 执行 build 命令
func runBuild(force, prebuilt bool, tags string, args []string) {
	goos := os.Getenv("GOOS")
	goarch := os.Getenv("GOARCH")
	if goos == "" {
		goos = runtime.GOOS
	}
	if goarch == "" {
		goarch = runtime.GOARCH
	}

	fmt.Printf("Build: GOOS: %s, GOARCH: %s, Force: %v, Prebuilt: %v, Tags: %v\n",
		goos, goarch, force, prebuilt, tags)

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
		Force:    force,
		Prebuilt: prebuilt,
		Tags:     tagArgs,
	}

	err = clibs.Build(buildConfig, libs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
