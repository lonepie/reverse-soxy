# Makefile for reverse-soxy

APP_NAME := reverse-soxy
BUILD_DIR := build

.PHONY: all clean linux_amd64 linux_arm64 darwin_amd64 darwin_arm64 windows_amd64 windows_arm64 detect_os_arch

all: detect_os_arch

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

detect_os_arch:
	@echo "Detecting OS and architecture..."
	@OS_TYPE=$$(uname -s | tr '[:upper:]' '[:lower:]') && \
	ARCH_TYPE=$$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/') && \
	case "$$OS_TYPE" in \
		linux*) \
			case "$$ARCH_TYPE" in \
				amd64) $(MAKE) linux_amd64 ;; \
				arm64) $(MAKE) linux_arm64 ;; \
				*) echo "Unsupported architecture: $$ARCH_TYPE"; exit 1 ;; \
			esac ;; \
		darwin*) \
			case "$$ARCH_TYPE" in \
				amd64) $(MAKE) darwin_amd64 ;; \
				arm64) $(MAKE) darwin_arm64 ;; \
				*) echo "Unsupported architecture: $$ARCH_TYPE"; exit 1 ;; \
			esac ;; \
		*) echo "Unsupported OS: $$OS_TYPE"; exit 1 ;; \
	esac

linux_amd64: | $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 ./cmd/reverse-soxy

linux_arm64: | $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 ./cmd/reverse-soxy

darwin_amd64: | $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 ./cmd/reverse-soxy

darwin_arm64: | $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 ./cmd/reverse-soxy

windows_amd64: | $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe ./cmd/reverse-soxy

windows_arm64: | $(BUILD_DIR)
	GOOS=windows GOARCH=arm64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-windows-arm64.exe ./cmd/reverse-soxy

clean:
	rm -rf $(BUILD_DIR)