#!/usr/bin/env bash
set -euo pipefail

RELEASE_TAG="master-700-c2df4e1"
RELEASE_URL="https://github.com/leejet/stable-diffusion.cpp/releases/download/$RELEASE_TAG"
SD_COMMIT="c2df4e1"
WIALS_VERSION="v2.12.0"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info()  { echo -e "${GREEN}[✓]${NC} $1"; }
warn()  { echo -e "${YELLOW}[!]${NC} $1"; }
err()   { echo -e "${RED}[✗]${NC} $1" >&2; }

# ---- Detect platform ----
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
SD_DIR=""
SYS_DEPS=""

case "$OS" in
  linux)
    case "$ARCH" in
      x86_64|amd64)
        SD_DIR="sd-master-${SD_COMMIT}-bin-Linux-Ubuntu-24.04-x86_64-vulkan"
        SYS_DEPS="libgtk-3-dev libwebkit2gtk-4.1-dev unzip"
        ;;
      *)
        err "Unsupported architecture: $ARCH (only x86_64 supported)"
        exit 1
        ;;
    esac
    ;;
  darwin)
    case "$ARCH" in
      arm64|aarch64)
        SD_DIR="sd-master-${SD_COMMIT}-bin-Darwin-macOS-15.7.7-arm64"
        ;;
      x86_64)
        SD_DIR="sd-master-${SD_COMMIT}-bin-Darwin-macOS-15.7.7-arm64"
        warn "Intel Mac not officially supported — trying ARM binary"
        ;;
    esac
    ;;
  mingw*|msys*|cygwin*)
    OS="windows"
    case "$ARCH" in
      x86_64|amd64)
        SD_DIR="sd-master-${SD_COMMIT}-bin-win-vulkan-x64"
        ;;
      *)
        err "Unsupported architecture: $ARCH"
        exit 1
        ;;
    esac
    ;;
  *)
    err "Unsupported OS: $OS"
    exit 1
    ;;
esac

echo "━━━ ComfyGo Setup ━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  OS:      $OS ($ARCH)"
echo "  Library: $SD_DIR"
echo ""

# ---- 1. Check prerequisites ----
PREREQ_OK=1

if ! command -v go &>/dev/null; then
  err "Go is not installed. Install Go 1.23+ from https://go.dev/dl/"
  PREREQ_OK=0
else
  GO_VER="$(go version | grep -oP 'go\K[0-9]+\.[0-9]+')"
  if [ "$(echo "$GO_VER" | cut -d. -f1)" -lt 1 ] || { [ "$(echo "$GO_VER" | cut -d. -f1)" -eq 1 ] && [ "$(echo "$GO_VER" | cut -d. -f2)" -lt 23 ]; }; then
    err "Go 1.23+ required (found $GO_VER)"
    PREREQ_OK=0
  else
    info "Go $GO_VER"
  fi
fi

if ! command -v node &>/dev/null; then
  err "Node.js is not installed. Install Node 20+ from https://nodejs.org/"
  PREREQ_OK=0
else
  NODE_VER="$(node --version | grep -oP 'v\K[0-9]+')"
  if [ "$NODE_VER" -lt 20 ]; then
    err "Node 20+ required (found $(node --version))"
    PREREQ_OK=0
  else
    info "Node $(node --version)"
  fi
fi

if [ "$PREREQ_OK" -eq 0 ]; then
  exit 1
fi

# ---- 2. Install system dependencies ----
if [ "$OS" = "linux" ] && [ -n "$SYS_DEPS" ]; then
  echo ""
  echo "  Installing system dependencies..."
  sudo apt-get update -qq
  sudo apt-get install -y -qq $SYS_DEPS
  info "System dependencies installed"
fi

if [ "$OS" = "darwin" ]; then
  # macOS has WebKit built-in; just ensure xcode tools
  if ! xcode-select -p &>/dev/null; then
    warn "Xcode Command Line Tools not found — run 'xcode-select --install'"
  fi
fi

# ---- 3. Install Wails CLI ----
echo ""
echo "  Installing Wails CLI..."
go install "github.com/wailsapp/wails/v2/cmd/wails@$WIALS_VERSION"
WIALS_BIN="$(go env GOPATH)/bin/wails"
if [ ! -f "$WIALS_BIN" ]; then
  err "Wails CLI not found at $WIALS_BIN — check GOPATH"
  exit 1
fi
info "Wails $("$WIALS_BIN" version 2>/dev/null || echo "$WIALS_VERSION")"

# ---- 4. Download prebuilt stable-diffusion.cpp ----
SDCPP_DIR="Sdcpp/$SD_DIR"
if [ -f "$SDCPP_DIR/libstable-diffusion.so" ] || [ -f "$SDCPP_DIR/libstable-diffusion.dylib" ] || [ -f "$SDCPP_DIR/libstable-diffusion.dll" ]; then
  info "Prebuilt library already exists at $SDCPP_DIR"
else
  echo ""
  echo "  Downloading stable-diffusion.cpp prebuilt binary..."
  mkdir -p "$SDCPP_DIR"
  ZIP_URL="$RELEASE_URL/$SD_DIR.zip"
  echo "    URL: $ZIP_URL"
  curl -sL "$ZIP_URL" -o /tmp/sdcpp.zip
  unzip -o -q /tmp/sdcpp.zip -d "$SDCPP_DIR"
  rm /tmp/sdcpp.zip

  # macOS ships .dylib — create a .so symlink for consistent cgo path
  if [ "$OS" = "darwin" ] && [ -f "$SDCPP_DIR/libstable-diffusion.dylib" ]; then
    ln -sf libstable-diffusion.dylib "$SDCPP_DIR/libstable-diffusion.so"
  fi

  info "Prebuilt library downloaded to $SDCPP_DIR"
fi

# ---- 5. Install frontend dependencies ----
echo ""
echo "  Installing frontend dependencies..."
npm install --prefix frontend --silent 2>/dev/null || npm install --prefix frontend
info "Frontend dependencies installed"

# ---- 6. Generate Wails bindings ----
echo ""
echo "  Generating Wails bindings..."
cd "$(dirname "$0")/.."
"$WIALS_BIN" generate module
info "Bindings generated"

# ---- Done ----
echo ""
echo "━━━ Setup complete ━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "  Next steps:"
echo "    cp your-model.safetensors ~/.comfygo/models/"
echo "    make dev"
echo ""
echo "  Or manually:"
echo "    wails dev -tags webkit2_41"
echo ""
