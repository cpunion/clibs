package build

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Export(config Config, libs []Lib) (exports [][2]string, err error) {
	for _, lib := range libs {
		libExports, err := lib.Export(config)
		if err != nil {
			return nil, err
		}
		exports = append(exports, libExports...)
	}
	return
}

func (p *Lib) Export(config Config) (exports [][2]string, err error) {
	if p.Config.Export == "" {
		return nil, nil
	}

	buildDir := getBuildDirByName(*p, BuildDirName, config.Goos, config.Goarch)
	buildEnv := getBuildEnv(*p, buildDir, config.Goos, config.Goarch)

	// Execute the export command using bash
	cmd := exec.Command("bash", "-c", p.Config.Export)
	cmd.Dir = p.Path
	cmd.Env = buildEnv

	output, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing export command: %v\n", err)
		fmt.Fprintf(os.Stderr, "command: %s\n", p.Config.Export)
		fmt.Printf("  Command output:\n%s\n", output)
		return nil, fmt.Errorf("export command failed: %v", err)
	}

	// Parse the output for key=value pairs
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // Skip lines that don't have the key=value format
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key != "" && value != "" {
			exports = append(exports, [2]string{key, value})
		}
	}

	if err := scanner.Err(); err != nil {
		return exports, fmt.Errorf("error scanning output: %v", err)
	}

	return exports, nil
}
