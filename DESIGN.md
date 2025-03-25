# LLGo clibs 构建系统设计文档

## 1. 介绍

LLGo clibs 构建系统用于管理、获取和构建 C/C++ 依赖库，简化跨平台应用开发。该系统使用 YAML 配置文件来描述库的来源和构建方式，支持从 Git 仓库或文件下载源码，并能够智能地检测何时需要重新获取和构建库。

## 2. 配置文件格式

### 2.1 `pkg.yaml` 配置文件规范

配置文件应位于库的根目录，命名为 `pkg.yaml`。格式如下：

```yaml
version: "1.0.0" # 库版本号 (必需)

git: # Git 仓库信息 (与 Files 二选一)
  repo: "https://github.com/example/repo.git" # Git 仓库地址
  ref: "v1.0.0" # 分支、标签或提交 ID

files: # 从文件下载 (与 Git 二选一)
  - url: "https://example.com/file.tar.gz" # 文件 URL
    filename: "file.tar.gz" # 保存的文件名
    extract: true # 是否解压文件

build: # 构建配置 (必需)
  command: "mkdir -p out && cd out && cmake .. && make" # 构建命令
```

### 2.2 字段说明

- **version**: 指定库的版本，用于跟踪和管理。
- **git**: 从 Git 仓库下载源码。
  - **repo**: Git 仓库的 URL。
  - **ref**: 要检出的分支、标签或提交 ID。
- **files**: 从文件列表下载源码。
  - **url**: 文件的下载 URL。
  - **filename**: 下载后保存的文件名。
  - **extract**: 是否解压缩下载的文件（支持 .zip, .tar.gz 等格式）。
- **build**: 构建配置。
  - **command**: 构建命令，目前是 bash shell。
    - 支持的环境变量:
      - `$CLIBS_BUILD_DIR`: 指向编译产物的目标目录
      - `$CLIBS_PACKAGE_DIR`: 指向模块的本地路径

## 3. 目录结构

每个库维护以下目录结构：

```
library_root/
  ├── pkg.yaml                 # 配置文件
  ├── _download/               # 下载的源码目录
  │   ├── _download_hash       # 下载源码的配置哈希
  │   └── ...                  # 源码文件
  ├── _build/                  # 编译后的产物
  │   └── platform_arch/       # 平台和架构特定的构建
  │       ├── _build_hash      # 构建状态文件（JSON）
  │       └── ...              # 编译后的文件（库、头文件等）
  └── _prebuilt/               # 预构建缓存目录
      └── platform_arch/       # 平台和架构特定的预构建
          ├── _build_hash      # 预构建状态文件（JSON）
          └── ...              # 预构建的文件
```

## 4. 处理流程

### 4.1 整体流程

1. 解析 `pkg.yaml` 配置文件
2. 检查预构建缓存、构建状态和下载状态
3. 如果需要，获取源码到 `_download` 目录
4. 如果需要，执行构建命令，生成产物到 `_build/{platform}_{arch}` 目录

### 4.2 详细处理流程

构建系统按照以下流程处理库的获取和构建：

1. **解析配置文件**：

   - 读取和解析 `pkg.yaml` 配置文件
   - 生成配置文件的哈希值，用于后续状态比较

2. **检查预构建缓存**：

   - 检查 `_prebuilt/{platform_arch}/_build_hash` 是否存在且内容与配置文件哈希一致
   - 如果一致，直接使用预构建缓存，跳过获取和构建步骤，返回 `_prebuilt/{platform_arch}` 目录

3. **检查构建状态**：

   - 如果预构建检查失败，检查 `_build/{platform_arch}/_build_hash` 是否存在且内容与配置文件哈希一致
   - 如果一致，直接使用已构建的版本，跳过获取和构建步骤，返回 `_build/{platform_arch}` 目录

4. **检查下载源码状态**：

   - 如果构建状态检查失败，检查 `_download` 目录是否存在且非空
   - 检查 `_download/_download_hash` 是否存在且内容与配置文件哈希一致
   - 如果检查失败，执行下载步骤

5. **下载源码**：

   - 创建临时下载目录 `_download_tmp`
   - 根据配置选择 Git 或文件下载方式获取源码
   - 下载完成后，将配置哈希写入 `_download_tmp/_download_hash`
   - 原子性地将整个临时目录移动到 `_download`，确保完整性

6. **构建库**：
   - 执行构建命令，生成产物到 `_build/{platform_arch}` 目录
   - 构建成功后，将配置哈希写入 `_build/{platform_arch}/_build_hash`

### 4.3 状态检测逻辑

系统使用以下机制跟踪获取和构建状态：

- **预构建缓存状态**: 通过 `_prebuilt/{platform_arch}/_build_hash` 文件的哈希值检查
- **构建状态**: 通过 `_build/{platform_arch}/_build_hash` 文件的哈希值检查
- **源码获取状态**: 通过 `_download/_download_hash` 文件的哈希值检查和目录内容检查

#### 重新获取条件

以下任一条件满足时，系统会重新获取源码：

1. `_download` 目录不存在或为空
2. `_download/_download_hash` 文件不存在或内容与配置哈希不一致

#### 重新构建条件

以下任一条件满足时，系统会重新构建：

1. `_prebuilt/{platform_arch}/_build_hash` 和 `_build/{platform_arch}/_build_hash` 文件不存在或内容与配置哈希不一致
2. 源码发生变化（通过 `_download/_download_hash` 检查）

### 4.4 原子性保障

为确保获取和构建过程的原子性：

1. 源码获取使用临时目录 `_download_tmp`
2. 下载完成后，将整个临时目录原子性地移动到 `_download`，同时写入 `_download_hash`
3. 只有在构建成功后才会更新 `_build/{platform_arch}/_build_hash` 状态文件
4. 预构建缓存中的 `_build_hash` 与正常构建使用相同格式，确保兼容性

### 4.5 状态文件格式

`_build_hash` 和 `_download_hash` 文件使用 JSON 格式存储：

```json
{
  "ConfigHash": "md5sum_of_config_file",
  "Timestamp": 1616414033
}
```

- **ConfigHash**: 配置文件的 MD5 哈希值
- **Timestamp**: Unix 时间戳

## 5. 命令执行环境

构建命令在库源码目录（`_download`）中执行，并设置以下环境变量：

- `CLIBS_PACKAGE_DIR`: 模块的本地路径（e.g. `$HOME/go/pkg/mod/github.com/goplus/clibs@v1.0.0`）
- `CLIBS_BUILD_TARGET`: 构建目标的三元组 (e.g. `wasm32-unknown-wasi`)
- `CLIBS_BUILD_CFLAGS`: 构建目标的 CFLAGS
- `CLIBS_BUILD_LDFLAGS`: 构建目标的 LDFLAGS
- `CLIBS_BUILD_DIR`: 平台和架构特定的构建输出目录（e.g. `$CLIBS_PACKAGE_DIR/_build/$CLIBS_BUILD_TARGET`）

`LLGo` 使用库时会设置 `CLIBS_LIB_DIR`, `CLIBS_INCLUDE_DIR` 来解析 `LLGoPackage` 和 `LLGoFiles`。

- `CLIBS_LIB_DIR`: 根据库的构建情况，指向 `_prebuilt/$CLIBS_BUILD_TARGET` 或 `_build/$CLIBS_BUILD_TARGET`

## 6. 用法示例

### 示例 1: 使用 Git 源码

```yaml
version: "8.2.8"
git:
  repo: "https://github.com/ivmai/bdwgc.git"
  ref: "v8.2.8"
build:
  command: "mkdir -p out && cd out && cmake -Dbuild_tests=OFF -DCMAKE_INSTALL_PREFIX=$CLIBS_BUILD_DIR .. && cmake --build . && cmake --install ."
```

### 示例 2: 使用文件下载

```yaml
version: "1.2.11"
files:
  - url: "https://zlib.net/zlib-1.2.11.tar.gz"
    filename: "zlib-1.2.11.tar.gz"
    extract: true
build:
  command: "mkdir -p out && cd out && cmake -DCMAKE_INSTALL_PREFIX=$CLIBS_BUILD_DIR .. && make && make install"
```
