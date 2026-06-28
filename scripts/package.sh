#!/usr/bin/env bash
set -euo pipefail

# ComfyGo packaging script
# Produces platform-specific distributable packages.

VERSION="0.1.5"
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

echo "━━━ ComfyGo Package v$VERSION ━━━━━━━━━━━━━━━━━"
echo "  OS:      $OS ($ARCH)"
echo ""

# ---- Ensure we have the app icon ----
if [ ! -f "build/appicon.png" ]; then
  echo "  Generating app icon..."
  python3 -c "
import struct, zlib

def create_png(width, height, r, g, b):
    def chunk(chunk_type, data):
        c = chunk_type + data
        return struct.pack('>I', len(data)) + c + struct.pack('>I', zlib.crc32(c) & 0xffffffff)
    header = b'\\x89PNG\\r\\n\\x1a\\n'
    ihdr = chunk(b'IHDR', struct.pack('>IIBBBBB', width, height, 8, 2, 0, 0, 0))
    raw = b''
    for y in range(height):
        raw += b'\\x00'
        for x in range(width):
            cx, cy = x - width//2, y - height//2
            d = (cx*cx + cy*cy) ** 0.5
            maxd = width//2
            if d < maxd * 0.7:
                rr, gg, bb = r, g, b
            elif d < maxd:
                t = (d - maxd*0.7) / (maxd*0.3)
                rr = int(r + (255-r)*t*0.5)
                gg = int(g + (255-g)*t*0.5)
                bb = int(b + (255-b)*t*0.5)
            else:
                rr, gg, bb = 15, 20, 40
            raw += struct.pack('BBB', min(rr,255), min(gg,255), min(bb,255))
    idat = chunk(b'IDAT', zlib.compress(raw))
    iend = chunk(b'IEND', b'')
    return header + ihdr + idat + iend

with open('build/appicon.png', 'wb') as f:
    f.write(create_png(512, 512, 124, 58, 237))
"
  echo "    Generated build/appicon.png"
fi

# ---- Ensure Sdcpp library exists ----
if [ ! -d "Sdcpp" ]; then
  echo "  stable-diffusion.cpp library not found — run 'make setup' first"
  exit 1
fi

# ---- Build the app binary ----
echo "  Building ComfyGo binary..."
case "$OS" in
  linux)
    wails build -tags webkit2_41 -o "comfygo-$VERSION-linux-$ARCH"
    ;;
  darwin)
    wails build -o "comfygo-$VERSION-darwin-$ARCH"
    ;;
  mingw*|msys*|cygwin*)
    wails build -o "comfygo-$VERSION-windows-$ARCH.exe"
    ;;
  *)
    echo "  Unknown OS, building without packaging"
    wails build -o "comfygo"
    ;;
esac
echo "    Binary: build/bin/comfygo*"

# ---- Platform-specific packages ----
case "$OS" in
  linux)
    echo ""
    echo "  Creating packages..."

    # .deb
    if command -v dpkg-deb &>/dev/null; then
      echo "    Building .deb..."
      DEB_ARCH="$ARCH"
      [ "$DEB_ARCH" = "x86_64" ] && DEB_ARCH="amd64"
      PKG_ROOT="build/pkg/comfygo_${VERSION}_${DEB_ARCH}"
      mkdir -p "$PKG_ROOT/DEBIAN"
      mkdir -p "$PKG_ROOT/usr/bin"
      mkdir -p "$PKG_ROOT/usr/share/applications"
      mkdir -p "$PKG_ROOT/usr/share/icons/hicolor/512x512/apps"
      mkdir -p "$PKG_ROOT/usr/lib/comfygo"

      cat > "$PKG_ROOT/DEBIAN/control" <<EOF
Package: comfygo
Version: $VERSION
Section: graphics
Priority: optional
Architecture: $DEB_ARCH
Maintainer: ComfyGo <dev@comfygo.app>
Description: Node-free Stable Diffusion desktop app
 A desktop application for Stable Diffusion 1.5 image generation.
 Go backend with Svelte frontend, powered by stable-diffusion.cpp.
EOF

      # Binary (pick newest by mtime)
      BIN="$(ls -t build/bin/comfygo* 2>/dev/null | head -1)"
      if [ -n "$BIN" ]; then
        cp "$BIN" "$PKG_ROOT/usr/bin/comfygo"
      fi
      # Library (find .so/.dylib/.dll in Sdcpp subdirs)
      find Sdcpp -name 'libstable-diffusion.*' -exec cp {} "$PKG_ROOT/usr/lib/comfygo/" \;
      # Desktop entry
      cp build/comfygo.desktop "$PKG_ROOT/usr/share/applications/"
      # Icon
      cp build/appicon.png "$PKG_ROOT/usr/share/icons/hicolor/512x512/apps/comfygo.png"

      dpkg-deb --build --root-owner-group "$PKG_ROOT" "build/bin/comfygo_${VERSION}_${DEB_ARCH}.deb"
      echo "    → build/bin/comfygo_${VERSION}_${DEB_ARCH}.deb"
      rm -rf "build/pkg"
    else
      echo "    dpkg-deb not found — skipping .deb"
    fi

    # .rpm
    if command -v rpmbuild &>/dev/null; then
      RPM_ARCH="$ARCH"
      echo "    Building .rpm (arch=$RPM_ARCH)..."
      mkdir -p "build/rpm/SOURCES"
      BIN="$(ls -t build/bin/comfygo* 2>/dev/null | head -1)"
      if [ -n "$BIN" ]; then
        cp "$BIN" "build/rpm/SOURCES/comfygo"
      fi

      cat > "build/rpm/comfygo.spec" <<EOF
Name: comfygo
Version: $VERSION
Release: 1
Summary: Node-free Stable Diffusion desktop app
License: MIT
URL: https://github.com/Trishvan/comfygo
Group: Graphics
BuildArch: $RPM_ARCH

%description
A desktop application for Stable Diffusion 1.5 image generation.
Go backend with Svelte frontend, powered by stable-diffusion.cpp.

%install
mkdir -p %{buildroot}%{_bindir}
mkdir -p %{buildroot}%{_datadir}/applications
mkdir -p %{buildroot}%{_datadir}/icons/hicolor/512x512/apps
mkdir -p %{buildroot}%{_libdir}/comfygo
install -m 755 %{_sourcedir}/comfygo %{buildroot}%{_bindir}/
install -m 644 build/comfygo.desktop %{buildroot}%{_datadir}/applications/
install -m 644 build/appicon.png %{buildroot}%{_datadir}/icons/hicolor/512x512/apps/comfygo.png
find Sdcpp -name 'libstable-diffusion.*' -exec cp {} %{buildroot}%{_libdir}/comfygo/ \;

%files
%{_bindir}/comfygo
%{_datadir}/applications/comfygo.desktop
%{_datadir}/icons/hicolor/512x512/apps/comfygo.png
%{_libdir}/comfygo/

%post
ldconfig

%postun
ldconfig
EOF

      rpmbuild -bb "build/rpm/comfygo.spec" --define "_topdir $(pwd)/build/rpm"
      mv "build/rpm/RPMS/$ARCH/comfygo-$VERSION-1.$ARCH.rpm" "build/bin/" 2>/dev/null || true
      echo "    → build/bin/comfygo-$VERSION-1.$RPM_ARCH.rpm"
      rm -rf "build/rpm"
    else
      echo "    rpmbuild not found — skipping .rpm"
    fi

    # .AppImage (requires appimagetool)
    if command -v appimagetool &>/dev/null; then
      echo "    Building .AppImage..."
      APPDIR="build/ComfyGo.AppDir"
      mkdir -p "$APPDIR/usr/bin"
      mkdir -p "$APPDIR/usr/share/applications"
      mkdir -p "$APPDIR/usr/share/icons/hicolor/512x512/apps"

      BIN="$(ls -t build/bin/comfygo* 2>/dev/null | head -1)"
      if [ -n "$BIN" ]; then
        cp "$BIN" "$APPDIR/usr/bin/comfygo"
      fi
      cp build/comfygo.desktop "$APPDIR/"
      cp build/appicon.png "$APPDIR/"
      cp build/appicon.png "$APPDIR/.DirIcon"

      cat > "$APPDIR/AppRun" <<'APPRUN'
#!/bin/bash
SELF=$(readlink -f "$0")
HERE=${SELF%/*}
export PATH="${HERE}/usr/bin/:${PATH}"
export LD_LIBRARY_PATH="${HERE}/usr/lib/:${LD_LIBRARY_PATH}"
exec "${HERE}/usr/bin/comfygo" "$@"
APPRUN
      chmod +x "$APPDIR/AppRun"

      appimagetool "$APPDIR" "build/bin/ComfyGo-$VERSION-$ARCH.AppImage"
      echo "    → build/bin/ComfyGo-$VERSION-$ARCH.AppImage"
      rm -rf "$APPDIR"
    else
      echo "    appimagetool not found — skipping .AppImage"
    fi
    ;;

  darwin)
    echo ""
    echo "  Creating macOS app bundle..."
    BIN="$(ls build/bin/comfygo* 2>/dev/null | head -1)"
    if [ -n "$BIN" ]; then
      BUNDLE="build/ComfyGo.app"
      mkdir -p "$BUNDLE/Contents/MacOS"
      mkdir -p "$BUNDLE/Contents/Resources"
      cp "$BIN" "$BUNDLE/Contents/MacOS/comfygo"
      cp build/appicon.png "$BUNDLE/Contents/Resources/icon.png"

      cat > "$BUNDLE/Contents/Info.plist" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleExecutable</key>
  <string>comfygo</string>
  <key>CFBundleIdentifier</key>
  <string>app.comfygo.desktop</string>
  <key>CFBundleName</key>
  <string>ComfyGo</string>
  <key>CFBundleVersion</key>
  <string>$VERSION</string>
  <key>CFBundleShortVersionString</key>
  <string>$VERSION</string>
  <key>CFBundleIconFile</key>
  <string>icon</string>
  <key>NSHighResolutionCapable</key>
  <true/>
</dict>
</plist>
EOF
      echo "    → build/ComfyGo.app"
    fi
    ;;

  windows)
    echo ""
    echo "  Windows binary built at build/bin/comfygo*.exe"
    echo "  Use an installer tool like NSIS to create a setup package."
    ;;
esac

echo ""
echo "━━━ Package complete ━━━━━━━━━━━━━━━━━━━━━━━"
ls -lh build/bin/ 2>/dev/null || true
ls -lh build/ComfyGo.app 2>/dev/null || true
echo ""
