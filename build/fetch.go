package build

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// fetchLib fetches the library source based on the configuration
func (p *Lib) fetchLib() error {
	// 获取下载目录
	downloadDir := GetDownloadDir(*p)

	// 临时下载目录，用于保证原子性
	downloadTmpDir := downloadDir + "_tmp"

	// 创建所需的目录
	if err := os.MkdirAll(filepath.Dir(downloadDir), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory for download: %v", err)
	}

	// 清理旧的临时目录（如果存在）
	if err := os.RemoveAll(downloadTmpDir); err != nil {
		return fmt.Errorf("failed to clean temporary download directory: %v", err)
	}

	// 创建新的临时目录
	if err := os.MkdirAll(downloadTmpDir, 0755); err != nil {
		return fmt.Errorf("failed to create temporary download directory: %v", err)
	}

	var fetchErr error

	// 根据配置选择下载方式
	if p.Config.Git != nil && p.Config.Git.Repo != "" {
		fmt.Printf("  Fetching from git repository: %s\n", p.Config.Git.Repo)
		fetchErr = fetchFromGit(p.Config.Git, downloadTmpDir)
	} else if len(p.Config.Files) > 0 {
		fmt.Printf("  Fetching from files\n")
		fetchErr = fetchFromFiles(p.Config.Files, downloadTmpDir, true)
	} else {
		fetchErr = fmt.Errorf("no valid fetch configuration found")
	}

	// 如果下载失败，清理临时目录并返回错误
	if fetchErr != nil {
		fmt.Printf("  Error fetching library: %v\n", fetchErr)
		os.RemoveAll(downloadTmpDir)
		return fetchErr
	}

	// 创建下载哈希文件
	if err := saveHash(downloadTmpDir, p.Config, false); err != nil {
		os.RemoveAll(downloadTmpDir)
		return fmt.Errorf("failed to create download hash file: %v", err)
	}

	// 如果原下载目录存在，先删除
	if err := os.RemoveAll(downloadDir); err != nil {
		os.RemoveAll(downloadTmpDir)
		return fmt.Errorf("failed to remove old download directory: %v", err)
	}

	// 原子性地将临时目录移动为正式目录
	if err := os.Rename(downloadTmpDir, downloadDir); err != nil {
		os.RemoveAll(downloadTmpDir)
		return fmt.Errorf("failed to move temporary directory: %v", err)
	}

	fmt.Printf("  下载完成，源码位于: %s\n", downloadDir)
	return nil
}

// fetchFromGit clones a git repository
func fetchFromGit(gitConfig *GitSpec, downloadDir string) error {
	// 清理下载目录，确保没有残留文件
	files, err := os.ReadDir(downloadDir)
	if err != nil {
		return err
	}
	for _, file := range files {
		filePath := filepath.Join(downloadDir, file.Name())
		if file.IsDir() {
			if err := os.RemoveAll(filePath); err != nil {
				return err
			}
		} else {
			if err := os.Remove(filePath); err != nil {
				return err
			}
		}
	}

	// 克隆仓库
	cmd := exec.Command("git", "clone", gitConfig.Repo, ".")
	cmd.Dir = downloadDir

	// 执行克隆命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone error: %s - %v", output, err)
	}

	// 如果指定了特定的引用（分支、标签或提交），执行checkout
	if gitConfig.Ref != "" {
		cmd = exec.Command("git", "checkout", gitConfig.Ref)
		cmd.Dir = downloadDir
		output, err = cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("git checkout error: %s - %v", output, err)
		}
	}

	fmt.Printf("  Git仓库克隆成功: %s 到 %s\n", gitConfig.Repo, downloadDir)
	if gitConfig.Ref != "" {
		fmt.Printf("  Checked out: %s\n", gitConfig.Ref)
	}

	return nil
}

// fetchFromFiles downloads files specified in the configuration
func fetchFromFiles(files []FileSpec, downloadDir string, clean bool) error {
	fmt.Printf("  开始下载文件: %#v\n", files)
	// 确保下载目录存在
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		return fmt.Errorf("failed to create download directory: %v", err)
	}

	if clean {
		// 清理下载目录，确保没有残留文件
		dirEntries, err := os.ReadDir(downloadDir)
		if err != nil {
			return err
		}
		for _, entry := range dirEntries {
			filePath := filepath.Join(downloadDir, entry.Name())
			if entry.IsDir() {
				if err := os.RemoveAll(filePath); err != nil {
					return err
				}
			} else {
				if err := os.Remove(filePath); err != nil {
					return err
				}
			}
		}
	}

	// 下载并处理每个文件
	for i, file := range files {
		if file.URL == "" {
			continue
		}

		// 从URL中提取文件名
		parts := strings.Split(file.URL, "/")
		filename := parts[len(parts)-1]
		tmpFilePath := filepath.Join(downloadDir, filename+".download") // 临时文件
		finalFilePath := filepath.Join(downloadDir, filename)           // 最终文件位置

		fmt.Printf("  正在下载 (%d/%d): %s\n", i+1, len(files), file.URL)

		// 下载文件
		resp, err := http.Get(file.URL)
		if err != nil {
			return fmt.Errorf("download failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("bad status: %s", resp.Status)
		}

		// 创建临时文件
		out, err := os.Create(tmpFilePath)
		if err != nil {
			return fmt.Errorf("failed to create file: %v", err)
		}

		// 写入文件内容
		_, err = io.Copy(out, resp.Body)
		out.Close() // 确保文件关闭，即使出错
		if err != nil {
			os.Remove(tmpFilePath) // 清理临时文件
			return fmt.Errorf("failed to write file: %v", err)
		}

		// 原子性地重命名文件
		if err := os.Rename(tmpFilePath, finalFilePath); err != nil {
			os.Remove(tmpFilePath) // 清理临时文件
			return fmt.Errorf("failed to finalize file: %v", err)
		}

		fmt.Printf("  文件下载完成: %s\n", finalFilePath)

		if file.NoExtract {
			continue
		}

		var extractDir string
		if file.ExtractDir != "" {
			extractDir = filepath.Join(downloadDir, file.ExtractDir)
		} else {
			extractDir = downloadDir
		}

		fmt.Printf("  Extracting to: %s\n", extractDir)
		if err := os.MkdirAll(extractDir, 0755); err != nil {
			return fmt.Errorf("failed to create extract directory: %v", err)
		}

		// 解压文件（如果需要）
		if strings.HasSuffix(filename, ".tar.gz") || strings.HasSuffix(filename, ".tgz") {
			fmt.Printf("  正在解压: %s\n", filename)
			cmd := exec.Command("tar", "-xzf", filename, "-C", extractDir)
			cmd.Dir = downloadDir
			if output, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("extraction error: %s - %v", output, err)
			}
			// 解压成功后删除原始压缩文件
			os.Remove(finalFilePath)
		} else if strings.HasSuffix(filename, ".zip") {
			fmt.Printf("  正在解压: %s\n", filename)
			cmd := exec.Command("unzip", filename, "-d", extractDir)
			cmd.Dir = downloadDir
			if output, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("extraction error: %s - %v", output, err)
			}
			// 解压成功后删除原始压缩文件
			os.Remove(finalFilePath)
		}
	}

	return nil
}
