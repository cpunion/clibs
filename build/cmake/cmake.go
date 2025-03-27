package cmake

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Port from rust cmake crate

type Config struct {
	path     string
	cflags   string
	cxxflags string
	asmflags string
	defines  [][2]string
	target   string
	outDir   string
	profile  string
}

func New(path string) *Config {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return &Config{
		path: filepath.Join(dir, path),
	}
}

func (c *Config) Define(name, value string) *Config {
	c.defines = append(c.defines, [2]string{name, value})
	return c
}

func (c *Config) Cflag(flag string) *Config {
	c.cflags += " " + flag
	return c
}

func (c *Config) Cxxflag(flag string) *Config {
	c.cxxflags += " " + flag
	return c
}

func (c *Config) Asmflag(flag string) *Config {
	c.asmflags += " " + flag
	return c
}

func (c *Config) Target(target string) *Config {
	c.target = target
	return c
}

func (c *Config) OutDir(outDir string) *Config {
	c.outDir = outDir
	return c
}

func (c *Config) Profile(profile string) *Config {
	c.profile = profile
	return c
}

// Build runs the CMake configuration and build process, returning the path to the installed directory.
// This is a Go implementation of the Rust build method provided.
func (c *Config) Build() string {
	// Determine target platform
	target := c.target
	if target == "" {
		target = runtime.GOOS
		if target == "darwin" && strings.Contains(c.cxxflags, "-std=c++11") {
			target += "11"
		}
	}

	// Get output directory
	outDir := c.outDir
	if outDir == "" {
		dir, err := os.Getwd()
		if err != nil {
			panic(fmt.Sprintf("failed to get current directory: %v", err))
		}
		outDir = filepath.Join(dir, "out")
	}

	// Create build directory
	buildDir := filepath.Join(outDir, "build")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		panic(fmt.Sprintf("failed to create build directory: %v", err))
	}

	// Build the CMake configure command
	cmd := exec.Command("cmake")

	// Handle system name and processor for cross-compilation
	if runtime.GOOS != target && !c.isDefined("CMAKE_SYSTEM_NAME") {
		// Set CMAKE_SYSTEM_NAME and CMAKE_SYSTEM_PROCESSOR when cross compiling
		var systemName, systemProcessor string

		switch {
		case target == "android":
			systemName, systemProcessor = "Android", runtime.GOARCH
		case target == "darwin" && runtime.GOARCH == "amd64":
			systemName, systemProcessor = "Darwin", "x86_64"
		case target == "darwin" && runtime.GOARCH == "arm64":
			systemName, systemProcessor = "Darwin", "arm64"
		case target == "freebsd" && runtime.GOARCH == "amd64":
			systemName, systemProcessor = "FreeBSD", "amd64"
		case target == "freebsd":
			systemName, systemProcessor = "FreeBSD", runtime.GOARCH
		case target == "linux":
			systemName = "Linux"
			switch runtime.GOARCH {
			case "ppc":
				systemProcessor = "ppc"
			case "ppc64":
				systemProcessor = "ppc64"
			case "ppc64le":
				systemProcessor = "ppc64le"
			default:
				systemProcessor = runtime.GOARCH
			}
		case target == "windows" && runtime.GOARCH == "amd64":
			systemName, systemProcessor = "Windows", "AMD64"
		case target == "windows" && runtime.GOARCH == "386":
			systemName, systemProcessor = "Windows", "X86"
		case target == "windows" && runtime.GOARCH == "arm64":
			systemName, systemProcessor = "Windows", "ARM64"
		default:
			systemName, systemProcessor = target, runtime.GOARCH
		}

		c.Define("CMAKE_SYSTEM_NAME", systemName)
		c.Define("CMAKE_SYSTEM_PROCESSOR", systemProcessor)
	}

	// Handle macOS architecture
	if strings.Contains(target, "darwin") && !c.isDefined("CMAKE_OSX_ARCHITECTURES") {
		if runtime.GOARCH == "amd64" {
			cmd.Args = append(cmd.Args, "-DCMAKE_OSX_ARCHITECTURES=x86_64")
		} else if runtime.GOARCH == "arm64" {
			cmd.Args = append(cmd.Args, "-DCMAKE_OSX_ARCHITECTURES=arm64")
		}
	}

	// Add all defines to the command
	for _, def := range c.defines {
		cmd.Args = append(cmd.Args, fmt.Sprintf("-D%s=%s", def[0], def[1]))
	}

	// Set install prefix if not already defined
	if !c.isDefined("CMAKE_INSTALL_PREFIX") {
		cmd.Args = append(cmd.Args, fmt.Sprintf("-DCMAKE_INSTALL_PREFIX=%s", outDir))
	}

	// Set build type if not already defined
	profile := c.profile
	if profile == "" {
		profile = "Release"
	}

	if !c.isDefined("CMAKE_BUILD_TYPE") {
		cmd.Args = append(cmd.Args, fmt.Sprintf("-DCMAKE_BUILD_TYPE=%s", profile))
	}

	// Set compiler flags
	if !c.isDefined("CMAKE_C_FLAGS") && c.cflags != "" {
		cmd.Args = append(cmd.Args, fmt.Sprintf("-DCMAKE_C_FLAGS=%s", c.cflags))
	}

	if !c.isDefined("CMAKE_CXX_FLAGS") && c.cxxflags != "" {
		cmd.Args = append(cmd.Args, fmt.Sprintf("-DCMAKE_CXX_FLAGS=%s", c.cxxflags))
	}

	if !c.isDefined("CMAKE_ASM_FLAGS") && c.asmflags != "" {
		cmd.Args = append(cmd.Args, fmt.Sprintf("-DCMAKE_ASM_FLAGS=%s", c.asmflags))
	}

	// Set verbose make if needed
	cmd.Args = append(cmd.Args, "-DCMAKE_VERBOSE_MAKEFILE:BOOL=ON")

	// Add the path to the command
	cmd.Args = append(cmd.Args, c.path)
	cmd.Dir = buildDir

	// Run the configure command
	fmt.Printf("Running CMake configure: %s\n", strings.Join(cmd.Args, " "))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(fmt.Sprintf("CMake configure failed: %v", err))
	}

	// Build the project
	buildCmd := exec.Command("cmake", "--build", ".", "--config", profile, "--target", "install")
	buildCmd.Dir = buildDir

	// Set parallel build if NUM_JOBS environment variable is set
	if numJobs := os.Getenv("NUM_JOBS"); numJobs != "" {
		buildCmd.Args = append(buildCmd.Args, "--parallel", numJobs)
	}

	fmt.Printf("Running CMake build: %s\n", strings.Join(buildCmd.Args, " "))
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		panic(fmt.Sprintf("CMake build failed: %v", err))
	}

	fmt.Printf("Build completed successfully. Output directory: %s\n", outDir)
	return outDir
}

// isDefined checks if a CMake variable is already defined in the Config
func (c *Config) isDefined(name string) bool {
	for _, def := range c.defines {
		if def[0] == name {
			return true
		}
	}
	return false
}
