package build

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// ConfigHashInfo represents status information
type ConfigHashInfo struct {
	ConfigHash string `json:"config_hash"`
	Timestamp  int64  `json:"timestamp"`
}

func GetBuildDirByName(baseDir, dirName, platform, arch string) string {
	return filepath.Join(baseDir, dirName, getTargetTriple(platform, arch))
}

// GetBuildDir returns the platform-specific build directory
func GetBuildDir(baseDir, platform, arch string) string {
	return filepath.Join(baseDir, BuildDirName, getTargetTriple(platform, arch))
}

// GetDownloadDir returns the download directory
func GetDownloadDir(baseDir string) string {
	return filepath.Join(baseDir, DownloadDirName)
}

// GetPrebuiltDir returns the platform-specific prebuilt directory
func GetPrebuiltDir(baseDir, platform, arch string) string {
	return filepath.Join(baseDir, PrebuiltDirName, getTargetTriple(platform, arch))
}

// checkHash 检查构建哈希是否匹配
func checkHash(dir string, config Config, build bool) (bool, error) {
	var configHash string
	if build {
		configHash = config.BuildHash()
	} else {
		configHash = config.DownloadHash()
	}

	// 检查构建哈希文件
	hashContent, err := os.ReadFile(filepath.Join(dir, BuildHashFile))
	if err != nil {
		return false, err
	}

	// 解析JSON格式
	var info ConfigHashInfo
	if err := json.Unmarshal(hashContent, &info); err != nil {
		// 如果不是有效的JSON，认为哈希不匹配
		return false, err
	}

	// 比较哈希值
	return info.ConfigHash == configHash, nil
}

func saveHash(dir string, config Config, build bool) error {
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

	// 写入哈希文件
	return os.WriteFile(filepath.Join(dir, BuildHashFile), content, 0644)
}

func md5sum(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}
