package build

import (
	"fmt"
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
			// TODO: download prebuilt package if it is a clean tag
			// if prebuiltDir, err := p.tryDownloadPrebuilt(config); err == nil && prebuiltDir != "" {
			// 	continue
			// }
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

// func tryDownloadPrebuilt(pkg Package, goos, goarch string) (string, error) {
// 	// v0.10.1/llgo0.10.1.darwin-amd64.tar.gz
// 	nameToks := strings.Split(pkg.Mod, "/")
// 	name := nameToks[len(nameToks)-1]
// 	target := getTargetTriple(goos, goarch)
// 	url := fmt.Sprintf("%s/%s/%s/%s/%s.tar.gz", PrebuiltDownloadPrefix, name, pkg.Config.Version, name, target)
// 	fmt.Printf("  Downloading prebuilt package: %s\n", url)
// }

func (pkg *Package) checkPrebuiltStatus(config BuildConfig) (string, error) {
	prebuiltDir := GetPrebuiltDir(pkg.Path, config.Goos, config.Goarch)
	if matched, err := checkHash(prebuiltDir, pkg.Config, true); err != nil || !matched {
		fmt.Printf("  No prebuilt package found in %s\n", prebuiltDir)
		return "", err
	}
	fmt.Printf("  Found prebuilt package in %s\n", prebuiltDir)
	return prebuiltDir, nil
}

// Build the library both build and prebuilt
func (pkg *Package) tryBuildPkg(config BuildConfig, buildDirName string) (string, error) {
	buildDir := GetBuildDirByName(pkg.Path, buildDirName, config.Goos, config.Goarch)
	if matched, err := checkHash(buildDir, pkg.Config, true); err == nil && matched {
		fmt.Printf("  Found build package in %s\n", buildDir)
		return buildDir, nil
	}
	fmt.Printf("  No build package found in %s\n", buildDir)

	downloadDir := GetDownloadDir(pkg.Path)
	if matched, err := checkHash(downloadDir, pkg.Config, false); err != nil || !matched {
		fmt.Printf("matched: %v, err: %v\n", matched, err)
		fmt.Printf("  No download package found in %s\n", downloadDir)
		// panic("===")
		if err := pkg.fetchLib(); err != nil {
			fmt.Printf("  Error fetching library: %v\n", err)
			return "", err
		}
	}
	fmt.Printf("  Found download package in %s\n", downloadDir)

	if err := pkg.buildLib(config, buildDir); err != nil {
		fmt.Printf("  Error building library: %v\n", err)
		return "", err
	}

	if err := saveHash(buildDir, pkg.Config, true); err != nil {
		fmt.Printf("  Error saving hash: %v\n", err)
		return "", err
	}
	return buildDir, nil
}
