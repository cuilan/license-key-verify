## 2.0.0 (2026-10-29)

破坏性重构：一行集成、零密钥文件、纯 RSA 签名。

### 变更

- **去掉 AES**：许可证格式改为明文 JSON + RSA-SHA256 签名（签名覆盖除 `signature` 外全部字段），删除对称加密层
- **公钥内嵌**：`lkctl embed` 将公钥生成 Go 源文件（XOR 混淆），部署机器零密钥文件
- **零侵入集成**：新增 `autocheck` 包（import 即校验，支持 `--fingerprint`）与 `license.Check()` / `license.MustCheck()` 一行调用
- **统一机器指纹**：SHA-256(平台稳定标识 + 排序 MAC)，替代旧的 MAC/UUID/CPUID 三元组；修复 Linux 下随机 UUID 回退与 CPUID 无区分度问题
- **命令重写**：`lkctl` 改为 `keygen / embed / sign / verify / info / fingerprint`；`verify` 内置公钥、支持 `--json` / `--quiet`，退出码 0/1/2；**删除 `lkverify`**（功能与 `lkctl verify` + autocheck 完全重叠）
- **不再提供演示密钥**：删除仓库内置的 `testdata/` 密钥对与 `pkg/license/key.go`（demo 内嵌公钥）；未内嵌公钥时 `verify` 需显式 `--public-key`
- **通用工具定位**：`sign` 只需 `--private-key` 即可签发（不再要求内嵌公钥，移除 `--force`）；`verify` 无内嵌公钥时提示改用 `--public-key`；内嵌公钥与签名私钥不配对时 `sign` 仅警告不阻断；`version` / `help` 改为纯子命令（去掉 `--` 形式）
- **Makefile 修复**：无 tag 时 VERSION 回退 `0.0.0`；`clean` 不再删除用户密钥（`keys/`、`*.lic`）；删除 `demo` / `verify-demo` 目标
- **默认路径**：`LICENSE_FILE` → 可执行文件同目录 `license.lic` → `/etc/<程序名>/license.lic`

### 迁移

- 旧 `lkctl gen / keys / get` 命令删除，改用 `lkctl sign / keygen / fingerprint`
- 旧 `pkg/license` API（`NewGenerator` / `NewVerifierFromFiles` 等）删除，改用 `license.Sign` / `license.Verify` / `license.Check`
- 旧版加密许可证文件（含 `data/algorithm` 字段）无法被新版本读取

## 变更内容

- fix(ci):fix decode_changelog (8d5e009)
