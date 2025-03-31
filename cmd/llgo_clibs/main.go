package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	// 定义子命令
	buildCmd := flag.NewFlagSet("build", flag.ExitOnError)
	exportCmd := flag.NewFlagSet("export", flag.ExitOnError)
	listCmd := flag.NewFlagSet("list", flag.ExitOnError)

	// build 命令的标志
	buildForce := buildCmd.Bool("force", false, "Force rebuild even if already built")
	buildPrebuilt := buildCmd.Bool("prebuilt", false, "Build to prebuilt directory")
	buildTags := buildCmd.String("tags", "", "A comma-separated list of build tags")

	// export 命令的标志
	exportPrebuilt := exportCmd.Bool("prebuilt", false, "Export from prebuilt directory")
	exportTags := exportCmd.String("tags", "", "A comma-separated list of build tags")

	// list 命令的标志
	listTags := listCmd.String("tags", "", "A comma-separated list of build tags")

	// 检查是否提供了子命令
	if len(os.Args) < 2 {
		fmt.Println("Expected 'build', 'export' or 'list' subcommands")
		os.Exit(1)
	}

	// 根据子命令执行相应的操作
	switch os.Args[1] {
	case "build":
		buildCmd.Parse(os.Args[2:])
		runBuild(*buildForce, *buildPrebuilt, *buildTags, buildCmd.Args())
	case "export":
		exportCmd.Parse(os.Args[2:])
		runExport(*exportPrebuilt, *exportTags, exportCmd.Args())
	case "list":
		listCmd.Parse(os.Args[2:])
		runList(*listTags, listCmd.Args())
	default:
		fmt.Printf("%s is not a valid command.\n", os.Args[1])
		fmt.Println("Expected 'build', 'export' or 'list' subcommands")
		os.Exit(1)
	}
}
