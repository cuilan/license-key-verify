# License Key Verify Tool

[![CI](https://github.com/cuilan/license-key-verify/actions/workflows/ci.yml/badge.svg)](https://github.com/cuilan/license-key-verify/actions/workflows/ci.yml)
[![Release](https://github.com/cuilan/license-key-verify/actions/workflows/release.yml/badge.svg)](https://github.com/cuilan/license-key-verify/actions/workflows/release.yml)
[![Docker](https://github.com/cuilan/license-key-verify/actions/workflows/docker.yml/badge.svg)](https://github.com/cuilan/license-key-verify/actions/workflows/docker.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

面向软件授权的许可证签发与验证工具：**一行代码集成、部署机器零密钥文件、纯 RSA 签名**。

**中文** | [English](README_EN.md)

## 核心特性

- **一行代码集成**：`import _ ".../autocheck"` 或 `license.MustCheck()`
- **零密钥文件**：公钥编译进二进制（`lkctl embed`），部署机器只有二进制 + license 文件
- **纯 RSA 签名**：RSA-2048 + SHA256，无对称加密，攻击面更小
- **机器绑定**：单指纹（平台 UUID + MAC 经 SHA-256），签发时锁定目标机器
- **离线验证**：无需网络，签名、有效期、指纹全部本地校验
- **纯标准库**：零第三方依赖

## 快速开始

### 1. 构建

```bash
make build   # 产出 bin/lkctl
```

### 2. 生成密钥对（你本地）

```bash
lkctl keygen --out keys/
# keys/private.pem  签发私钥，务必保密
# keys/public.pem   验证公钥，用于内嵌
```

### 3. 客户机器获取指纹

```bash
# 方式一：你的应用内置了 autocheck，直接带参数运行
./your-app --fingerprint
# 输出：090499363c0b2202d7f025cd7dc2dedce65421c0a25eaccb55b890e64bf73625

# 方式二：在客户机器上运行 lkctl
lkctl fingerprint
```

### 4. 签发许可证（你本地）

```bash
lkctl sign \
  --private-key keys/private.pem \
  --fingerprint 090499363c0b... \
  --customer "某某公司" \
  --product "leosap" \
  --features "full" \
  --duration 365 \
  --out license.lic
```

### 5. 交付与验证

把 `license.lic` 发给客户，放到应用默认路径（或用 `LICENSE_FILE` 环境变量指定），应用启动时自动校验。也可手动验证：

```bash
lkctl verify license.lic          # 退出码 0=有效 1=无效
lkctl verify license.lic --json   # JSON 格式结果
lkctl info license.lic        # 查看许可证内容
```

## 应用集成

### 模式 A：import 即校验（零代码改动）

```go
package main

import _ "github.com/cuilan/license-key-verify/autocheck"

func main() {
    // init() 中已完成校验，失败直接退出
}
```

### 模式 B：显式一行调用

```go
package main

import "github.com/cuilan/license-key-verify/pkg/license"

func main() {
    license.MustCheck() // 失败输出原因并退出码 1

    // 或需要自定义处理时：
    // if err := license.Check(); err != nil { ... }
}
```

许可证路径解析顺序：

1. 环境变量 `LICENSE_FILE`
2. 可执行文件同目录下的 `license.lic`
3. Unix: `/etc/<程序名>/license.lic`

### 公钥内嵌（生产必需）

仓库内置的 `pkg/license/key.go` 是**演示密钥**（配对私钥在 `testdata/private.pem`，仅用于测试），生产环境必须替换为你自己的：

```bash
lkctl embed --public-key keys/public.pem --out pkg/license/key.go
```

生成的文件将公钥经 XOR 混淆后内嵌（防止 `strings` 直接搜出），随你的应用一起编译。部署机器上没有任何 `.pem` / `.key` 文件。

> 直接以 Go module 方式引用本项目、不便替换 `key.go` 时，改用运行时注入：
> `license.UsePublicKeyPEM(publicKeyPEM)` 后再调用 `license.MustCheck()`。

## 命令参考

### lkctl（签发方）

```bash
lkctl keygen --out keys/                      生成 RSA 密钥对
lkctl embed --public-key <pem> --out <file>   公钥生成为 Go 源文件
lkctl fingerprint                             打印本机指纹
lkctl sign --private-key <pem> --fingerprint <fp> --out license.lic
        [--duration 365] [--customer ...] [--product ...] [--features a,b]
lkctl verify <license-file> [--public-key <pem>] [--json] [--quiet]
                                              验证许可证（退出码 0=有效 1=无效 2=参数错误）
lkctl info <license-file> [--public-key <pem>]     查看许可证内容
```

## 许可证文件格式

明文 JSON + RSA 签名（签名覆盖除 `signature` 外的全部字段）：

```json
{
  "license_id": "2366ee07-1c29-979a-6106-e253ed47301c",
  "customer": "某某公司",
  "product": "leosap",
  "fingerprint": "090499363c0b2202d7f025cd7dc2dedce65421c0a25eaccb55b890e64bf73625",
  "issued_at": "2026-10-27T00:00:00Z",
  "expires_at": "2027-10-27T00:00:00Z",
  "features": ["full"],
  "signature": "<base64 RSA-SHA256 签名>"
}
```

## SDK

```go
// 签发（签发方）
lic, err := license.Sign(&license.SignOptions{
    Customer:    "某某公司",
    Product:     "leosap",
    Fingerprint: fp,
    Features:    []string{"full"},
    Duration:    365 * 24 * time.Hour,
}, privateKeyPEM)
err = lic.Save("license.lic")

// 验证（应用侧）
result := license.VerifyFile("license.lic")
if result.Valid {
    fmt.Println("剩余天数:", result.ExpiresIn/(24*3600))
}
```

## 机器指纹

指纹 = `SHA-256(版本 | 平台稳定标识 | 排序后的 MAC 列表)`

- 平台稳定标识：Linux 优先 `product_uuid`（无 DMI 时用 `machine-id`），macOS 用 `IOPlatformUUID`，Windows 用系统 UUID
- 不使用随机值兜底，同一机器重复计算结果一致
- 重装系统 / 更换硬件 / 变更网卡会导致指纹变化，需要重新签发

## 安全说明

1. **私钥只在你手里**：签发私钥绝不下发，泄露即可伪造许可证
2. **演示密钥仅用于测试**：仓库自带的 `testdata/private.pem` 与内嵌公钥配对，公开可见，生产前必须执行 `lkctl keygen` + `lkctl embed` 替换
3. **防篡改靠签名**：许可证为明文 JSON，任何人可读，但无法在无私钥的情况下修改内容
4. **无防回拨/吊销**：离线验证不校验系统时间回拨，也不支持吊销列表，需要时请自行扩展

## Docker

```bash
docker build -t license-key-verify .
docker run --rm license-key-verify --help
docker run --rm -v $(pwd):/workspace license-key-verify fingerprint
```

## 开发

```bash
make test          # 单元测试
make test-e2e      # 端到端测试
make build-all     # 全平台构建
make demo          # 生成示例许可证
make verify-demo   # 验证示例
```

## 常见问题

**Q: 客户复制 license 文件到别的机器能用吗？**
不能。指纹不匹配，验证失败。

**Q: 忘记内嵌新公钥会怎样？**
应用仍携带演示公钥，你的许可证验证失败（`signature verification failed`）。执行 `lkctl embed` 后重新编译即可。

**Q: 支持 Windows / macOS / Linux 吗？**
支持。指纹采集与构建均覆盖三大平台。

## 许可证

MIT，详见 LICENSE 文件。
