package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

// GetClibPkgs 获取当前项目依赖的所有C库包
func GetClibPkgs(pkgs ...string) ([]Package, error) {
	var packages []Package

	// 获取项目依赖的所有模块或指定的模块
	var modules []string
	if len(pkgs) == 0 {
		// 获取所有模块
		cmd := exec.Command("go", "list", "-m", "all")
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to list modules: %v", err)
		}
		modules = strings.Split(string(output), "\n")
	} else {
		// 获取指定的模块
		fmt.Printf("Executing: go list -m %s\n", strings.Join(pkgs, " "))
		args := append([]string{"list", "-m"}, pkgs...)
		cmd := exec.Command("go", args...)
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to list specified modules: %v", err)
		}
		modules = strings.Split(strings.TrimSpace(string(output)), "\n")
	}

	// 处理模块列表
	for _, mod := range modules {
		if mod == "" {
			continue
		}

		// 提取模块路径（例如 github.com/user/repo）
		modPath := mod
		if idx := strings.Index(mod, " "); idx > 0 {
			modPath = strings.TrimSpace(mod[:idx])
		}

		// 获取模块本地路径
		cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", modPath)
		modOutput, err := cmd.Output()
		if err != nil {
			fmt.Printf("- %s: Error finding local path: %v\n", modPath, err)
			continue
		}

		modLocalPath := strings.TrimSpace(string(modOutput))

		// 检查是否存在pkg.yaml
		pkgYamlPath := filepath.Join(modLocalPath, "pkg.yaml")
		if _, err := os.Stat(pkgYamlPath); err == nil {
			if len(pkgs) == 0 {
				fmt.Printf("- Found pkg.yaml in %s\n", modPath)
			} else {
				fmt.Printf("- Found pkg.yaml in %s:\n", modPath)
			}

			// 创建包对象
			pkg := Package{
				Mod:  modPath,
				Path: modLocalPath,
			}

			// 读取配置文件
			configContent, err := os.ReadFile(pkgYamlPath)
			if err != nil {
				fmt.Printf("  Error reading config %s: %v\n", modPath, err)
				continue
			}
			// 解析YAML
			var config Config
			if err := yaml.Unmarshal(configContent, &config); err != nil {
				fmt.Printf("  Error parsing YAML %s: %v\n", modPath, err)
				continue
			}

			pkg.Config = config
			packages = append(packages, pkg)
		} else {
			fmt.Printf("- %s: No pkg.yaml found\n", modPath)
		}
	}

	return packages, nil
}

func Build(goos, goarch string, pkgs ...string) error {
	if goos == "" {
		goos = runtime.GOOS
	}
	if goarch == "" {
		goarch = runtime.GOARCH
	}

	// 获取要处理的C库包
	packages, err := GetClibPkgs(pkgs...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting C library packages: %v\n", err)
		return err
	}

	// 构建每个包
	if len(pkgs) == 0 {
		fmt.Println("\nBuilding C library packages:")
	} else {
		fmt.Println("\nChecking specified modules for pkg.yaml files:")
	}

	for _, pkg := range packages {
		fmt.Printf("  %#v\n", pkg)
		if prebuiltDir, err := checkPrebuiltStatus(pkg, goos, goarch); err == nil && prebuiltDir != "" {
			continue
		}
		// if prebuiltDir, err := tryDownloadPrebuilt(pkg, goos, goarch); err == nil && prebuiltDir != "" {
		// 	continue
		// }
		// TODO: download prebuilt package if it is a clean tag
		if _, err := tryBuildPkg(pkg, BuildDirName, goos, goarch); err != nil {
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

func Prebuild(goos, goarch string, pkgs ...string) error {
	if goos == "" {
		goos = runtime.GOOS
	}
	if goarch == "" {
		goarch = runtime.GOARCH
	}

	// 获取要处理的C库包
	packages, err := GetClibPkgs(pkgs...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting C library packages: %v\n", err)
		return err
	}

	// 构建每个包
	if len(pkgs) == 0 {
		fmt.Println("\nBuilding C library packages:")
	} else {
		fmt.Println("\nChecking specified modules for pkg.yaml files:")
	}

	for _, pkg := range packages {
		fmt.Printf("  %#v\n", pkg)
		if _, err := tryBuildPkg(pkg, PrebuiltDirName, goos, goarch); err != nil {
			return fmt.Errorf("error prebuilding %s: %v", pkg.Mod, err)
		}
	}

	return nil
}

func checkPrebuiltStatus(pkg Package, goos, goarch string) (string, error) {
	prebuiltDir := GetPrebuiltDir(pkg.Path, goos, goarch)
	if matched, err := checkHash(prebuiltDir, pkg.Config, true); err != nil || !matched {
		fmt.Printf("  No prebuilt package found in %s\n", prebuiltDir)
		return "", err
	}
	fmt.Printf("  Found prebuilt package in %s\n", prebuiltDir)
	return prebuiltDir, nil
}

// Build the library both build and prebuilt
func tryBuildPkg(pkg Package, buildDirName, goos, goarch string) (string, error) {
	buildDir := GetBuildDirByName(pkg.Path, buildDirName, goos, goarch)
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
		if err := fetchLib(&pkg.Config, pkg.Path); err != nil {
			fmt.Printf("  Error fetching library: %v\n", err)
			return "", err
		}
	}
	fmt.Printf("  Found download package in %s\n", downloadDir)

	if err := buildLib(&pkg.Config, buildDir, pkg.Path, goos, goarch); err != nil {
		fmt.Printf("  Error building library: %v\n", err)
		return "", err
	}

	if err := saveHash(buildDir, pkg.Config, true); err != nil {
		fmt.Printf("  Error saving hash: %v\n", err)
		return "", err
	}
	return buildDir, nil
}
