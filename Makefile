# License Key Verify Tool Makefile

# 项目信息
PROJECT_NAME = license-key-verify
VERSION = $(shell git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//' || echo "0.0.0")
BUILD_TIME = $(shell date +%Y-%m-%d_%H:%M:%S)
GIT_COMMIT = $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go 参数
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get
GOMOD = $(GOCMD) mod

# 构建参数
LDFLAGS = -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# 输出目录
DIST_DIR = dist
BIN_DIR = bin

# 目标平台
PLATFORMS = darwin/amd64,darwin/arm64,linux/amd64,linux/arm64,windows/amd64

.PHONY: all build build-all clean test test-e2e deps help install uninstall release

# 默认目标
all: clean deps test build test-e2e

# 构建当前平台的二进制文件
build:
	@echo "构建 lkctl..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/lkctl ./cmd/lkctl
	@echo "构建 lkverify..."
	$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/lkverify ./cmd/lkverify
	@echo "构建完成!"

# 构建所有平台的二进制文件
build-all:
	@echo "构建所有平台的二进制文件..."
	@mkdir -p $(DIST_DIR)
	@for platform in $(shell echo $(PLATFORMS) | tr ',' ' '); do \
		os=$$(echo $$platform | cut -d'/' -f1); \
		arch=$$(echo $$platform | cut -d'/' -f2); \
		echo "构建 $$os/$$arch..."; \
		if [ "$$os" = "windows" ]; then \
			GOOS=$$os GOARCH=$$arch $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/lkctl_$${os}_$${arch}.exe ./cmd/lkctl; \
			GOOS=$$os GOARCH=$$arch $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/lkverify_$${os}_$${arch}.exe ./cmd/lkverify; \
		else \
			GOOS=$$os GOARCH=$$arch $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/lkctl_$${os}_$${arch} ./cmd/lkctl; \
			GOOS=$$os GOARCH=$$arch $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/lkverify_$${os}_$${arch} ./cmd/lkverify; \
		fi; \
	done
	@echo "所有平台构建完成!"

# 安装到系统路径
install: build
	@echo "安装到系统路径..."
	@if [ "$(shell id -u)" -ne 0 ]; then \
		echo "错误: 需要 root 权限来安装"; \
		exit 1; \
	fi
	@cp $(BIN_DIR)/lkctl /usr/local/bin/
	@cp $(BIN_DIR)/lkverify /usr/local/bin/
	@echo "安装完成!"

# 从系统路径卸载
uninstall:
	@echo "从系统路径卸载..."
	@if [ "$(shell id -u)" -ne 0 ]; then \
		echo "错误: 需要 root 权限来卸载"; \
		exit 1; \
	fi
	@rm -f /usr/local/bin/lkctl
	@rm -f /usr/local/bin/lkverify
	@echo "卸载完成!"

# 清理构建文件
clean:
	@echo "清理构建文件..."
	@$(GOCLEAN)
	@rm -rf $(BIN_DIR)
	@rm -rf $(DIST_DIR)
	@rm -rf keys/
	@rm -f *.lic
	@rm -rf demo/
	@rm -rf tests/test_run/
	@rm -f coverage.out coverage.html
	@echo "清理完成!"

# 运行单元测试
test:
	@echo "运行单元测试..."
	@$(GOTEST) -v ./...

# 运行端到端测试
test-e2e: build
	@echo "运行端到端测试..."
	@chmod +x tests/e2e_test.sh
	@./tests/e2e_test.sh

# 运行测试并生成覆盖率报告
test-coverage:
	@echo "运行测试并生成覆盖率报告..."
	@$(GOTEST) -v -coverprofile=coverage.out ./...
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告已生成: coverage.html"

# 下载依赖
deps:
	@echo "下载依赖..."
	@$(GOMOD) download
	@$(GOMOD) tidy

# 格式化代码
fmt:
	@echo "格式化代码..."
	@$(GOCMD) fmt ./...

# 代码检查
lint:
	@echo "代码检查..."
	@golangci-lint run ./...

# 生成示例密钥和许可证
demo: build
	@echo "生成示例..."
	@mkdir -p demo
	@# lkctl gen 会自动在 demo/ 目录中创建密钥文件
	@./$(BIN_DIR)/lkctl gen \
		--keys-dir demo \
		--mac $(shell ./$(BIN_DIR)/lkctl get mac) \
		--uuid $(shell ./$(BIN_DIR)/lkctl get uuid) \
		--cpuid $(shell ./$(BIN_DIR)/lkctl get cpuid) \
		--customer "示例客户" \
		--product "示例产品" \
		--version "1.0.0" \
		--duration 30 \
		--features "feature1,feature2" \
		--max-users 10 \
		demo/license.lic
	@echo "示例文件已生成到 demo/ 目录"

# 验证示例许可证
verify-demo: demo
	@echo "验证示例许可证..."
	@./$(BIN_DIR)/lkverify demo/license.lic --keys-dir demo

# 创建发布包
release: build-all
	@echo "创建发布包..."
	@mkdir -p $(DIST_DIR)/release
	@for platform in $(shell echo $(PLATFORMS) | tr ',' ' '); do \
		os=$$(echo $$platform | cut -d'/' -f1); \
		arch=$$(echo $$platform | cut -d'/' -f2); \
		release_name="$(PROJECT_NAME)_$(VERSION)_$${os}_$${arch}"; \
		echo "打包 $$release_name..."; \
		mkdir -p $(DIST_DIR)/release/$$release_name; \
		if [ "$$os" = "windows" ]; then \
			if [ -f "$(DIST_DIR)/lkctl_$${os}_$${arch}.exe" ]; then \
				cp $(DIST_DIR)/lkctl_$${os}_$${arch}.exe $(DIST_DIR)/release/$$release_name/lkctl.exe; \
			else \
				echo "警告: $(DIST_DIR)/lkctl_$${os}_$${arch}.exe 不存在"; \
			fi; \
			if [ -f "$(DIST_DIR)/lkverify_$${os}_$${arch}.exe" ]; then \
				cp $(DIST_DIR)/lkverify_$${os}_$${arch}.exe $(DIST_DIR)/release/$$release_name/lkverify.exe; \
			else \
				echo "警告: $(DIST_DIR)/lkverify_$${os}_$${arch}.exe 不存在"; \
			fi; \
		else \
			if [ -f "$(DIST_DIR)/lkctl_$${os}_$${arch}" ]; then \
				cp $(DIST_DIR)/lkctl_$${os}_$${arch} $(DIST_DIR)/release/$$release_name/lkctl; \
				chmod +x $(DIST_DIR)/release/$$release_name/lkctl; \
			else \
				echo "警告: $(DIST_DIR)/lkctl_$${os}_$${arch} 不存在"; \
			fi; \
			if [ -f "$(DIST_DIR)/lkverify_$${os}_$${arch}" ]; then \
				cp $(DIST_DIR)/lkverify_$${os}_$${arch} $(DIST_DIR)/release/$$release_name/lkverify; \
				chmod +x $(DIST_DIR)/release/$$release_name/lkverify; \
			else \
				echo "警告: $(DIST_DIR)/lkverify_$${os}_$${arch} 不存在"; \
			fi; \
		fi; \
		if [ -f "README.md" ]; then \
			cp README.md $(DIST_DIR)/release/$$release_name/; \
		fi; \
		(cd $(DIST_DIR)/release && zip -r $$release_name.zip $$release_name/ && rm -rf $$release_name); \
	done
	@echo "发布包已创建到 $(DIST_DIR)/release/ 目录"

# 显示帮助信息
help:
	@echo "License Key Verify Tool - 构建帮助"
	@echo ""
	@echo "可用目标:"
	@echo "  all          - 清理、下载依赖、测试、构建（默认）"
	@echo "  build        - 构建当前平台的二进制文件"
	@echo "  build-all    - 构建所有平台的二进制文件"
	@echo "  install      - 安装到系统路径"
	@echo "  uninstall    - 从系统路径卸载"
	@echo "  clean        - 清理构建文件"
	@echo "  test         - 运行单元测试"
	@echo "  test-e2e     - 运行端到端测试"
	@echo "  test-coverage- 运行测试并生成覆盖率报告"
	@echo "  deps         - 下载依赖"
	@echo "  fmt          - 格式化代码"
	@echo "  lint         - 代码检查"
	@echo "  demo         - 生成示例密钥和许可证"
	@echo "  verify-demo  - 验证示例许可证"
	@echo "  release      - 创建发布包"
	@echo "  help         - 显示此帮助信息"
	@echo ""
	@echo "项目信息:"
	@echo "  项目名称: $(PROJECT_NAME)"
	@echo "  版本: $(VERSION)"
	@echo "  构建时间: $(BUILD_TIME)"
	@echo "  Git提交: $(GIT_COMMIT)" 