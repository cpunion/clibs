package build

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
	Repo string `json:"repo,omitempty"`
	Ref  string `json:"ref,omitempty"`
}

type FileSpec struct {
	URL string `json:"url,omitempty"`
}

type BuildSpec struct {
	Command string `json:"command,omitempty"`
}

type PkgSpec struct {
	Name    string     `json:"name,omitempty"`
	Version string     `json:"version,omitempty"`
	Git     *GitSpec   `json:"git,omitempty"`
	Files   []FileSpec `json:"files,omitempty"`
	Build   *BuildSpec `json:"build,omitempty"`
	Export  string     `json:"export,omitempty"`
}

func (c *PkgSpec) DownloadHash() PkgSpec {
	hashConfig := *c
	hashConfig.Build = nil
	hashConfig.Export = ""
	return hashConfig
}

func (c *PkgSpec) BuildHash() PkgSpec {
	return *c
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
