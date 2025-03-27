package build

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ModuleInfo represents the JSON output from go list -m -json
type ModuleInfo struct {
	Path    string
	Version string
	Dir     string
	Sum     string
}

// ListPkgs 获取当前项目依赖的所有C库包
func ListPkgs(pkgs ...string) ([]Package, error) {
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

		// 获取模块详细信息，包括Sum字段
		cmd := exec.Command("go", "list", "-m", "-json", modPath)
		modOutput, err := cmd.Output()
		if err != nil {
			fmt.Printf("- %s: Error finding module info: %v\n", modPath, err)
			continue
		}

		// 解析JSON输出
		var moduleInfo ModuleInfo
		if err := json.Unmarshal(modOutput, &moduleInfo); err != nil {
			fmt.Printf("- %s: Error parsing module info: %v\n", modPath, err)
			continue
		}

		modLocalPath := moduleInfo.Dir
		if modLocalPath == "" {
			fmt.Printf("- %s: No local path found\n", modPath)
			continue
		}

		// 检查是否存在pkg.yaml
		pkgYamlPath := filepath.Join(modLocalPath, "pkg.yaml")
		if _, err := os.Stat(pkgYamlPath); err == nil {
			// 创建包对象
			pkg := Package{
				Mod:  modPath,
				Path: modLocalPath,
				Sum:  moduleInfo.Sum,
			}

			// 读取配置文件
			configContent, err := os.ReadFile(pkgYamlPath)
			if err != nil {
				fmt.Printf("  Error reading config %s: %v\n", modPath, err)
				continue
			}
			// 解析YAML
			var config PkgSpec
			if err := yaml.Unmarshal(configContent, &config); err != nil {
				fmt.Printf("  Error parsing YAML %s: %v\n", modPath, err)
				continue
			}
			pkg.Config = config
			packages = append(packages, pkg)
		}
	}

	return packages, nil
}
