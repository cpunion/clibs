package build

import "encoding/json"

// StatusFile constants for tracking library status
const (
	BuildDirName    = "_build"
	DownloadDirName = "_download"
	PrebuiltDirName = "_prebuilt"
	BuildHashFile   = "_llgo_clib_build_config_hash.json"

	PkgConfigFile = "pkg.yaml"

	ReleaseUrlPrefix = "https://api.github.com/repos/goplus/llgo/releases/tags"
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
	Version string
	Git     *GitSpec
	Files   []FileSpec
	Build   *BuildSpec
	Export  string
}

func (c *PkgSpec) DownloadHash() string {
	hashConfig := *c
	hashConfig.Build = nil
	data, err := json.Marshal(hashConfig)
	if err != nil {
		panic(err)
	}
	return md5sum(data)
}

func (c *PkgSpec) BuildHash() string {
	data, err := json.Marshal(*c)
	if err != nil {
		panic(err)
	}
	return md5sum(data)
}

// Package represents a package to be built

type Package struct {
	Mod    string
	Path   string
	Config PkgSpec
}

type BuildConfig struct {
	Goos     string
	Goarch   string
	Prebuilt bool
	Force    bool
	Verbose  bool
}
