# GitHub Actions 工作流说明

本项目包含以下 GitHub Actions 工作流：

## 🔄 CI 工作流 (ci.yml)

**触发条件：**
- 推送到 `main` 或 `develop` 分支
- 针对 `main` 分支的 Pull Request

**功能：**
- 代码格式检查
- 运行单元测试
- 生成测试覆盖率报告
- 构建所有平台的二进制文件
- 上传构建产物

## 🚀 Release 工作流 (release.yml)

**触发条件：**
- 推送版本标签（格式：`v*`，如 `v1.0.0`）

**功能：**
- 运行完整测试套件
- 构建所有平台的发布包
- 自动生成变更日志
- 创建 GitHub Release
- 上传发布产物

**使用方法：**
```bash
# 创建并推送版本标签
git tag v1.0.0
git push origin v1.0.0
```

## 🐳 Docker 工作流 (docker.yml)

**触发条件：**
- 推送到 `main` 分支
- 推送版本标签
- 针对 `main` 分支的 Pull Request

**功能：**
- 构建多架构 Docker 镜像（amd64, arm64）
- 推送到 GitHub Container Registry
- 自动标签管理

**镜像地址：**
```
ghcr.io/cuilan/license-key-verify:latest
ghcr.io/cuilan/license-key-verify:v1.0.0
```

## 📋 状态徽章

你可以在 README.md 中添加以下徽章来显示构建状态：

```markdown
[![CI](https://github.com/cuilan/license-key-verify/actions/workflows/ci.yml/badge.svg)](https://github.com/cuilan/license-key-verify/actions/workflows/ci.yml)
[![Release](https://github.com/cuilan/license-key-verify/actions/workflows/release.yml/badge.svg)](https://github.com/cuilan/license-key-verify/actions/workflows/release.yml)
[![Docker](https://github.com/cuilan/license-key-verify/actions/workflows/docker.yml/badge.svg)](https://github.com/cuilan/license-key-verify/actions/workflows/docker.yml)
```

## 🔧 配置说明

### 权限要求

工作流需要以下权限：
- `contents: write` - 用于创建 Release
- `packages: write` - 用于推送 Docker 镜像

### 密钥配置

所有工作流都使用 GitHub 自动提供的 `GITHUB_TOKEN`，无需额外配置。

### 自定义配置

如需自定义工作流，可以修改以下文件：
- `.github/workflows/ci.yml` - CI 配置
- `.github/workflows/release.yml` - 发布配置
- `.github/workflows/docker.yml` - Docker 配置

## 📝 发布流程

1. **开发阶段**：推送代码到 `develop` 分支，触发 CI 检查
2. **合并阶段**：创建 PR 到 `main` 分支，触发 CI 和 Docker 构建
3. **发布阶段**：合并 PR 后，创建版本标签触发正式发布

## 🐛 问题排查

如果工作流失败，请检查：
1. Go 版本兼容性
2. 测试是否通过
3. 代码格式是否正确
4. 依赖是否正确安装

更多详情请查看 [Actions 页面](https://github.com/cuilan/license-key-verify/actions)。 