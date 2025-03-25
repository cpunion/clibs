package build

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// getTargetTriple 根据平台和架构生成LLVM目标三元组
func getTargetTriple(platform, arch string) string {
	// 平台映射
	platformMap := map[string]string{
		"darwin":  "apple-darwin",
		"linux":   "unknown-linux-gnu",
		"windows": "pc-windows-msvc",
		"js":      "unknown-emscripten", // WebAssembly (浏览器)
		"wasi":    "unknown-wasi",       // WebAssembly (非浏览器)
	}

	// 架构映射
	archMap := map[string]string{
		"amd64":  "x86_64",
		"386":    "i386",
		"arm":    "arm",
		"arm64":  "aarch64",
		"mips":   "mips",
		"mips64": "mips64",
		"wasm":   "wasm32",
	}

	// 获取平台和架构的映射值，如果没有则使用原值
	platformStr, ok := platformMap[platform]
	if !ok {
		platformStr = platform
	}

	archStr, ok := archMap[arch]
	if !ok {
		archStr = arch
	}

	// 特殊情况处理
	if platform == "js" || platform == "wasi" {
		// 对于WebAssembly，使用wasm32-unknown-emscripten或wasm32-unknown-wasi
		return fmt.Sprintf("wasm32-%s", platformStr)
	}

	// 生成三元组
	return fmt.Sprintf("%s-%s", archStr, platformStr)
}

// getBuildFlags 根据目标三元组生成CFLAGS和LDFLAGS
func getBuildFlags(triple string) (cflags string, ldflags string) {
	// 基本标志
	cflags = "-O2"
	ldflags = ""

	// 根据三元组添加特定标志
	if strings.Contains(triple, "wasm32") {
		// WebAssembly特定标志
		cflags = fmt.Sprintf("%s -D__WASM__ -flto", cflags)
		ldflags = "-flto"
	} else if strings.Contains(triple, "windows") {
		// Windows特定标志
		cflags = fmt.Sprintf("%s -D_WIN32", cflags)
	} else if strings.Contains(triple, "darwin") {
		// macOS特定标志
		cflags = fmt.Sprintf("%s -D__APPLE__", cflags)
	} else if strings.Contains(triple, "linux") {
		// Linux特定标志
		cflags = fmt.Sprintf("%s -D__linux__", cflags)
	}

	return cflags, ldflags
}

// getBuildEnv 准备构建环境变量
func getBuildEnv(modLocalPath, buildDir, platform, arch string) []string {
	// 生成目标三元组
	targetTriple := getTargetTriple(platform, arch)

	// 生成构建标志
	cflags, ldflags := getBuildFlags(targetTriple)

	// 创建环境变量
	env := os.Environ()
	env = append(env,
		fmt.Sprintf("CLIBS_PACKAGE_DIR=%s", modLocalPath),
		fmt.Sprintf("CLIBS_BUILD_TARGET=%s", targetTriple),
		fmt.Sprintf("CLIBS_BUILD_CFLAGS=%s", cflags),
		fmt.Sprintf("CLIBS_BUILD_LDFLAGS=%s", ldflags),
		fmt.Sprintf("CLIBS_BUILD_DIR=%s", buildDir),
	)

	return env
}

// buildLib builds the library using the appropriate build method
func buildLib(config *Config, buildDir, modLocalPath string, platform, arch string) error {
	// 获取下载目录
	downloadDir := GetDownloadDir(modLocalPath)
	if _, err := os.Stat(downloadDir); err != nil {
		// 如果下载目录不存在，尝试创建它
		if os.IsNotExist(err) {
			if err := os.MkdirAll(downloadDir, 0755); err != nil {
				return fmt.Errorf("failed to create download directory: %v", err)
			}
			// 目录创建成功，但因为没有下载内容，返回错误提示需要先获取源码
			return fmt.Errorf("download directory created but empty, please fetch library source first")
		}
		return fmt.Errorf("download directory access error: %v", err)
	}

	// 确保构建目录存在
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return fmt.Errorf("failed to create build directory: %v", err)
	}

	fmt.Printf("  Build directory: %s\n", buildDir)

	// 首先检查配置中是否有构建命令
	if config.Build != nil && config.Build.Command != "" {
		fmt.Printf("  Executing build command: %s\n", config.Build.Command)

		// 获取构建环境变量
		buildEnv := getBuildEnv(modLocalPath, buildDir, platform, arch)

		// 对于包含shell特殊字符的命令，使用shell执行
		cmd := exec.Command("bash", "-c", config.Build.Command)
		cmd.Dir = downloadDir
		cmd.Env = buildEnv

		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("  Build command error: %s\n", output)
			return fmt.Errorf("build command failed: %v - %s", err, output)
		}

		fmt.Printf("  Build command output:\n%s\n", output)
		return nil
	}

	return fmt.Errorf("no build command specified in configuration")
}
