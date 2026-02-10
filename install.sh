#!/usr/bin/env bash
set -e

REPO="coolamit/mermaid-cli"
BINARY="mmd-cli"

# ── Helpers ──────────────────────────────────────────────────────────────────

info()  { printf '\033[1;34m::\033[0m %s\n' "$*"; }
warn()  { printf '\033[1;33mwarning:\033[0m %s\n' "$*"; }
error() { printf '\033[1;31merror:\033[0m %s\n' "$*" >&2; }
die()   { error "$*"; exit 1; }

need() {
  command -v "$1" >/dev/null 2>&1 || die "Required tool '$1' not found. Please install it and try again."
}

# ── Detect OS and architecture ───────────────────────────────────────────────

detect_platform() {
  OS=$(uname -s)
  case "$OS" in
    Linux)  OS="linux" ;;
    Darwin) OS="macos" ;;
    *)      die "Unsupported operating system: $OS" ;;
  esac

  ARCH=$(uname -m)
  case "$ARCH" in
    x86_64)       ARCH="x64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *)            die "Unsupported architecture: $ARCH" ;;
  esac

  info "Detected platform: ${OS}/${ARCH}"
}

# ── Check installed version ──────────────────────────────────────────────────

check_installed_version() {
  if command -v "$BINARY" >/dev/null 2>&1; then
    INSTALLED_VERSION=$("$BINARY" --version 2>/dev/null | awk '{print $NF}' || true)
  else
    INSTALLED_VERSION=""
  fi
}

# ── Chrome / Chromium check & install ────────────────────────────────────────

has_chrome() {
  # Check PATH for common binary names
  for bin in chromium chromium-browser google-chrome google-chrome-stable; do
    command -v "$bin" >/dev/null 2>&1 && return 0
  done

  # Check well-known macOS app locations
  [ -d "/Applications/Google Chrome.app" ] && return 0
  [ -d "/Applications/Chromium.app" ] && return 0

  return 1
}

ensure_chrome() {
  if has_chrome; then
    info "Chrome/Chromium found"
    return
  fi

  warn "Chrome or Chromium is required but was not found."

  case "$OS" in
    linux)
      if command -v apt-get >/dev/null 2>&1; then
        info "Installing Chromium via apt..."
        sudo apt-get update -qq
        sudo apt-get install -y chromium-browser || sudo apt-get install -y chromium
      elif command -v dnf >/dev/null 2>&1; then
        info "Installing Chromium via dnf..."
        sudo dnf install -y chromium
      elif command -v yum >/dev/null 2>&1; then
        info "Installing Chromium via yum..."
        sudo yum install -y chromium
      else
        die "Could not detect a supported package manager. Please install Chrome or Chromium manually."
      fi
      ;;
    macos)
      if command -v brew >/dev/null 2>&1; then
        info "Installing Chromium via Homebrew..."
        brew install --cask chromium
      else
        echo ""
        echo "  Please install Chrome or Chromium manually:"
        echo "    Google Chrome  - https://www.google.com/chrome/"
        echo "    Chromium       - https://www.chromium.org/getting-involved/download-chromium/"
        echo ""
        die "No supported installation method found for Chrome/Chromium on macOS."
      fi
      ;;
  esac

  has_chrome || die "Chrome/Chromium installation appears to have failed. Please install it manually."
  info "Chrome/Chromium installed successfully"
}

# ── Determine version ────────────────────────────────────────────────────────

get_version() {
  if [ -n "$1" ]; then
    VERSION="$1"
    # Strip leading 'v' if present for the download URL
    VERSION_NUM="${VERSION#v}"
    info "Using requested version: ${VERSION_NUM}"
    return
  fi

  need curl

  info "Fetching latest release version..."
  VERSION=$(curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' \
    | sed -E 's/.*"tag_name":\s*"([^"]+)".*/\1/')

  [ -n "$VERSION" ] || die "Could not determine latest version. Check https://github.com/${REPO}/releases"

  VERSION_NUM="${VERSION#v}"
  info "Latest version: ${VERSION_NUM}"
}

# ── Download and install ─────────────────────────────────────────────────────

install_binary() {
  need curl
  need tar

  ARCHIVE="${BINARY}_${VERSION_NUM}_${OS}_${ARCH}.tar.gz"
  URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE}"

  TMPDIR=$(mktemp -d)
  trap 'rm -rf "$TMPDIR"' EXIT

  info "Downloading ${URL}..."
  curl -sSL -o "${TMPDIR}/${ARCHIVE}" "$URL" \
    || die "Download failed. Check that version '${VERSION}' exists at https://github.com/${REPO}/releases"

  info "Extracting..."
  tar -xzf "${TMPDIR}/${ARCHIVE}" -C "$TMPDIR"

  [ -f "${TMPDIR}/${BINARY}" ] || die "Binary '${BINARY}' not found in archive"

  chmod +x "${TMPDIR}/${BINARY}"

  # Choose install directory
  INSTALL_DIR=""

  if [ -d "$HOME/.local/bin" ] || mkdir -p "$HOME/.local/bin" 2>/dev/null; then
    INSTALL_DIR="$HOME/.local/bin"
  elif [ -w "/usr/local/bin" ]; then
    INSTALL_DIR="/usr/local/bin"
  else
    die "Cannot find a writable install directory. Create ~/.local/bin or run with sudo."
  fi

  mv "${TMPDIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
  info "Installed ${BINARY} to ${INSTALL_DIR}/${BINARY}"
}

# ── Verify installation ─────────────────────────────────────────────────────

verify() {
  if command -v "$BINARY" >/dev/null 2>&1; then
    INSTALLED_VERSION=$("$BINARY" --version 2>/dev/null || true)
    info "Verified: ${BINARY} ${INSTALLED_VERSION}"
  elif [ -x "${INSTALL_DIR}/${BINARY}" ]; then
    INSTALLED_VERSION=$("${INSTALL_DIR}/${BINARY}" --version 2>/dev/null || true)
    info "Verified: ${BINARY} ${INSTALLED_VERSION}"

    # Check if install dir is in PATH
    case ":$PATH:" in
      *":${INSTALL_DIR}:"*) ;;
      *)
        echo ""
        warn "${INSTALL_DIR} is not in your PATH."
        echo "  Add it by running:"
        echo ""
        echo "    export PATH=\"${INSTALL_DIR}:\$PATH\""
        echo ""
        echo "  To make this permanent, add the line above to your ~/.bashrc, ~/.zshrc, or equivalent."
        echo ""
        ;;
    esac
  else
    warn "Could not verify installation. You may need to add ${INSTALL_DIR} to your PATH."
  fi
}

# ── Main ─────────────────────────────────────────────────────────────────────

main() {
  detect_platform
  check_installed_version
  ensure_chrome
  get_version "$1"

  if [ -n "$INSTALLED_VERSION" ] && [ "$INSTALLED_VERSION" = "$VERSION_NUM" ]; then
    info "${BINARY} ${VERSION_NUM} is already installed and up-to-date."
    exit 0
  fi

  if [ -n "$INSTALLED_VERSION" ]; then
    info "Updating ${BINARY} from ${INSTALLED_VERSION} to ${VERSION_NUM}..."
  fi

  install_binary
  verify

  info "Done!"
}

main "$@"
