package clibs

import (
	"fmt"
	"net/url"
	"runtime"
)

func Build(config Config, libs []*Lib) error {
	if config.Goos == "" {
		config.Goos = runtime.GOOS
	}
	if config.Goarch == "" {
		config.Goarch = runtime.GOARCH
	}

	if len(libs) == 0 {
		fmt.Println("\nBuilding C library libs:")
	} else {
		fmt.Println("\nChecking specified libs for lib.yaml files:")
	}

	for _, lib := range libs {
		fmt.Printf("  %#v\n", lib)
		if !config.Force {
			if !config.Prebuilt {
				if prebuiltDir, err := lib.checkPrebuiltStatus(config); err == nil && prebuiltDir != "" {
					continue
				}
			}
			if prebuiltDir, err := lib.tryDownloadPrebuilt(config); err == nil && prebuiltDir != "" {
				continue
			}
		}
		dirName := BuildDirName
		if config.Prebuilt {
			dirName = PrebuiltDirName
		}
		if _, err := lib.tryBuildLib(config, dirName); err != nil {
			fmt.Printf("  Error processing %s: %v\n", lib.ModName, err)
			return err
		}
	}

	return nil
}

func (lib *Lib) tryDownloadPrebuilt(config Config) (string, error) {
	name := lib.Config.Name
	target := getTargetTriple(config.Goos, config.Goarch)
	prebuiltRootDir := getPrebuiltDir(lib)
	uriEncodedTag := url.PathEscape(fmt.Sprintf("%s/%s", name, lib.Config.Version))
	url := fmt.Sprintf("%s/%s/%s-%s-%s.tar.gz", ReleaseUrlPrefix, uriEncodedTag, name, lib.Config.Version, target)
	fmt.Printf("  Downloading prebuilt lib: %s\n", url)
	fmt.Printf("    to: %s\n", prebuiltRootDir)
	if err := fetchFromFiles([]FileSpec{{URL: url}}, prebuiltRootDir, false); err != nil {
		return "", err
	}
	prebuiltTargetDir := getBuildDirByName(lib, PrebuiltDirName, config.Goos, config.Goarch)
	lib.Env = getBuildEnv(lib, prebuiltTargetDir, config.Goos, config.Goarch)
	return prebuiltTargetDir, nil
}

func (lib *Lib) checkPrebuiltStatus(config Config) (string, error) {
	prebuiltTargetDir := getBuildDirByName(lib, PrebuiltDirName, config.Goos, config.Goarch)
	if matched, err := checkHash(prebuiltTargetDir, lib.Config, true); err != nil || !matched {
		fmt.Printf("  No prebuilt lib  found in %s\n", prebuiltTargetDir)
		return "", err
	}
	fmt.Printf("  Found prebuilt lib in %s\n", prebuiltTargetDir)
	lib.Env = getBuildEnv(lib, prebuiltTargetDir, config.Goos, config.Goarch)
	return prebuiltTargetDir, nil
}

// Build the library both build and prebuilt
func (lib *Lib) tryBuildLib(config Config, buildDirName string) (string, error) {
	buildTargetDir := getBuildDirByName(lib, buildDirName, config.Goos, config.Goarch)
	if !config.Force {
		if matched, err := checkHash(buildTargetDir, lib.Config, true); err == nil && matched {
			fmt.Printf("  Found built lib in %s\n", buildTargetDir)
			return buildTargetDir, nil
		}
	}
	fmt.Printf("  No built lib found in %s\n", buildTargetDir)

	downloadDir := getDownloadDir(lib)
	if matched, err := checkHash(downloadDir, lib.Config, false); err != nil || !matched {
		fmt.Printf("matched: %v, err: %v\n", matched, err)
		fmt.Printf("  No download lib found in %s\n", downloadDir)
		if err := lib.fetchLib(); err != nil {
			fmt.Printf("  Error fetching library: %v\n", err)
			return "", err
		}
	}
	fmt.Printf("  Found download lib in %s\n", downloadDir)

	if err := lib.buildLib(config, buildTargetDir); err != nil {
		fmt.Printf("  Error building lib: %v\n", err)
		return "", err
	}

	if err := saveHash(buildTargetDir, lib.Config, true); err != nil {
		fmt.Printf("  Error saving hash: %v\n", err)
		return "", err
	}
	return buildTargetDir, nil
}
