package build

import (
	"fmt"
	"net/url"
	"runtime"
)

func Build(config BuildConfig, packages []Package) error {
	if config.Goos == "" {
		config.Goos = runtime.GOOS
	}
	if config.Goarch == "" {
		config.Goarch = runtime.GOARCH
	}

	// 构建每个包
	if len(packages) == 0 {
		fmt.Println("\nBuilding C library packages:")
	} else {
		fmt.Println("\nChecking specified modules for pkg.yaml files:")
	}

	for _, pkg := range packages {
		fmt.Printf("  %#v\n", pkg)
		if !config.Force {
			if !config.Prebuilt {
				if prebuiltDir, err := pkg.checkPrebuiltStatus(config); err == nil && prebuiltDir != "" {
					continue
				}
			}
			if prebuiltDir, err := pkg.tryDownloadPrebuilt(config); err == nil && prebuiltDir != "" {
				continue
			}
		}
		dirName := BuildDirName
		if config.Prebuilt {
			dirName = PrebuiltDirName
		}
		if _, err := pkg.tryBuildPkg(config, dirName); err != nil {
			fmt.Printf("  Error processing %s: %v\n", pkg.Mod, err)
			return err
		}
	}

	return nil
}

func (pkg *Package) tryDownloadPrebuilt(config BuildConfig) (string, error) {
	name := pkg.Config.Name
	target := getTargetTriple(config.Goos, config.Goarch)
	prebuiltRootDir := GetPrebuiltDir(*pkg)
	uriEncodedTag := url.PathEscape(fmt.Sprintf("%s/%s", name, pkg.Config.Version))
	url := fmt.Sprintf("%s/%s/%s-%s-%s.tar.gz", ReleaseUrlPrefix, uriEncodedTag, name, pkg.Config.Version, target)
	fmt.Printf("  Downloading prebuilt package: %s\n", url)
	fmt.Printf("    to: %s\n", prebuiltRootDir)
	if err := fetchFromFiles([]FileSpec{{URL: url}}, prebuiltRootDir, false); err != nil {
		return "", err
	}
	return prebuiltRootDir, nil
}

func (pkg *Package) checkPrebuiltStatus(config BuildConfig) (string, error) {
	prebuiltTargetDir := GetBuildDirByName(*pkg, PrebuiltDirName, config.Goos, config.Goarch)
	if matched, err := checkHash(prebuiltTargetDir, pkg.Config, true); err != nil || !matched {
		fmt.Printf("  No prebuilt package found in %s\n", prebuiltTargetDir)
		return "", err
	}
	fmt.Printf("  Found prebuilt package in %s\n", prebuiltTargetDir)
	return prebuiltTargetDir, nil
}

// Build the library both build and prebuilt
func (pkg *Package) tryBuildPkg(config BuildConfig, buildDirName string) (string, error) {
	buildTargetDir := GetBuildDirByName(*pkg, buildDirName, config.Goos, config.Goarch)
	if !config.Force {
		if matched, err := checkHash(buildTargetDir, pkg.Config, true); err == nil && matched {
			fmt.Printf("  Found build package in %s\n", buildTargetDir)
			return buildTargetDir, nil
		}
	}
	fmt.Printf("  No build package found in %s\n", buildTargetDir)

	downloadDir := GetDownloadDir(*pkg)
	if matched, err := checkHash(downloadDir, pkg.Config, false); err != nil || !matched {
		fmt.Printf("matched: %v, err: %v\n", matched, err)
		fmt.Printf("  No download package found in %s\n", downloadDir)
		if err := pkg.fetchLib(); err != nil {
			fmt.Printf("  Error fetching library: %v\n", err)
			return "", err
		}
	}
	fmt.Printf("  Found download package in %s\n", downloadDir)

	if err := pkg.buildLib(config, buildTargetDir); err != nil {
		fmt.Printf("  Error building library: %v\n", err)
		return "", err
	}

	if err := saveHash(buildTargetDir, pkg.Config, true); err != nil {
		fmt.Printf("  Error saving hash: %v\n", err)
		return "", err
	}
	return buildTargetDir, nil
}
