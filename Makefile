# 定义变量，使配置更容易管理
BINARY_NAME=FinDocOCR
OUTPUT_DIR=./release
LDFLAGS=-w -s
ICON_FILE=icon.ico
VERSION_INFO=resource.syso

# 定义不同平台的输出文件名
WINDOWS_OUTPUT=$(OUTPUT_DIR)/$(BINARY_NAME)_windows_amd64.exe
LINUX_OUTPUT=$(OUTPUT_DIR)/$(BINARY_NAME)_linux_amd64
DARWIN_OUTPUT=$(OUTPUT_DIR)/$(BINARY_NAME)_darwin_amd64

# 默认目标：构建所有平台
.PHONY: all
all: windows linux darwin

# 创建输出目录
$(OUTPUT_DIR):
	mkdir -p $(OUTPUT_DIR)

# 生成 Windows 资源文件（包含图标）
$(VERSION_INFO): $(ICON_FILE)
	goversioninfo -icon=$(ICON_FILE)

# Windows 构建目标
.PHONY: windows
windows: $(OUTPUT_DIR) $(VERSION_INFO)
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(WINDOWS_OUTPUT) -p 4
	@if command -v upx >/dev/null 2>&1; then \
		echo "Compressing Windows binary with UPX..."; \
		upx --best --lzma $(WINDOWS_OUTPUT); \
	fi

# Linux 构建目标
.PHONY: linux
linux: $(OUTPUT_DIR)
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(LINUX_OUTPUT) -p 4
	@if command -v upx >/dev/null 2>&1; then \
		echo "Compressing Linux binary with UPX..."; \
		upx --best --lzma $(LINUX_OUTPUT); \
	fi

# macOS 构建目标
.PHONY: darwin
darwin: $(OUTPUT_DIR)
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(DARWIN_OUTPUT) -p 4
	@if command -v upx >/dev/null 2>&1; then \
		echo "Compressing macOS binary with UPX..."; \
		upx --best --lzma $(DARWIN_OUTPUT); \
	fi

# 清理构建产物
.PHONY: clean
clean:
	rm -rf $(OUTPUT_DIR)
	rm -f $(VERSION_INFO)

# 帮助信息
.PHONY: help
help:
	@echo "可用的构建目标:"
	@echo "  make all     - 构建所有平台的版本"
	@echo "  make windows - 仅构建 Windows 版本"
	@echo "  make linux   - 仅构建 Linux 版本"
	@echo "  make darwin  - 仅构建 macOS 版本"
	@echo "  make clean   - 清理构建产物"