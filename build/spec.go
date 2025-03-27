package build

import "encoding/json"

// StatusFile constants for tracking library status
const (
	BuildDirName    = "_build"
	DownloadDirName = "_download"
	PrebuiltDirName = "_prebuilt"
	BuildHashFile   = "_llgo_clib_build_config_hash.json"

	PkgConfigFile = "pkg.yaml"

	ReleaseUrlPrefix = "https://github.com/cpunion/clibs/releases/download"
)

type GitSpec struct {
	Repo string
	Ref  string
}

type FileSpec struct {
	URL string
}

type BuildSpec struct {
	Command string
}

type PkgSpec struct {
	Name    string
	Version string
	Git     *GitSpec
	Files   []FileSpec
	Build   *BuildSpec
	Export  string
}

func (c *PkgSpec) DownloadHash() string {
	hashConfig := *c
	hashConfig.Build = nil
	hashConfig.Export = ""
	data, err := json.Marshal(hashConfig)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (c *PkgSpec) BuildHash() string {
	data, err := json.Marshal(*c)
	if err != nil {
		panic(err)
	}
	return string(data)
}

// Package represents a package to be built

type Package struct {
	Mod    string
	Path   string
	Sum    string
	Config PkgSpec
}

type BuildConfig struct {
	Goos     string
	Goarch   string
	Prebuilt bool
	Force    bool
	Verbose  bool
}
