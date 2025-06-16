# License Key Verify Tool

[![CI](https://github.com/cuilan/license-key-verify/actions/workflows/ci.yml/badge.svg)](https://github.com/cuilan/license-key-verify/actions/workflows/ci.yml)
[![Release](https://github.com/cuilan/license-key-verify/actions/workflows/release.yml/badge.svg)](https://github.com/cuilan/license-key-verify/actions/workflows/release.yml)
[![Docker](https://github.com/cuilan/license-key-verify/actions/workflows/docker.yml/badge.svg)](https://github.com/cuilan/license-key-verify/actions/workflows/docker.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/cuilan/license-key-verify)](https://goreportcard.com/report/github.com/cuilan/license-key-verify)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

一个功能完整的许可证密钥生成、验证工具，支持机器绑定和数字签名。

## 功能特性

- ✅ **机器信息获取**: 支持获取MAC地址、系统UUID、CPU ID
- ✅ **许可证生成**: 生成加密的许可证文件，支持机器绑定
- ✅ **许可证验证**: 验证许可证的有效性和机器匹配性
- ✅ **数字签名**: 使用RSA+AES混合加密保证安全性
- ✅ **跨平台支持**: 支持Windows、macOS、Linux
- ✅ **命令行工具**: 提供易用的命令行界面
- ✅ **SDK集成**: 可作为库集成到其他项目中

## 项目结构

```
license-key-verify/
├── cmd/
│   ├── lkctl/          # 命令行工具
│   │   └── main.go
│   └── lkverify/       # 验证程序
│       └── main.go
├── pkg/
│   ├── license/        # 许可证核心功能
│   │   ├── types.go
│   │   ├── generator.go
│   │   └── verifier.go
│   ├── machine/        # 机器信息获取
│   │   └── info.go
│   └── crypto/         # 加密解密
│       └── crypto.go
├── Makefile           # 构建脚本
├── go.mod             # Go模块定义
└── README.md          # 项目文档
```

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
./bin/lkctl get mac

# 获取系统UUID
./bin/lkctl get uuid

# 获取CPU ID
./bin/lkctl get cpuid
```

### 3. 生成许可证

```bash
# 生成密钥对
./bin/lkctl keys --output keys

# 生成许可证
./bin/lkctl gen \
  --mac "00:11:22:33:44:55" \
  --uuid "12345678-1234-1234-1234-123456789012" \
  --cpuid "abcdef1234567890" \
  --customer "示例客户" \
  --product "示例产品" \
  --duration 365 \
  license.lic
```

### 4. 验证许可证

```bash
# 使用lkctl验证
./bin/lkctl verify license.lic

# 使用lkverify验证
./bin/lkverify license.lic

# JSON格式输出
./bin/lkverify license.lic --json
```

## 命令行工具使用说明

### lkctl 工具

`lkctl` 是主要的命令行工具，提供许可证管理的完整功能。

#### 获取机器信息

```bash
lkctl get mac     # 获取MAC地址
lkctl get uuid    # 获取系统UUID
lkctl get cpuid   # 获取CPU ID
```

#### 生成密钥对

```bash
lkctl keys --output <目录>
```

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
  --json               以JSON格式输出结果
  --quiet              安静模式，只输出退出码

退出码:
  0  许可证有效
  1  许可证无效或其他错误
  2  参数错误
```

## API使用示例

### 作为Go库使用

```go
package main

import (
    "fmt"
    "time"
    
    "license-key-verify/pkg/license"
    "license-key-verify/pkg/machine"
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

### 安装到系统

```bash
# 安装到 /usr/local/bin
make install
```

### 创建发布包

```bash
# 创建所有平台的发布包
make release
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

A: 密钥文件保存在`keys/`目录下，建议备份整个目录。私钥用于生成许可证，公钥和AES密钥用于验证。

### Q: 支持离线验证吗？

A: 支持。验证过程完全离线进行，不需要网络连接。

## 许可证

本项目采用MIT许可证，详见LICENSE文件。

## 版本历史

- v1.0.0 - 初始版本
  - 支持许可证生成和验证
  - 支持机器绑定
  - 提供命令行工具和SDK
