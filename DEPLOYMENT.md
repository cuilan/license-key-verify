# 部署和发布指南

本文档描述了如何部署和发布 License Key Verify Tool 项目。

## 📋 前置要求

- Go 1.23+
- Git
- Docker（可选，用于容器化部署）
- GitHub 账户（用于 CI/CD）

## 🚀 自动化发布流程

### 1. 创建发布版本

```bash
# 1. 确保代码已提交并推送到 main 分支
git add .
git commit -m "准备发布 v1.0.0"
git push origin main

# 2. 创建并推送版本标签
git tag v1.0.0
git push origin v1.0.0
```

### 2. GitHub Actions 自动化

推送标签后，GitHub Actions 将自动执行以下操作：

1. **运行测试** - 确保代码质量
2. **构建所有平台** - 生成多平台二进制文件
3. **创建发布包** - 打包为 tar.gz 格式
4. **生成变更日志** - 基于 Git 提交历史
5. **创建 GitHub Release** - 自动发布到 GitHub
6. **构建 Docker 镜像** - 推送到 GitHub Container Registry

### 3. 发布产物

发布完成后，用户可以从以下位置获取：

- **GitHub Releases**: https://github.com/cuilan/license-key-verify/releases
- **Docker 镜像**: `ghcr.io/cuilan/license-key-verify:v1.0.0`

## 🐳 Docker 部署

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

### 本地构建镜像

```bash
# 构建镜像
docker build -t license-key-verify .

# 运行容器
docker run --rm license-key-verify --help
```

## 📦 手动构建和发布

### 本地构建

```bash
# 构建当前平台
make build

# 构建所有平台
make build-all

# 创建发布包
make release
```

### 安装到系统

```bash
# 安装到 /usr/local/bin
sudo make install

# 卸载
sudo make uninstall
```

## 🔧 配置管理

### 环境变量

项目支持以下环境变量：

- `LKCTL_KEYS_DIR` - 默认密钥目录
- `LKCTL_LOG_LEVEL` - 日志级别（debug, info, warn, error）

### 配置文件

可以在用户主目录创建 `.lkctl.yaml` 配置文件：

```yaml
keys_dir: ~/.lkctl/keys
log_level: info
default_duration: 365
```

## 📊 监控和日志

### 构建状态

通过以下徽章监控构建状态：

- [![CI](https://github.com/cuilan/license-key-verify/actions/workflows/ci.yml/badge.svg)](https://github.com/cuilan/license-key-verify/actions/workflows/ci.yml)
- [![Release](https://github.com/cuilan/license-key-verify/actions/workflows/release.yml/badge.svg)](https://github.com/cuilan/license-key-verify/actions/workflows/release.yml)
- [![Docker](https://github.com/cuilan/license-key-verify/actions/workflows/docker.yml/badge.svg)](https://github.com/cuilan/license-key-verify/actions/workflows/docker.yml)

### 日志收集

在生产环境中，建议配置日志收集：

```bash
# 使用 journald（systemd 环境）
lkctl --log-format json | systemd-cat -t lkctl

# 使用文件日志
lkctl --log-file /var/log/lkctl.log
```

## 🔒 安全考虑

### 密钥管理

1. **私钥保护**：私钥应存储在安全位置，设置适当的文件权限
2. **密钥轮换**：定期更换密钥对
3. **备份策略**：安全备份密钥文件

```bash
# 设置密钥文件权限
chmod 600 keys/private.pem
chmod 644 keys/public.pem
chmod 600 keys/aes.key
```

### 网络安全

1. **HTTPS 传输**：通过 HTTPS 分发许可证文件
2. **访问控制**：限制密钥文件的访问权限
3. **审计日志**：记录许可证生成和验证操作

## 🚨 故障排除

### 常见问题

1. **构建失败**
   ```bash
   # 检查 Go 版本
   go version
   
   # 清理并重新构建
   make clean
   make build
   ```

2. **测试失败**
   ```bash
   # 运行详细测试
   go test -v ./...
   
   # 运行特定测试
   go test -v ./pkg/crypto
   ```

3. **Docker 构建失败**
   ```bash
   # 检查 Docker 版本
   docker version
   
   # 清理 Docker 缓存
   docker system prune -f
   ```

### 日志分析

```bash
# 查看 GitHub Actions 日志
# 访问：https://github.com/cuilan/license-key-verify/actions

# 查看本地构建日志
make build 2>&1 | tee build.log
```

## 📈 性能优化

### 构建优化

1. **并行构建**：使用 `make -j$(nproc)` 并行构建
2. **缓存利用**：GitHub Actions 自动缓存 Go 模块
3. **镜像优化**：使用多阶段构建减小镜像大小

### 运行时优化

1. **内存使用**：监控内存使用情况
2. **CPU 使用**：优化加密算法性能
3. **磁盘 I/O**：优化文件读写操作

## 📝 版本管理

### 语义化版本

项目遵循 [语义化版本](https://semver.org/lang/zh-CN/) 规范：

- `MAJOR.MINOR.PATCH`
- 例如：`v1.2.3`

### 发布周期

- **主版本**：重大功能更新或破坏性变更
- **次版本**：新功能添加，向后兼容
- **补丁版本**：Bug 修复和小改进

### 分支策略

- `main` - 稳定版本分支
- `develop` - 开发分支
- `feature/*` - 功能分支
- `hotfix/*` - 热修复分支

## 🤝 贡献指南

参与项目开发请遵循以下流程：

1. Fork 项目
2. 创建功能分支
3. 提交变更
4. 创建 Pull Request
5. 等待代码审查

详细信息请参考 [CONTRIBUTING.md](CONTRIBUTING.md)。 