# License Key Verify Tool

[![CI](https://github.com/cuilan/license-key-verify/actions/workflows/ci.yml/badge.svg)](https://github.com/cuilan/license-key-verify/actions/workflows/ci.yml)
[![Release](https://github.com/cuilan/license-key-verify/actions/workflows/release.yml/badge.svg)](https://github.com/cuilan/license-key-verify/actions/workflows/release.yml)
[![Docker](https://github.com/cuilan/license-key-verify/actions/workflows/docker.yml/badge.svg)](https://github.com/cuilan/license-key-verify/actions/workflows/docker.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/cuilan/license-key-verify)](https://goreportcard.com/report/github.com/cuilan/license-key-verify)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

一个功能完整的许可证密钥生成、验证工具，支持机器绑定和数字签名。

**中文** | [English](README_EN.md)

## 功能特性

- ✅ **机器信息获取**: 支持获取MAC地址、系统UUID、CPU ID
- ✅ **许可证生成**: 生成加密的许可证文件，支持机器绑定
- ✅ **许可证验证**: 验证许可证的有效性和机器匹配性
- ✅ **数字签名**: 使用RSA+AES混合加密保证安全性
- ✅ **跨平台支持**: 支持Windows、macOS、Linux
- ✅ **命令行工具**: 提供易用的命令行界面
- ✅ **SDK集成**: 可作为库集成到其他项目中

## 快速开始

### 1. 构建项目

```bash
# 构建所有二进制文件
make build

# 或者构建特定平台
make build-all
```

### 2. 获取机器信息

```bash
# 获取MAC地址
lkctl get mac

# 获取系统UUID
lkctl get uuid

# 获取CPU ID
lkctl get cpuid

# 获取所有机器信息
lkctl get all
```

### 3. 生成许可证

```bash
# 生成许可证。如果密钥文件不存在，会自动生成并保存到 --keys-dir 指定的目录（默认为 keys/）
lkctl gen \
  --keys-dir ./mykeys \
  --mac "00:11:22:33:44:55" \
  --uuid "12345678-1234-1234-1234-123456789012" \
  --cpuid "abcdef1234567890" \
  --customer "示例客户" \
  --product "示例产品" \
  --duration 365 \
  license.lic

# 使用现有密钥生成许可证
lkctl gen \
  --private-key ./mykeys/private.pem \
  --aes-key ./mykeys/aes.key \
  --customer "另一个客户" \
  license2.lic
```

### 4. 验证许可证

```bash
# 使用lkctl验证
lkctl verify license.lic

# 使用lkverify验证
lkverify license.lic

# JSON格式输出
lkverify license.lic --json
```

## 命令行工具使用说明

### lkctl 工具

`lkctl` 是主要的命令行工具，提供许可证管理的完整功能。

#### 获取机器信息

```bash
lkctl get mac     # 获取MAC地址
lkctl get uuid    # 获取系统UUID
lkctl get cpuid   # 获取CPU ID
lkctl get all     # 获取所有机器信息
```

#### 生成密钥对

```bash
lkctl keys --output <目录>
```
> **注意**: `lkctl gen` 命令在未提供密钥时也会自动生成密钥。此 `keys` 命令用于仅需要生成密钥文件的场景。

#### 生成许可证

```bash
lkctl gen [选项] <输出文件>

选项:
  --mac <mac>              指定MAC地址
  --uuid <uuid>            指定系统UUID
  --cpuid <cpuid>          指定CPU ID
  --duration <天数>        有效期（天）
  --customer <客户名>      客户名称
  --product <产品名>       产品名称
  --version <版本>         产品版本
  --features <功能列表>    功能列表（逗号分隔）
  --max-users <数量>       最大用户数
  --keys-dir <目录>        新密钥的保存目录 (默认: keys)
  --private-key <文件>     用于签名的私钥文件路径。如果未提供，则生成新的。
  --aes-key <文件>         用于加密的AES密钥文件路径。如果未提供，则生成新的。
```

#### 验证许可证

```bash
lkctl verify <许可证文件>   # 验证许可证
lkctl info <许可证文件>     # 查看许可证信息
```

### lkverify 工具

`lkverify` 是专门的验证工具，适合集成到其他程序中。

```bash
lkverify <许可证文件> [选项]

选项:
  --keys-dir <目录>     指定密钥文件目录（默认: keys）
  --public-key <文件>   指定公钥文件路径 (会覆盖 --keys-dir)
  --aes-key <文件>      指定AES密钥文件路径 (会覆盖 --keys-dir)
  --json               以JSON格式输出结果
  --quiet              安静模式，只输出退出码

退出码:
  0  许可证有效
  1  许可证无效或其他错误
  2  参数错误
```

## 在其他项目中使用

### 下载依赖库

```bash
go get github.com/cuilan/license-key-verify
```

### 或者使用本地模块（开发时）

```bash
# 在你的项目中添加本地依赖
go mod edit -replace license-key-verify=../path/to/license-key-verify
go mod tidy
```

### 在Go项目中集成使用

#### 1. 初始化你的Go项目

```bash
# 创建新项目目录
mkdir my-licensed-app
cd my-licensed-app

# 初始化Go模块
go mod init my-licensed-app

# 下载license-key-verify库
go get github.com/cuilan/license-key-verify
```

#### 2. 作为Go库使用

```go
package main

import (
    "fmt"
    "time"
    
    "github.com/cuilan/license-key-verify/pkg/license"
    "github.com/cuilan/license-key-verify/pkg/machine"
)

func main() {
    // 1. 生成许可证
    generator, err := license.NewGenerator()
    if err != nil {
        panic(err)
    }
    
    // 获取机器信息
    machineInfo, err := machine.GetAllInfo()
    if err != nil {
        panic(err)
    }
    
    options := &license.GenerateOptions{
        ProductName:  "我的产品",
        CustomerName: "客户名称",
        MAC:          machineInfo.MAC,
        UUID:         machineInfo.UUID,
        CPUID:        machineInfo.CPUID,
        Duration:     30 * 24 * time.Hour, // 30天
        Features:     []string{"feature1", "feature2"},
        MaxUsers:     10,
    }
    
    lic, err := generator.Generate(options)
    if err != nil {
        panic(err)
    }
    
    // 保存许可证
    err = generator.SaveToFile(lic, "license.lic")
    if err != nil {
        panic(err)
    }
    
    // 2. 验证许可证
    verifier, err := license.NewVerifierFromFiles("keys/public.pem", "keys/aes.key")
    if err != nil {
        panic(err)
    }
    
    result, err := verifier.VerifyFile("license.lic")
    if err != nil {
        panic(err)
    }
    
    if result.Valid {
        fmt.Println("许可证验证通过")
        fmt.Printf("剩余天数: %d\n", result.ExpiresIn/(24*3600))
    } else {
        fmt.Printf("许可证验证失败: %s\n", result.Error)
    }
}
```

#### 3. 简化的集成示例

创建一个简单的许可证检查函数：

```go
// license_check.go
package main

import (
    "fmt"
    "os"
    
    "github.com/cuilan/license-key-verify/pkg/license"
)

func checkLicense(licenseFile string) bool {
    // 从默认keys目录加载验证器
    verifier, err := license.NewVerifierFromFiles("keys/public.pem", "keys/aes.key")
    if err != nil {
        fmt.Printf("无法加载密钥文件: %v\n", err)
        return false
    }
    
    // 验证许可证文件
    result, err := verifier.VerifyFile(licenseFile)
    if err != nil {
        fmt.Printf("验证过程出错: %v\n", err)
        return false
    }
    
    if result.Valid {
        fmt.Println("✓ 许可证验证通过")
        if result.ExpiresIn > 0 {
            days := result.ExpiresIn / (24 * 3600)
            fmt.Printf("许可证还有 %d 天到期\n", days)
        }
        return true
    } else {
        fmt.Printf("✗ 许可证验证失败: %s\n", result.Error)
        return false
    }
}

func main() {
    if len(os.Args) < 2 {
        fmt.Println("用法: go run main.go <许可证文件>")
        os.Exit(1)
    }
    
    licenseFile := os.Args[1]
    
    if checkLicense(licenseFile) {
        fmt.Println("应用程序已授权，正常启动...")
        // 你的应用程序逻辑
    } else {
        fmt.Println("未授权访问，程序退出")
        os.Exit(1)
    }
}
```

#### 4. 编译和运行

```bash
# 构建你的应用
go build -o my-app

# 运行（需要先准备好密钥文件和许可证文件）
./my-app license.lic
```

## 构建和部署

### 本地构建

```bash
# 构建当前平台
make build

# 构建所有平台
make build-all

# 运行测试
make test

# 生成示例
make demo
```

### GitHub Actions 手动构建

项目使用手动触发的 GitHub Actions 工作流：

1. 访问 [GitHub Actions](https://github.com/cuilan/license-key-verify/actions)
2. 选择 "CI" 工作流
3. 点击 "Run workflow" 手动触发构建
4. 可选择是否运行测试和构建所有平台

详细说明请参考：[手动触发指南](docs/manual-trigger-guide.md)

### 安装到系统

```bash
# 安装到 /usr/local/bin
sudo make install

# 卸载
sudo make uninstall
```

## 安全特性

1. **混合加密**: 使用AES-256-GCM对称加密 + RSA-2048非对称签名
2. **机器绑定**: 通过MAC地址、UUID、CPU ID进行机器绑定
3. **防篡改**: 数字签名确保许可证文件不被篡改
4. **时间验证**: 支持许可证有效期验证
5. **功能控制**: 支持按功能模块授权

## 许可证文件格式

许可证文件采用JSON格式，包含以下字段：

```json
{
  "data": "加密的许可证数据（Base64编码）",
  "signature": "数字签名（Base64编码）",
  "algorithm": "加密算法标识",
  "version": "文件格式版本"
}
```

## Docker 支持

### 使用预构建镜像

```bash
# 拉取最新镜像
docker pull ghcr.io/cuilan/license-key-verify:latest

# 运行容器
docker run --rm ghcr.io/cuilan/license-key-verify:latest --help

# 挂载本地目录进行许可证操作
docker run --rm -v $(pwd):/workspace \
  ghcr.io/cuilan/license-key-verify:latest \
  get mac
```

### 本地构建

```bash
# 构建镜像
docker build -t license-key-verify .

# 运行容器
docker run --rm license-key-verify --help
```

## 开发和贡献

### 开发环境要求

- Go 1.23+
- Make工具

### 代码规范

```bash
# 格式化代码
make fmt

# 代码检查
make lint

# 运行测试
make test
```

## 常见问题

### Q: 如何在新机器上验证许可证？

A: 许可证文件绑定了机器信息，只能在对应的机器上验证通过。如需在新机器上使用，需要重新生成许可证。

### Q: 许可证文件可以复制到其他机器使用吗？

A: 不可以。许可证文件包含了机器绑定信息，在不匹配的机器上验证会失败。

### Q: 如何备份和恢复密钥？

A: 使用 `lkctl gen` 命令时，可以通过 `--keys-dir` 参数指定密钥的生成目录（默认为 `keys/`）。建议备份整个目录。私钥用于生成许可证，公钥和AES密钥用于验证。请务必安全保管您的私钥。

### Q: 支持离线验证吗？

A: 支持。验证过程完全离线进行，不需要网络连接。

## 许可证

本项目采用MIT许可证，详见LICENSE文件。

## 致谢

- 仅使用 Go 标准库构建
- 使用 RSA 和 AES 加密算法
- 受现代软件许可实践启发

## 支持

- 📖 [文档](docs/)
- 🐛 [问题跟踪](https://github.com/cuilan/license-key-verify/issues)
- 💬 [讨论区](https://github.com/cuilan/license-key-verify/discussions)
- 📧 联系：[创建 Issue](https://github.com/cuilan/license-key-verify/issues/new)
