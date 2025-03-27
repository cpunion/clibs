package build

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ConfigHashInfo represents status information
type ConfigHashInfo struct {
	ConfigHash string `json:"config_hash"`
	Timestamp  int64  `json:"timestamp"`
}

func GetBuildDirByName(pkg Package, dirName, platform, arch string) string {
	return filepath.Join(getBuildBaseDir(pkg), dirName, getTargetTriple(platform, arch))
}

// GetDownloadDir returns the download directory
func GetDownloadDir(pkg Package) string {
	return filepath.Join(getBuildBaseDir(pkg), DownloadDirName)
}

func GetPrebuiltDir(pkg Package) string {
	return filepath.Join(getBuildBaseDir(pkg), PrebuiltDirName)
}

func getBuildBaseDir(pkg Package) string {
	if pkg.Sum == "" {
		return pkg.Path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	subDir := strings.TrimLeft(pkg.Sum, "h1:")
	return filepath.Join(home, ".llgo/", "clibs_build", pkg.Mod, subDir)
}

// checkHash 检查构建哈希是否匹配
func checkHash(dir string, config PkgSpec, build bool) (bool, error) {
	var configHash string
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

	// 解析JSON格式
	var info ConfigHashInfo
	if err := json.Unmarshal(hashContent, &info); err != nil {
		// 如果不是有效的JSON，认为哈希不匹配
		fmt.Printf("parse hash file failed: %v, %s", err, filepath.Join(dir, BuildHashFile))
		return false, err
	}

	// 比较哈希值
	fmt.Printf("  Checking hash, equal: %v, %s, %s\n", info.ConfigHash == configHash, info.ConfigHash, configHash)
	return info.ConfigHash == configHash, nil
}

func saveHash(dir string, config PkgSpec, build bool) error {
	var configHash string
	if build {
		configHash = config.BuildHash()
	} else {
		configHash = config.DownloadHash()
	}

	// 创建状态信息
	info := ConfigHashInfo{
		ConfigHash: configHash,
		Timestamp:  time.Now().Unix(),
	}

	// 序列化为JSON
	content, err := json.Marshal(info)
	if err != nil {
		return err
	}

	fmt.Printf("  Saving hash: %s to %s\n", configHash, filepath.Join(dir, BuildHashFile))
	// 写入哈希文件
	return os.WriteFile(filepath.Join(dir, BuildHashFile), content, 0644)
}

func md5sum(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}
