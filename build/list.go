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

// LibInfo represents the JSON output from go list -m -json
type LibInfo struct {
	Path    string
	Version string
	Dir     string
	Sum     string
}

// ListLibs 获取当前项目依赖的所有C库包
func ListLibs(patterns ...string) ([]Lib, error) {
	var libs []Lib

	// 获取项目依赖的所有模块或指定的模块
	var clibs []string
	if len(patterns) == 0 {
		// 获取所有模块
		cmd := exec.Command("go", "list", "-m", "all")
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to list clibs: %v", err)
		}
		clibs = strings.Split(string(output), "\n")
	} else {
		// 获取指定的模块
		fmt.Printf("Executing: go list %s\n", strings.Join(patterns, " "))
		args := append([]string{"list"}, patterns...)
		cmd := exec.Command("go", args...)
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to list specified clibs: %v", err)
		}
		clibs = strings.Split(strings.TrimSpace(string(output)), "\n")
	}

	// 处理模块列表
	for _, lib := range clibs {
		if lib == "" {
			continue
		}

		// 提取模块路径（例如 github.com/user/repo）
		libPath := lib
		if idx := strings.Index(lib, " "); idx > 0 {
			libPath = strings.TrimSpace(lib[:idx])
		}

		// 获取模块详细信息，包括Sum字段
		cmd := exec.Command("go", "list", "-m", "-json", libPath)
		libOutput, err := cmd.Output()
		if err != nil {
			fmt.Printf("- %s: Error finding libule info: %v\n", libPath, err)
			continue
		}

		// 解析JSON输出
		var libuleInfo LibInfo
		if err := json.Unmarshal(libOutput, &libuleInfo); err != nil {
			fmt.Printf("- %s: Error parsing libule info: %v\n", libPath, err)
			continue
		}

		libLocalPath := libuleInfo.Dir
		if libLocalPath == "" {
			fmt.Printf("- %s: No local path found\n", libPath)
			continue
		}

		// 检查是否存在lib.yaml
		libYamlPath := filepath.Join(libLocalPath, "lib.yaml")
		fmt.Printf("  Checking for lib.yaml: %s\n", libYamlPath)
		if _, err := os.Stat(libYamlPath); err == nil {
			// 创建包对象
			lib := Lib{
				ModName: libPath,
				Path:    libLocalPath,
				Sum:     libuleInfo.Sum,
			}

			// 读取配置文件
			configContent, err := os.ReadFile(libYamlPath)
			if err != nil {
				fmt.Printf("  Error reading config %s: %v\n", libPath, err)
				continue
			}
			// 解析YAML
			var config LibSpec
			if err := yaml.Unmarshal(configContent, &config); err != nil {
				fmt.Printf("  Error parsing YAML %s: %v\n", libPath, err)
				continue
			}
			fmt.Printf("  Found lib.yaml: %s at %s\n", libPath, libYamlPath)
			fmt.Printf("  Config: %v\n", config)
			lib.Config = config
			libs = append(libs, lib)
		}
	}

	return libs, nil
}
