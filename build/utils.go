package build

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func getBuildDirByName(lib Lib, dirName, platform, arch string) string {
	return filepath.Join(getBuildBaseDir(lib), dirName, getTargetTriple(platform, arch))
}

// getDownloadDir returns the download directory
func getDownloadDir(lib Lib) string {
	return filepath.Join(getBuildBaseDir(lib), DownloadDirName)
}

func getPrebuiltDir(lib Lib) string {
	return filepath.Join(getBuildBaseDir(lib), PrebuiltDirName)
}

func getBuildBaseDir(lib Lib) string {
	if lib.Sum == "" {
		return lib.Path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	subDir := strings.TrimLeft(lib.Sum, "h1:")
	return filepath.Join(home, ".llgo/", "clibs_build", lib.ModName, subDir)
}

// checkHash 检查构建哈希是否匹配
func checkHash(dir string, config LibSpec, build bool) (bool, error) {
	var configHash LibSpec
	if build {
		configHash = config.BuildHash()
	} else {
		configHash = config.DownloadHash()
	}

	// 检查构建哈希文件
	hashContent, err := os.ReadFile(filepath.Join(dir, BuildHashFile))
	if err != nil {
		fmt.Printf("read hash file failed: %v, %s", err, filepath.Join(dir, BuildHashFile))
		return false, err
	}

	hash, err := json.MarshalIndent(configHash, "", "  ")
	if err != nil {
		return false, err
	}
	hashStr := string(hash)

	// 比较哈希值
	fmt.Printf("  Checking hash, equal: %v, %s, %s\n", hashStr == string(hashContent), hashStr, string(hashContent))
	return hashStr == string(hashContent), nil
}

func saveHash(dir string, config LibSpec, build bool) error {
	var configHash LibSpec
	if build {
		configHash = config.BuildHash()
	} else {
		configHash = config.DownloadHash()
	}

	// 序列化为JSON
	content, err := json.MarshalIndent(configHash, "", "  ")
	if err != nil {
		return err
	}

	fmt.Printf("  Saving hash: %#v\n     to %s\n", configHash, filepath.Join(dir, BuildHashFile))
	// 写入哈希文件
	return os.WriteFile(filepath.Join(dir, BuildHashFile), content, 0644)
}
