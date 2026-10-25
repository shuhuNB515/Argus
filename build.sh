#!/usr/bin/env bash
# Argus cross-platform release builder.
#
# The default mode builds one target and embeds the already-exported frontend.
# `./build.sh --release` builds and packages all supported desktop/server targets.
#
# Environment variables:
#   ARGUS_TARGET_OS=linux             One target OS in single-target mode.
#   ARGUS_TARGET_ARCH=amd64           One target arch in single-target mode.
#   ARGUS_TARGETS=linux/amd64,...     Comma-separated targets for multi-target mode.
#   ARGUS_BUILD_VERSION=v0.3.3        Version embedded in the binary and archive name.
#   ARGUS_OUTPUT=/path/to/argus       Explicit binary path in single-target mode.
#   ARGUS_OUTPUT_DIR=dist             Directory for default binary paths.
#   ARGUS_PACKAGE=1                   Create a zip archive for each target.
#   ARGUS_PACKAGE_DIR=dist            Directory for release archives.
#   ARGUS_COMPRESS=off                UPX mode: off, auto, or required.
#   ARGUS_UPX_ARGS="--best --lzma"    Arguments passed to UPX.
#   ARGUS_SKIP_FRONTEND=1             Reuse server/webui/dist (for CI artifact builds).
#   ARGUS_SKIP_NPM_CI=1               Skip npm ci while rebuilding the frontend.
#   ARGUS_GOSUMDB=sum.golang.org      Go checksum database.
set -euo pipefail

cd "$(cd "$(dirname "$0")" && pwd)"

info() { printf '\033[36m[*]\033[0m %s\n' "$*"; }
ok() { printf '\033[32m[+]\033[0m %s\n' "$*"; }
warn() { printf '\033[33m[!]\033[0m %s\n' "$*" >&2; }
die() { printf '\033[31m[x]\033[0m %s\n' "$*" >&2; exit 1; }

usage() {
  cat <<'EOF'
Usage:
  ./build.sh                         Build for the current system and architecture.
  ./build.sh --target linux/amd64   Build one selected target.
  ./build.sh --release               Build and package all supported targets.

Options:
  --release              Build Linux, macOS, and Windows amd64/arm64 targets and create zip archives
  --target OS/ARCH       Select one target, for example windows/amd64
  --upx                  Compress binaries with UPX
  --no-compress          Disable UPX compression
  --help                 Show this help

Override ARGUS_TARGETS, for example:
  ARGUS_TARGETS=linux/amd64,windows/amd64 ./build.sh --release
EOF
}

RELEASE_TARGETS_DEFAULT="linux/amd64,linux/arm64,darwin/amd64,darwin/arm64,windows/amd64"
ARGUS_RELEASE="${ARGUS_RELEASE:-0}"
ARGUS_COMPRESS="${ARGUS_COMPRESS:-off}"
ARGUS_PACKAGE="${ARGUS_PACKAGE:-0}"

while [ "$#" -gt 0 ]; do
  case "$1" in
    --release)
      ARGUS_RELEASE=1
      ARGUS_PACKAGE=1
      shift
      ;;
    --target)
      [ "$#" -ge 2 ] || die "--target requires an OS/ARCH value"
      target_arg="$2"
      case "$target_arg" in
        */*)
          ARGUS_TARGET_OS="${target_arg%%/*}"
          ARGUS_TARGET_ARCH="${target_arg##*/}"
          ARGUS_TARGETS="$target_arg"
          ;;
        *) die "target must use OS/ARCH format, for example linux/amd64" ;;
      esac
      shift 2
      ;;
    --no-compress)
      ARGUS_COMPRESS=0
      shift
      ;;
    --upx)
      ARGUS_COMPRESS=required
      shift
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *) die "unknown argument:  (run --help for usage)" ;;
  esac
done

command -v go >/dev/null 2>&1 || die "Go 1.26 or newer is required"

ARGUS_GOSUMDB="${ARGUS_GOSUMDB:-sum.golang.org}"
if [ -z "${ARGUS_BUILD_VERSION:-}" ]; then
  if command -v git >/dev/null 2>&1 && git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    ARGUS_BUILD_VERSION="$(git describe --tags --always --dirty)"
  else
    ARGUS_BUILD_VERSION="dev"
  fi
fi
# Release tags are commonly passed as v0.3.3; keep the binary version consistent.
ARGUS_BUILD_VERSION="${ARGUS_BUILD_VERSION#v}"
ARGUS_OUTPUT_DIR="${ARGUS_OUTPUT_DIR:-dist}"
ARGUS_PACKAGE_DIR="${ARGUS_PACKAGE_DIR:-$ARGUS_OUTPUT_DIR}"
ARGUS_UPX_ARGS="${ARGUS_UPX_ARGS:---best --lzma}"

if [ "${ARGUS_RELEASE}" = "1" ]; then
  ARGUS_TARGETS="${ARGUS_TARGETS:-$RELEASE_TARGETS_DEFAULT}"
else
  ARGUS_TARGET_OS="${ARGUS_TARGET_OS:-$(GOSUMDB="$ARGUS_GOSUMDB" go env GOOS)}"
  ARGUS_TARGET_ARCH="${ARGUS_TARGET_ARCH:-$(GOSUMDB="$ARGUS_GOSUMDB" go env GOARCH)}"
  ARGUS_TARGETS="${ARGUS_TARGETS:-${ARGUS_TARGET_OS}/${ARGUS_TARGET_ARCH}}"
fi

if [ "${ARGUS_SKIP_FRONTEND:-0}" = "1" ]; then
  [ -d server/webui/dist ] || die "ARGUS_SKIP_FRONTEND=1 requires server/webui/dist"
else
  command -v npm >/dev/null 2>&1 || die "Node.js/npm is required to build the frontend"
  command -v rsync >/dev/null 2>&1 || die "rsync is required"
  info "Building frontend assets"
  if [ "${ARGUS_SKIP_NPM_CI:-0}" != "1" ]; then
    (cd web && npm ci)
  fi
  (cd web && npm run build:static)
  info "Copying frontend assets into server/webui/dist"
  mkdir -p server/webui/dist
  rsync -a --delete web/out/ server/webui/dist/
fi

compress_binary() {
  binary="$1"
  goos="$2"
  case "$ARGUS_COMPRESS" in
    0|off|false|none)
      info "Skipping UPX: $binary"
      return 0
      ;;
    auto|required|true|1) ;;
    *) die "ARGUS_COMPRESS must be off, auto, or required" ;;
  esac

  if ! command -v upx >/dev/null 2>&1; then
    if [ "$ARGUS_COMPRESS" = "required" ]; then
      die "ARGUS_COMPRESS=required but upx was not found"
    fi
    warn "upx was not found; keeping the linker-compressed binary: $binary"
    return 0
  fi

  before=$(wc -c < "$binary" | tr -d ' ')
  upx_args="$ARGUS_UPX_ARGS"
  [ "$goos" = "darwin" ] && upx_args="$upx_args --force-macos"
  # shellcheck disable=SC2086
  if ! upx $upx_args -- "$binary"; then
    if [ "$ARGUS_COMPRESS" = "required" ]; then
      die "UPX compression failed: $binary"
    fi
    warn "UPX does not support this target; keeping the uncompressed binary: $binary"
    return 0
  fi
  after=$(wc -c < "$binary" | tr -d ' ')
  ok "UPX compression complete: $binary (${before} -> ${after} bytes)"
}

package_binary() {
  binary="$1"
  goos="$2"
  goarch="$3"
	package_name="argus-${ARGUS_BUILD_VERSION}-${goos}-${goarch}"
  package_root="${ARGUS_PACKAGE_DIR}/${package_name}"
  archive="${ARGUS_PACKAGE_DIR}/${package_name}.zip"

  command -v zip >/dev/null 2>&1 || die "zip is required for packaging"
  rm -rf "$package_root" "$archive"
  mkdir -p "$package_root"
  if [ "$goos" = "darwin" ]; then
    app_root="$package_root/Argus.app/Contents"
    mkdir -p "$app_root/MacOS" "$app_root/Resources"
    cp "$binary" "$app_root/MacOS/Argus-bin"
    cp packaging/macos/Argus "$app_root/MacOS/Argus"
    chmod +x "$app_root/MacOS/Argus" "$app_root/MacOS/Argus-bin"
    sed "s/__ARGUS_VERSION__/${ARGUS_BUILD_VERSION}/g" packaging/macos/Info.plist > "$app_root/Info.plist"
    cp -R skills "$app_root/Resources/"
    cp config.example.json "$app_root/Resources/"
    if [ -f README.package.md ]; then
      cp README.package.md "$app_root/Resources/README.md"
    elif [ -f README.md ]; then
      cp README.md "$app_root/Resources/README.md"
    fi
    # Keep a copy beside the app for users who open the archive before installing.
    if [ -f README.package.md ]; then
      cp README.package.md "$package_root/README.md"
    elif [ -f README.md ]; then
      cp README.md "$package_root/"
    fi
  else
    cp "$binary" "$package_root/"
  fi
  if [ "$goos" = "windows" ]; then
    cp start.bat "$package_root/"
  elif [ "$goos" != "darwin" ]; then
    cp start.sh "$package_root/"
    chmod +x "$package_root/start.sh"
  fi
  if [ "$goos" != "darwin" ]; then
    cp -R skills "$package_root/"
    cp config.example.json "$package_root/"
    if [ -f README.package.md ]; then
      cp README.package.md "$package_root/README.md"
    elif [ -f README.md ]; then
      cp README.md "$package_root/"
    fi
  fi
  (cd "$ARGUS_PACKAGE_DIR" && zip -q -r -9 "$(basename "$archive")" "$(basename "$package_root")")
  if [ "$goos" = "darwin" ] && command -v hdiutil >/dev/null 2>&1; then
    dmg_stage=$(mktemp -d "${TMPDIR:-/tmp}/argus-dmg.XXXXXX")
    cp -R "$package_root/Argus.app" "$dmg_stage/"
    if [ -f "$package_root/README.md" ]; then cp "$package_root/README.md" "$dmg_stage/"; fi
    ln -s /Applications "$dmg_stage/Applications"
    dmg="$ARGUS_PACKAGE_DIR/${package_name}.dmg"
    rm -f "$dmg"
    hdiutil create -quiet -volname Argus -srcfolder "$dmg_stage" -ov -format UDZO "$dmg"
    rm -rf "$dmg_stage"
    ok "macOS installer image: $dmg"
  fi
  rm -rf "$package_root"
  ok "Release archive: $archive"
}

build_target() {
  target="$1"
  case "$target" in
    */*) ;;
    *) die "invalid target:  (expected OS/ARCH)" ;;
  esac
  goos="${target%%/*}"
  goarch="${target##*/}"
  case "$goos" in
    linux|darwin|windows) ;;
    *) die "unsupported system:  (supported: linux, darwin, windows)" ;;
  esac

  binary_name="argus"
  [ "$goos" = "windows" ] && binary_name="argus.exe"
  if [ -n "${ARGUS_OUTPUT:-}" ] && [ "$ARGUS_RELEASE" != "1" ]; then
    output="$ARGUS_OUTPUT"
  else
    output="${ARGUS_OUTPUT_DIR}/argus-${goos}-${goarch}/${binary_name}"
  fi
  mkdir -p "$(dirname "$output")"

  info "Building ${goos}/${goarch}, version ${ARGUS_BUILD_VERSION}"
  GOSUMDB="$ARGUS_GOSUMDB" \
  CGO_ENABLED=0 \
  GOOS="$goos" \
  GOARCH="$goarch" \
  go build \
    -tags embedui \
    -trimpath \
    -ldflags "-s -w -buildid= -X main.version=${ARGUS_BUILD_VERSION}" \
    -o "$output" \
    ./cmd/argus

  compress_binary "$output" "$goos"
  if command -v file >/dev/null 2>&1; then file "$output"; fi
  if [ "$ARGUS_PACKAGE" = "1" ]; then package_binary "$output" "$goos" "$goarch"; fi
  ok "Build complete: $output"
}

write_checksums() {
  [ "$ARGUS_PACKAGE" = "1" ] || return 0
  checksum_file="$ARGUS_PACKAGE_DIR/SHA256SUMS"
  if command -v sha256sum >/dev/null 2>&1; then
    (cd "$ARGUS_PACKAGE_DIR" && for archive in *.zip *.dmg; do if [ -f "$archive" ]; then sha256sum "$archive"; fi; done > "$(basename "$checksum_file")")
  elif command -v shasum >/dev/null 2>&1; then
    (cd "$ARGUS_PACKAGE_DIR" && for archive in *.zip *.dmg; do if [ -f "$archive" ]; then shasum -a 256 "$archive"; fi; done > "$(basename "$checksum_file")")
  else
    warn "sha256sum or shasum not found; skipping SHA256SUMS"
    return 0
  fi
  ok "Checksum file: $checksum_file"
}

mkdir -p "$ARGUS_OUTPUT_DIR"
if [ "$ARGUS_PACKAGE" = "1" ]; then mkdir -p "$ARGUS_PACKAGE_DIR"; fi

old_ifs="$IFS"
IFS=','
read -r -a targets <<< "$ARGUS_TARGETS"
IFS="$old_ifs"
[ "${#targets[@]}" -gt 0 ] || die "ARGUS_TARGETS cannot be empty"
for target in "${targets[@]}"; do
  target="${target//[[:space:]]/}"
  [ -n "$target" ] || continue
  build_target "$target"
done

if [ "$ARGUS_PACKAGE" = "1" ]; then
  write_checksums
  info "Release packages created: $ARGUS_PACKAGE_DIR"
fi
