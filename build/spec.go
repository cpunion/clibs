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

type GitConfig struct {
	Repo string
	Ref  string
}

type FileConfig struct {
	URL string
}

type BuildConfig struct {
	Command string
}

type Config struct {
	Version string
	Git     *GitConfig
	Files   []FileConfig
	Build   *BuildConfig
}

func (c *Config) DownloadHash() string {
	hashConfig := *c
	hashConfig.Build = nil
	data, err := json.Marshal(hashConfig)
	if err != nil {
		panic(err)
	}
	return md5sum(data)
}

func (c *Config) BuildHash() string {
	data, err := json.Marshal(*c)
	if err != nil {
		panic(err)
	}
	return md5sum(data)
}

type Package struct {
	Mod    string
	Path   string
	Config Config
}
