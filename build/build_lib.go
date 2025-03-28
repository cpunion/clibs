package build

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// getBuildEnv prepares build environment variables
func getBuildEnv(lib Lib, buildDir, platform, arch string) []string {
	// Generate target triple
	targetTriple := getTargetTriple(platform, arch)

	// Generate build flags
	cflags, ldflags := getBuildFlags(targetTriple)

	downloadDir := getDownloadDir(lib)

	// Create environment variables
	env := os.Environ()
	env = append(env,
		fmt.Sprintf("%s=%s", EnvPackageDir, lib.Path),
		fmt.Sprintf("%s=%s", EnvDownloadDir, downloadDir),
		fmt.Sprintf("%s=%s", EnvBuildGoos, platform),
		fmt.Sprintf("%s=%s", EnvBuildGoarch, arch),
		fmt.Sprintf("%s=%s", EnvBuildTarget, targetTriple),
		fmt.Sprintf("%s=%s", EnvBuildCflags, cflags),
		fmt.Sprintf("%s=%s", EnvBuildLdflags, ldflags),
		fmt.Sprintf("%s=%s", EnvBuildDir, buildDir),
	)

	return env
}

// getTargetTriple generates LLVM target triple based on platform and architecture
func getTargetTriple(platform, arch string) string {
	// Platform mapping
	platformMap := map[string]string{
		"darwin":  "apple-darwin",
		"linux":   "unknown-linux-gnu",
		"windows": "pc-windows-msvc",
		"js":      "wasi",   // WebAssembly (browser)
		"wasip1":  "wasip1", // WebAssembly (non-browser)
	}

	// Architecture mapping
	archMap := map[string]string{
		"amd64":   "x86_64",
		"386":     "i386",
		"arm":     "arm",
		"arm64":   "aarch64",
		"mips":    "mips",
		"mips64":  "mips64",
		"wasm":    "wasm32",
		"riscv":   "riscv64",
		"riscv64": "riscv64",
	}

	// Get platform string
	platformStr, ok := platformMap[platform]
	if !ok {
		platformStr = "unknown-unknown"
	}

	// Get architecture string
	archStr, ok := archMap[arch]
	if !ok {
		archStr = arch
	}

	// Special case for WebAssembly
	if platform == "js" || platform == "wasi" {
		return fmt.Sprintf("wasm32-%s", platformStr)
	}

	return fmt.Sprintf("%s-%s", archStr, platformStr)
}

// getBuildFlags generates build flags based on target triple
func getBuildFlags(targetTriple string) (cflags, ldflags string) {
	// Default flags
	cflags = "-O2"
	ldflags = ""

	// Add target-specific flags
	if strings.Contains(targetTriple, "wasm32") {
		cflags += " -D__wasm__"
	} else if strings.Contains(targetTriple, "windows") {
		cflags += " -D_WIN32"
	} else if strings.Contains(targetTriple, "darwin") {
		cflags += " -D__APPLE__"
	} else if strings.Contains(targetTriple, "linux") {
		cflags += " -D__linux__"
	}

	return
}

// buildLib builds the library using the appropriate build method
func (lib *Lib) buildLib(config Config, buildDir string) error {
	// Get download directory
	downloadDir := getDownloadDir(*lib)
	if _, err := os.Stat(downloadDir); err != nil {
		// If download directory doesn't exist, try to create it
		if os.IsNotExist(err) {
			if err := os.MkdirAll(downloadDir, 0755); err != nil {
				return fmt.Errorf("failed to create download directory: %v", err)
			}
		} else {
			return fmt.Errorf("failed to check download directory: %v", err)
		}
	}

	// Create build directory if it doesn't exist
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return fmt.Errorf("failed to create build directory: %v", err)
	}

	fmt.Printf("  Build directory: %s\n", buildDir)

	// Check if we need to download files
	if lib.Config.Git != nil {
		// TODO: Implement Git checkout
		return fmt.Errorf("git checkout not implemented yet")
	} else if len(lib.Config.Files) > 0 {
		// Download files
		if err := lib.fetchLib(); err != nil {
			return fmt.Errorf("failed to fetch library: %v", err)
		}
	}

	// If there's a build command, execute it
	if lib.Config.Build != nil && lib.Config.Build.Command != "" {
		fmt.Printf("  Executing build command\n")

		// Get environment variables
		env := getBuildEnv(*lib, buildDir, config.Goos, config.Goarch)

		// Create the build command
		cmd := exec.Command("bash", "-c", lib.Config.Build.Command)
		cmd.Dir = downloadDir
		cmd.Env = env
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// Execute the build command
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("build command failed: %v", err)
		}
	}

	// Write hash file to mark successful build
	if err := saveHash(buildDir, lib.Config, true); err != nil {
		return fmt.Errorf("failed to write hash file: %v", err)
	}

	return nil
}
