APP_NAME := lux-app
APP_ID := github.com.tbxark.lux-app
PKG := .
BIN_DIR := bin
DIST_DIR := dist
GO_CACHE_DIR := $(CURDIR)/.go-build-cache
ZIG_CACHE_DIR := $(CURDIR)/.zig-cache
ZIG_GLOBAL_CACHE_DIR := $(CURDIR)/.zig-global-cache
COMMA := ,

FYNE ?= fyne
ZIG ?= zig
ZIG_WRAP := $(CURDIR)/scripts/zig-wrap
GOOS ?= $(shell go env GOOS 2>/dev/null)
GOARCH ?= $(shell go env GOARCH 2>/dev/null)
GO_BUILD := go build -trimpath
LDFLAGS := -s -w
WINDOWS_LDFLAGS := $(LDFLAGS) -H=windowsgui
HOST_CGO_LDFLAGS := $(if $(filter darwin,$(shell go env GOOS 2>/dev/null)),-Wl$(COMMA)-no_warn_duplicate_libraries,)
DARWIN_SDK ?= $(shell xcrun --sdk macosx --show-sdk-path 2>/dev/null)
DARWIN_ZIG_FLAGS := $(if $(DARWIN_SDK),-isysroot $(DARWIN_SDK) -I$(DARWIN_SDK)/usr/include -L$(DARWIN_SDK)/usr/lib -F$(DARWIN_SDK)/System/Library/Frameworks,) -Wno-nullability-completeness -Wno-error=nullability-completeness -Wno-unguarded-availability-new -Wno-deprecated-declarations -Wno-availability -Wno-unknown-warning-option

ZIG_TARGET_darwin_amd64 := x86_64-macos
ZIG_TARGET_darwin_arm64 := aarch64-macos
ZIG_TARGET_windows_amd64 := x86_64-windows-gnu
ZIG_TARGET_windows_arm64 := aarch64-windows-gnu

.PHONY: all build test tidy clean package package-native dist dist-linux dist-all \
	darwin-amd64 darwin-arm64 windows-amd64 windows-arm64 \
    format

all: build

build:
	mkdir -p $(BIN_DIR) $(GO_CACHE_DIR)
	GOCACHE="$(GO_CACHE_DIR)" CGO_ENABLED=1 CGO_LDFLAGS="$(HOST_CGO_LDFLAGS)" $(GO_BUILD) -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(APP_NAME) $(PKG)

test:
	mkdir -p $(GO_CACHE_DIR)
	GOCACHE="$(GO_CACHE_DIR)" CGO_LDFLAGS="$(HOST_CGO_LDFLAGS)" go test ./...

tidy:
	go mod tidy

clean:
	rm -rf $(BIN_DIR) $(DIST_DIR) $(GO_CACHE_DIR) $(ZIG_CACHE_DIR) $(ZIG_GLOBAL_CACHE_DIR) $(APP_NAME).app $(APP_NAME).exe $(APP_NAME).tar.*

package-native:
	$(MAKE) package GOOS=$(shell go env GOOS) GOARCH=$(shell go env GOARCH) CGO_LDFLAGS='$(HOST_CGO_LDFLAGS)'

package:
	mkdir -p $(BIN_DIR) $(DIST_DIR) $(GO_CACHE_DIR) $(ZIG_CACHE_DIR) $(ZIG_GLOBAL_CACHE_DIR)
	GOCACHE="$(GO_CACHE_DIR)" ZIG_LOCAL_CACHE_DIR="$(ZIG_CACHE_DIR)" ZIG_GLOBAL_CACHE_DIR="$(ZIG_GLOBAL_CACHE_DIR)" CGO_ENABLED=1 GOOS=$(GOOS) GOARCH=$(GOARCH) $(if $(CC),CC="$(CC)") $(if $(CXX),CXX="$(CXX)") CGO_LDFLAGS="$(CGO_LDFLAGS)" $(GO_BUILD) -ldflags "$(if $(filter windows,$(GOOS)),$(WINDOWS_LDFLAGS),$(LDFLAGS))" -o "$(BIN_DIR)/$(APP_NAME)-$(GOOS)-$(GOARCH)$(EXE_EXT)" $(PKG)
	GOCACHE="$(GO_CACHE_DIR)" ZIG_LOCAL_CACHE_DIR="$(ZIG_CACHE_DIR)" ZIG_GLOBAL_CACHE_DIR="$(ZIG_GLOBAL_CACHE_DIR)" GOOS=$(GOOS) GOARCH=$(GOARCH) $(if $(CC),CC="$(CC)") $(if $(CXX),CXX="$(CXX)") CGO_LDFLAGS="$(CGO_LDFLAGS)" $(FYNE) package --target "$(GOOS)" --executable "$(BIN_DIR)/$(APP_NAME)-$(GOOS)-$(GOARCH)$(EXE_EXT)" --name "$(APP_NAME)" --app-id "$(APP_ID)"
	@case "$(GOOS)" in \
		darwin) rm -rf "$(DIST_DIR)/$(APP_NAME)-$(GOOS)-$(GOARCH).app"; mv "$(APP_NAME).app" "$(DIST_DIR)/$(APP_NAME)-$(GOOS)-$(GOARCH).app" ;; \
		windows) mv "$(BIN_DIR)/$(APP_NAME)-$(GOOS)-$(GOARCH).exe" "$(DIST_DIR)/$(APP_NAME)-$(GOOS)-$(GOARCH).exe" ;; \
		linux) artifact=$$(ls "$(APP_NAME)".tar.*); mv "$$artifact" "$(DIST_DIR)/$(APP_NAME)-$(GOOS)-$(GOARCH).$${artifact#$(APP_NAME).}" ;; \
	esac

dist: darwin-amd64 darwin-arm64 windows-amd64 windows-arm64

dist-linux:
	@echo "Linux desktop packages need Linux X11/OpenGL development files; run 'make package-native' on Linux."

dist-all: dist dist-linux

darwin-amd64:
	$(MAKE) package GOOS=darwin GOARCH=amd64 CC='$(ZIG) cc -target $(ZIG_TARGET_darwin_amd64) $(DARWIN_ZIG_FLAGS)' CXX='$(ZIG) c++ -target $(ZIG_TARGET_darwin_amd64) $(DARWIN_ZIG_FLAGS)'

darwin-arm64:
	$(MAKE) package GOOS=darwin GOARCH=arm64 CC='$(ZIG) cc -target $(ZIG_TARGET_darwin_arm64) $(DARWIN_ZIG_FLAGS)' CXX='$(ZIG) c++ -target $(ZIG_TARGET_darwin_arm64) $(DARWIN_ZIG_FLAGS)'

windows-amd64:
	$(MAKE) package GOOS=windows GOARCH=amd64 EXE_EXT=.exe CC='bash $(ZIG_WRAP) cc $(ZIG_TARGET_windows_amd64)' CXX='bash $(ZIG_WRAP) c++ $(ZIG_TARGET_windows_amd64)'

windows-arm64:
	$(MAKE) package GOOS=windows GOARCH=arm64 EXE_EXT=.exe CC='bash $(ZIG_WRAP) cc $(ZIG_TARGET_windows_arm64)' CXX='bash $(ZIG_WRAP) c++ $(ZIG_TARGET_windows_arm64)'

format:
	go fix ./...
	go fmt ./...
	go vet ./...
	go get ./...
	go test ./...
	go mod tidy
	golangci-lint fmt --no-config --enable gofmt,goimports
	golangci-lint run --no-config --fix
	nilaway -include-pkgs="$(MODULE)" ./...
