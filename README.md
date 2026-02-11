# Mermaid CLI

A drop-in replacement for `@mermaid-js/mermaid-cli` written in Go. Produces a single static binary with no Node.js dependency. Requires Chrome or Chromium at runtime.

Converts Mermaid diagram definitions into SVG, PNG, and PDF files using a headless Chrome browser.

---

## Table of Contents

- [How it Works](#how-it-works)
- [Requirements](#requirements)
- [Installation](#installation)
  - [Quick Install/Update](#quick-installupdate)
  - [Go Install](#go-install)
  - [Manual Download](#manual-download)
  - [From Source](#from-source)
  - [Cross-Compilation](#cross-compilation)
  - [Docker](#docker)
- [Usage](#usage)
- [CLI Flags](#cli-flags)
- [Configuration Files](#configuration-files)
  - [Mermaid Config (-c)](#mermaid-config--c)
  - [Browser Config (-p)](#browser-config--p)
  - [CSS File (-C)](#css-file--c)
- [Docker](#docker-1)
  - [Start / Stop](#start--stop)
  - [Run Commands](#run-commands)
  - [One-off (Ephemeral)](#one-off-ephemeral)
  - [Without Docker Compose](#without-docker-compose)

## How it Works

- Single Go binary with mermaid.js embedded via `go:embed`
- Launches headless Chrome via [chromedp](https://github.com/chromedp/chromedp) (Chrome DevTools Protocol)
- Builds an HTML page with the mermaid diagram definition
- Chrome renders the diagram, then extracts SVG / captures PNG screenshot / prints PDF
- Browser instance is reused across multiple renders for efficiency

## Requirements

- Go 1.25+ (for building from source)
- Chrome or Chromium installed on the system (required at runtime)
  - On macOS: bundled with the system or install via `brew install --cask chromium`
  - On Linux: `apt install chromium-browser` or `apk add chromium`
  - Or use the [Docker image](#docker) which bundles Chromium

## Installation

### Quick Install/Update

Detects your OS/architecture, ensures Chrome/Chromium is available, and installs/updates the latest pre-built binary:

```sh
curl -L https://raw.githubusercontent.com/coolamit/mermaid-cli/master/install.sh | sh
```

### Go Install

```sh
go install github.com/coolamit/mermaid-cli/cmd/mmd-cli@latest
```

### Manual Download

Download the archive for your platform from [GitHub Releases](https://github.com/coolamit/mermaid-cli/releases), extract it, and move the `mmd-cli` binary to a directory in your PATH.

### From Source

```bash
# Build from source
make build
```

### Cross-Compilation

| Make target         | Platform | Architecture                        | Output binary         |
|---------------------|----------|-------------------------------------|-----------------------|
| `build-linux-x64`   | Linux    | x64 (Intel/AMD)                     | `mmd-cli-linux-x64`   |
| `build-linux-arm64` | Linux    | ARM64 (Apple Silicon, AWS Graviton) | `mmd-cli-linux-arm64` |
| `build-macos-x64`   | macOS    | x64 (Intel Mac)                     | `mmd-cli-macos-x64`   |
| `build-macos-arm64` | macOS    | ARM64 (Apple Silicon)               | `mmd-cli-macos-arm64` |

```bash
# Build for all platforms
make build-all

# Individual targets
make build-linux-x64
make build-linux-arm64
make build-macos-x64
make build-macos-arm64
```

### Docker

```bash
# Using Docker Compose (recommended)
make docker-up

# Or build manually
docker build -t mmd-cli .
```

The Docker image pre-bundles Chromium — no local Chrome installation needed. See [Docker usage](#docker-1) for full details.

## Usage

```bash
# Basic SVG output
mmd-cli -i diagram.mmd -o diagram.svg

# PNG output with scale factor
mmd-cli -i diagram.mmd -o diagram.png -s 2

# PDF output fitted to content
mmd-cli -i diagram.mmd -o diagram.pdf -f

# SVG with dimensions matching the diagram size
mmd-cli -i diagram.mmd -o diagram.svg --svgFit

# Read from stdin, write to stdout
echo "graph TD; A-->B" | mmd-cli -i - -o - -e svg

# Process markdown file (renders all mermaid blocks)
mmd-cli -i document.md -o output.md

# Use dark theme
mmd-cli -i diagram.mmd -o diagram.svg -t dark

# With custom mermaid config
mmd-cli -i diagram.mmd -o diagram.svg -c config.json

# With browser config
mmd-cli -i diagram.mmd -o diagram.svg -p browser-config.json

# Transparent background PNG
mmd-cli -i diagram.mmd -o diagram.png -b transparent

# With icon packs
mmd-cli -i diagram.mmd -o diagram.svg --iconPacks @iconify-json/logos
```

## CLI Flags

| Flag                      | Short | Default       | Description                              |
|---------------------------|-------|---------------|------------------------------------------|
| `--input`                 | `-i`  | (required)    | Input mermaid file. Use `-` for stdin.   |
| `--output`                | `-o`  | `{input}.svg` | Output file. Use `-` for stdout.         |
| `--artefacts`             | `-a`  | output dir    | Artefacts output path (markdown mode)    |
| `--theme`                 | `-t`  | `default`     | Theme: default, forest, dark, neutral    |
| `--width`                 | `-w`  | `800`         | Page width                               |
| `--height`                | `-H`  | `600`         | Page height                              |
| `--backgroundColor`       | `-b`  | `white`       | Background color                         |
| `--outputFormat`          | `-e`  | auto          | Output format: svg, png, pdf             |
| `--scale`                 | `-s`  | `1`           | Scale factor                             |
| `--pdfFit`                | `-f`  | `false`       | Scale PDF to fit chart                   |
| `--svgFit`                |       | `false`       | Set SVG dimensions to match diagram size |
| `--svgId`                 | `-I`  |               | SVG element id attribute                 |
| `--configFile`            | `-c`  |               | Mermaid JSON config file                 |
| `--cssFile`               | `-C`  |               | CSS file for styling                     |
| `--puppeteerConfigFile`   | `-p`  |               | Browser JSON config file                 |
| `--iconPacks`             |       |               | Icon packs (e.g. @iconify-json/logos)    |
| `--iconPacksNamesAndUrls` |       |               | Icon packs as name#url                   |
| `--quiet`                 | `-q`  | `false`       | Suppress log output                      |
| `--version`               |       |               | Show version                             |

## Configuration Files

### Mermaid Config (-c)

JSON file with [mermaid.js configuration options](https://mermaid.js.org/config/schema-docs/config.html). Passed via `--configFile` / `-c`.

```json
{
  "theme": "forest",
  "flowchart": {
    "curve": "basis"
  },
  "sequence": {
    "mirrorActors": false
  }
}
```

### Browser Config (-p)

JSON file with browser launch options. Passed via `--puppeteerConfigFile` / `-p`.

| Field            | Type     | Description                         |
|------------------|----------|-------------------------------------|
| `executablePath` | string   | Path to Chrome/Chromium binary      |
| `args`           | string[] | Extra command-line flags for Chrome |
| `timeout`        | int      | Browser launch timeout (ms)         |
| `headless`       | string   | Headless mode (`"new"` or `"old"`)  |

```json
{
  "executablePath": "/usr/bin/chromium-browser",
  "args": ["--no-sandbox", "--disable-gpu"],
  "timeout": 30000,
  "headless": "new"
}
```

### CSS File (-C)

Custom CSS file applied to the diagram page. Passed via `--cssFile` / `-C`. Useful for custom fonts or overriding default mermaid styles.

## Docker

### Start / Stop

```bash
# Build image (if needed) and start the container in the background
make docker-up

# Stop the container (preserves it for quick restart)
make docker-down

# Stop + remove container, image, and volumes
make docker-clean
```

### Run Commands

With the container running (`make docker-up`), exec commands into it:

```bash
# SVG output
make docker-run -- mmd-cli -i diagram.mmd -o diagram.svg

# PNG output with scale factor
make docker-run -- mmd-cli -i diagram.mmd -o diagram.png -s 2

# PDF output fitted to content
make docker-run -- mmd-cli -i diagram.mmd -o diagram.pdf -f

# Check version
make docker-run -- mmd-cli --version
```

### One-off (Ephemeral)

Spin up a temporary container, run `mmd-cli`, and destroy it — no `docker-up` needed:

```bash
make docker-run-aloof -- -i diagram.mmd -o diagram.svg
```

### Without Docker Compose

If you prefer bare `docker run` commands:

```bash
# Build
docker build -t mmd-cli .

# SVG output
docker run --rm -v "$(pwd)":/data mmd-cli -i diagram.mmd -o diagram.svg

# PNG output
docker run --rm -v "$(pwd)":/data mmd-cli -i diagram.mmd -o diagram.png -s 2
```

The Docker image sets `CHROME_BIN=/usr/bin/chromium-browser` and `WORKDIR /data` automatically, so local filenames work directly.
