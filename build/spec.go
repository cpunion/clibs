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
	Repo string `json:"repo,omitempty" yaml:"repo,omitempty"`
	Ref  string `json:"ref,omitempty" yaml:"ref,omitempty"`
}

type FileSpec struct {
	URL        string `json:"url,omitempty" yaml:"url,omitempty"`
	NoExtract  bool   `json:"no-extract,omitempty" yaml:"no-extract,omitempty"`
	ExtractDir string `json:"extract-dir,omitempty" yaml:"extract-dir,omitempty"`
}

type BuildSpec struct {
	Command string `json:"command,omitempty" yaml:"command,omitempty"`
}

type PkgSpec struct {
	Name    string     `json:"name,omitempty" yaml:"name,omitempty"`
	Version string     `json:"version,omitempty" yaml:"version,omitempty"`
	Git     *GitSpec   `json:"git,omitempty" yaml:"git,omitempty"`
	Files   []FileSpec `json:"files,omitempty" yaml:"files,omitempty"`
	Build   *BuildSpec `json:"build,omitempty" yaml:"build,omitempty"`
	Export  string     `json:"export,omitempty" yaml:"export,omitempty"`
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
