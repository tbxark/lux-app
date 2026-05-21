APP_NAME := lux-app
PKG := .
BIN_DIR := bin
DIST_DIR := dist
GO_CACHE_DIR := $(CURDIR)/.go-build-cache
ZIG_CACHE_DIR := $(CURDIR)/.zig-cache
ZIG_GLOBAL_CACHE_DIR := $(CURDIR)/.zig-global-cache
ZIG ?= zig
DARWIN_SDK ?= $(shell xcrun --sdk macosx --show-sdk-path 2>/dev/null)
DARWIN_MIN_VERSION ?= 10.15
LINUX_SYSROOT ?=
LINUX_PKG_CONFIG_PATH ?=
LINUX_ZIG_FLAGS := $(if $(LINUX_SYSROOT),--sysroot $(LINUX_SYSROOT),)
HOST_GOOS := $(shell go env GOOS 2>/dev/null)
HOST_CGO_LDFLAGS :=
ifeq ($(HOST_GOOS),darwin)
HOST_CGO_LDFLAGS := -Wl,-no_warn_duplicate_libraries
endif
ifneq ($(DARWIN_SDK),)
DARWIN_CFLAGS := -isysroot $(DARWIN_SDK) -I$(DARWIN_SDK)/usr/include -L$(DARWIN_SDK)/usr/lib -F$(DARWIN_SDK)/System/Library/Frameworks -mmacosx-version-min=$(DARWIN_MIN_VERSION) -Wno-nullability-completeness -Wno-unguarded-availability-new -Wno-deprecated-declarations -Wno-availability -Wno-unknown-warning-option
else
DARWIN_CFLAGS := -mmacosx-version-min=$(DARWIN_MIN_VERSION)
endif
ifeq ($(HOST_GOOS),darwin)
DARWIN_NATIVE_CFLAGS := $(DARWIN_CFLAGS) -Wno-unused-command-line-argument
DARWIN_AMD64_CC := clang -arch x86_64 $(DARWIN_NATIVE_CFLAGS)
DARWIN_AMD64_CXX := clang++ -arch x86_64 $(DARWIN_NATIVE_CFLAGS)
DARWIN_ARM64_CC := clang -arch arm64 $(DARWIN_NATIVE_CFLAGS)
DARWIN_ARM64_CXX := clang++ -arch arm64 $(DARWIN_NATIVE_CFLAGS)
DARWIN_CGO_LDFLAGS := $(HOST_CGO_LDFLAGS)
else
DARWIN_AMD64_CC := $(ZIG) cc -target x86_64-macos $(DARWIN_CFLAGS)
DARWIN_AMD64_CXX := $(ZIG) c++ -target x86_64-macos $(DARWIN_CFLAGS)
DARWIN_ARM64_CC := $(ZIG) cc -target aarch64-macos $(DARWIN_CFLAGS)
DARWIN_ARM64_CXX := $(ZIG) c++ -target aarch64-macos $(DARWIN_CFLAGS)
DARWIN_CGO_LDFLAGS :=
endif
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
GO_BUILD := go build -trimpath
LDFLAGS := -s -w -X main.version=$(VERSION)
WINDOWS_LDFLAGS := $(LDFLAGS) -H=windowsgui -extldflags=-Wl,--subsystem,windows

.PHONY: all build test tidy clean dist dist-linux dist-all \
	darwin-amd64 darwin-arm64 linux-amd64 linux-arm64 windows-amd64 windows-arm64

all: build

build:
	mkdir -p $(BIN_DIR) $(GO_CACHE_DIR)
	GOCACHE="$(GO_CACHE_DIR)" CGO_LDFLAGS="$(HOST_CGO_LDFLAGS)" CGO_ENABLED=1 $(GO_BUILD) -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(APP_NAME) $(PKG)

test:
	mkdir -p $(GO_CACHE_DIR)
	GOCACHE="$(GO_CACHE_DIR)" CGO_LDFLAGS="$(HOST_CGO_LDFLAGS)" go test ./...

tidy:
	go mod tidy

clean:
	rm -rf $(BIN_DIR) $(DIST_DIR) $(GO_CACHE_DIR) $(ZIG_CACHE_DIR) $(ZIG_GLOBAL_CACHE_DIR)

dist: darwin-amd64 darwin-arm64 windows-amd64 windows-arm64

dist-linux: linux-amd64 linux-arm64

dist-all: dist dist-linux

darwin-amd64:
	mkdir -p $(DIST_DIR) $(GO_CACHE_DIR) $(ZIG_CACHE_DIR) $(ZIG_GLOBAL_CACHE_DIR)
	GOCACHE="$(GO_CACHE_DIR)" CGO_LDFLAGS="$(DARWIN_CGO_LDFLAGS)" GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 \
		CC="$(DARWIN_AMD64_CC)" \
		CXX="$(DARWIN_AMD64_CXX)" \
		$(GO_BUILD) -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(APP_NAME)-darwin-amd64 $(PKG)

darwin-arm64:
	mkdir -p $(DIST_DIR) $(GO_CACHE_DIR) $(ZIG_CACHE_DIR) $(ZIG_GLOBAL_CACHE_DIR)
	GOCACHE="$(GO_CACHE_DIR)" CGO_LDFLAGS="$(DARWIN_CGO_LDFLAGS)" GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 \
		CC="$(DARWIN_ARM64_CC)" \
		CXX="$(DARWIN_ARM64_CXX)" \
		$(GO_BUILD) -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(APP_NAME)-darwin-arm64 $(PKG)

linux-amd64:
	mkdir -p $(DIST_DIR) $(GO_CACHE_DIR) $(ZIG_CACHE_DIR) $(ZIG_GLOBAL_CACHE_DIR)
	GOCACHE="$(GO_CACHE_DIR)" GOOS=linux GOARCH=amd64 CGO_ENABLED=1 \
		ZIG_LOCAL_CACHE_DIR="$(ZIG_CACHE_DIR)" \
		ZIG_GLOBAL_CACHE_DIR="$(ZIG_GLOBAL_CACHE_DIR)" \
		PKG_CONFIG_SYSROOT_DIR="$(LINUX_SYSROOT)" \
		PKG_CONFIG_PATH="$(LINUX_PKG_CONFIG_PATH)" \
		CC="$(ZIG) cc -target x86_64-linux-gnu $(LINUX_ZIG_FLAGS)" \
		CXX="$(ZIG) c++ -target x86_64-linux-gnu $(LINUX_ZIG_FLAGS)" \
		$(GO_BUILD) -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(APP_NAME)-linux-amd64 $(PKG)

linux-arm64:
	mkdir -p $(DIST_DIR) $(GO_CACHE_DIR) $(ZIG_CACHE_DIR) $(ZIG_GLOBAL_CACHE_DIR)
	GOCACHE="$(GO_CACHE_DIR)" GOOS=linux GOARCH=arm64 CGO_ENABLED=1 \
		ZIG_LOCAL_CACHE_DIR="$(ZIG_CACHE_DIR)" \
		ZIG_GLOBAL_CACHE_DIR="$(ZIG_GLOBAL_CACHE_DIR)" \
		PKG_CONFIG_SYSROOT_DIR="$(LINUX_SYSROOT)" \
		PKG_CONFIG_PATH="$(LINUX_PKG_CONFIG_PATH)" \
		CC="$(ZIG) cc -target aarch64-linux-gnu $(LINUX_ZIG_FLAGS)" \
		CXX="$(ZIG) c++ -target aarch64-linux-gnu $(LINUX_ZIG_FLAGS)" \
		$(GO_BUILD) -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(APP_NAME)-linux-arm64 $(PKG)

windows-amd64:
	mkdir -p $(DIST_DIR) $(GO_CACHE_DIR) $(ZIG_CACHE_DIR) $(ZIG_GLOBAL_CACHE_DIR)
	GOCACHE="$(GO_CACHE_DIR)" GOOS=windows GOARCH=amd64 CGO_ENABLED=1 \
		ZIG_LOCAL_CACHE_DIR="$(ZIG_CACHE_DIR)" \
		ZIG_GLOBAL_CACHE_DIR="$(ZIG_GLOBAL_CACHE_DIR)" \
		CC="$(ZIG) cc -target x86_64-windows-gnu" \
		CXX="$(ZIG) c++ -target x86_64-windows-gnu" \
		$(GO_BUILD) -ldflags "$(WINDOWS_LDFLAGS)" -o $(DIST_DIR)/$(APP_NAME)-windows-amd64.exe $(PKG)

windows-arm64:
	mkdir -p $(DIST_DIR) $(GO_CACHE_DIR) $(ZIG_CACHE_DIR) $(ZIG_GLOBAL_CACHE_DIR)
	GOCACHE="$(GO_CACHE_DIR)" GOOS=windows GOARCH=arm64 CGO_ENABLED=1 \
		ZIG_LOCAL_CACHE_DIR="$(ZIG_CACHE_DIR)" \
		ZIG_GLOBAL_CACHE_DIR="$(ZIG_GLOBAL_CACHE_DIR)" \
		CC="$(ZIG) cc -target aarch64-windows-gnu" \
		CXX="$(ZIG) c++ -target aarch64-windows-gnu" \
		$(GO_BUILD) -ldflags "$(WINDOWS_LDFLAGS)" -o $(DIST_DIR)/$(APP_NAME)-windows-arm64.exe $(PKG)
