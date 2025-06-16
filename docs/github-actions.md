# GitHub Actions 工作流

本项目使用 GitHub Actions 实现自动化的 CI/CD 流程。

## 🔄 工作流概览

### 1. CI 工作流 (`.github/workflows/ci.yml`)

**触发条件：**
- 手动触发（workflow_dispatch）

**手动触发参数：**
- `run_tests`: 是否运行测试（默认：true）
- `build_all_platforms`: 是否构建所有平台（默认：true）

**执行步骤：**
1. **测试阶段**
   - 代码格式检查
   - 运行单元测试
   - 生成测试覆盖率报告
   - 上传覆盖率到 Codecov

2. **构建阶段**
   - 构建所有平台的二进制文件
   - 上传构建产物

### 2. Release 工作流 (`.github/workflows/release.yml`)

**触发条件：**
- 推送版本标签（格式：`v*`）

**执行步骤：**
1. 运行完整测试
2. 构建发布包
3. 生成变更日志
4. 创建 GitHub Release
5. 上传发布产物

### 3. Docker 工作流 (`.github/workflows/docker.yml`)

**触发条件：**
- 手动触发（workflow_dispatch）
- 推送版本标签

**手动触发参数：**
- `push_to_registry`: 是否推送到容器注册表（默认：false）
- `platforms`: 构建平台（默认：linux/amd64,linux/arm64）

**执行步骤：**
1. 构建多架构 Docker 镜像
2. 推送到 GitHub Container Registry

## 🚀 使用方法

### 手动触发 CI 工作流

1. 访问 GitHub Actions 页面：`https://github.com/cuilan/license-key-verify/actions`
2. 选择 "CI" 工作流
3. 点击 "Run workflow" 按钮
4. 选择参数：
   - 是否运行测试
   - 是否构建所有平台
5. 点击 "Run workflow" 执行

### 手动触发 Docker 构建

1. 访问 GitHub Actions 页面
2. 选择 "Docker" 工作流
3. 点击 "Run workflow" 按钮
4. 选择参数：
   - 是否推送到注册表
   - 构建平台
5. 点击 "Run workflow" 执行

### 创建发布版本

```bash
# 1. 确保代码已提交
git add .
git commit -m "准备发布 v1.0.0"
git push origin main

# 2. 创建并推送标签（自动触发发布）
git tag v1.0.0
git push origin v1.0.0
```

### 自动化流程

推送标签后，GitHub Actions 将：
1. ✅ 运行所有测试
2. 🔨 构建多平台二进制文件
3. 📦 创建发布包
4. 📝 生成变更日志
5. 🎉 创建 GitHub Release
6. 🐳 构建并推送 Docker 镜像

## 📊 状态监控

通过以下徽章监控构建状态：

```markdown
[![CI](https://github.com/cuilan/license-key-verify/actions/workflows/ci.yml/badge.svg)](https://github.com/cuilan/license-key-verify/actions/workflows/ci.yml)
[![Release](https://github.com/cuilan/license-key-verify/actions/workflows/release.yml/badge.svg)](https://github.com/cuilan/license-key-verify/actions/workflows/release.yml)
[![Docker](https://github.com/cuilan/license-key-verify/actions/workflows/docker.yml/badge.svg)](https://github.com/cuilan/license-key-verify/actions/workflows/docker.yml)
```

## 🔧 配置说明

### 缓存优化

由于项目只使用 Go 标准库，缓存配置使用 `go.mod` 文件作为缓存键：

```yaml
key: ${{ runner.os }}-go-${{ hashFiles('**/go.mod') }}
```

### 权限设置

工作流需要以下权限：
- `contents: write` - 创建 Release
- `packages: write` - 推送 Docker 镜像

### 环境要求

- Go 1.23+
- Ubuntu Latest
- Docker Buildx

## 🐛 故障排除

### 常见问题

1. **构建失败**
   - 检查 Go 版本兼容性
   - 确认代码格式正确
   - 验证测试通过

2. **发布失败**
   - 确认标签格式正确（`v*`）
   - 检查权限设置
   - 验证变更日志生成

3. **Docker 构建失败**
   - 检查 Dockerfile 语法
   - 确认多架构支持
   - 验证镜像推送权限

### 调试方法

1. **查看工作流日志**
   ```
   https://github.com/cuilan/license-key-verify/actions
   ```

2. **本地测试**
   ```bash
   # 测试构建
   make build-all
   
   # 测试 Docker
   docker build -t test .
   ```

3. **手动触发**
   - 可以在 GitHub Actions 页面手动触发工作流

## 📝 最佳实践

1. **分支策略**
   - `main` - 稳定版本
   - `develop` - 开发版本
   - `feature/*` - 功能分支

2. **标签规范**
   - 使用语义化版本：`v1.0.0`
   - 预发布版本：`v1.0.0-beta.1`

3. **提交信息**
   - 使用清晰的提交信息
   - 遵循约定式提交规范

4. **测试覆盖**
   - 保持高测试覆盖率
   - 添加集成测试

## 🔄 工作流更新

如需修改工作流：

1. 编辑 `.github/workflows/` 下的 YAML 文件
2. 提交并推送更改
3. 在下次触发时生效

注意：工作流更改只在推送到默认分支后生效。 