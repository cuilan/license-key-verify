## 2.0.1 (2026-11-02)

### 修复

- **e2e 模块路径**：`tests/e2e_test.sh` 生成的临时模块仍使用旧模块路径，导致 `go mod tidy` 转而下载远程模块、测试不再验证本地代码
- **e2e 路径转换（WSL）**：WSL 下把 `/mnt/c/...` 直接传给 Windows 版 `lkctl.exe`，文件被写到 `C:\mnt\c\...`；现统一经 `to_native_path()` 转换（Git Bash / Linux 行为不变）

### 变更

- **签发前校验指纹格式**（`license.Sign` 与 `lkctl sign`）：必须是 64 位十六进制，避免签出永远无法验证的许可证
- **生效时间容差**：`issued_at` 允许 5 分钟时钟偏差，避免签发机与客户机时钟略有差异时误报 "not yet valid"
- **新增 `Result.HasFeature(name)`**：按 features 列表做功能开关判断
- **CI 增加 push / pull_request 触发**（此前仅支持手动触发）
- **Release 增加 `SHA256SUMS.txt`**：便于校验下载产物
- **`make` 版本号**：改用 `git describe --tags --always --dirty`，未打 tag 的提交与脏工作区会显示为 `2.0.0-3-gxxxxxxx-dirty`，不再统一显示最近的 tag
- **`.gitignore` 不再忽略 `key.go`**：fork 场景下 `lkctl embed` 生成的 `pkg/license/key.go` 需要提交，否则 CI 从干净 checkout 构建时缺少内嵌公钥
- **新增 `SECURITY.md`**：威胁模型、私钥保管、密钥轮换与吊销策略

## 2.0.0 (2026-10-29)

破坏性重构：一行集成、零密钥文件、纯 RSA 签名。

### 变更

- **模块路径升级 `/v2`**：`module github.com/cuilan/license-key-verify/v2`（Go 大版本规则：v2+ tag 要求路径后缀），外部引用 `go get github.com/cuilan/license-key-verify/v2@v2.0.0`，import 路径相应加 `/v2`
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
- import 路径需加 `/v2` 前缀：`github.com/cuilan/license-key-verify/v2/pkg/license`、`.../v2/autocheck`

## 变更内容

- fix(ci):fix decode_changelog (8d5e009)
