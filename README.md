# TUI File Server 📁⚡

A fast, lightweight, terminal-based file browser and HTTP file server built in Go with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lip Gloss](https://github.com/charmbracelet/lipgloss).

Browse your filesystem through an interactive TUI, select any file to instantly stream or share over HTTP (Local and LAN), and monitor incoming requests in real-time.

---

## ✨ Features

- **Interactive TUI Navigator**: Browse directories with arrow keys or Vim keybindings (`h`/`j`/`k`/`l`). Includes Windows drive selection ("This PC") and parent navigation (`..`).
- **Instant HTTP File Server**: Select any file and press `Enter` to serve it immediately.
- **LAN & Local Access**: Automatically detects and displays your local (`localhost`) and network LAN IP addresses.
- **Media Streaming Ready**: Supports HTTP Range Requests (`Accept-Ranges: bytes`) and CORS headers for seamless streaming in VLC, mpv, web browsers, or `curl`.
- **Smart Port Fallback**: Automatically finds an available port if the default or requested port is in use.
- **Live Request Logging**: Real-time request monitoring with status codes, transfer sizes, response durations, and scrollable log history.

---

## 🚀 Installation & Usage

### Prerequisites
- [Go 1.26+](https://go.dev/) (or modern Go compiler)

### Running directly
```sh
# Start in the current directory (default port: 8000)
go run main.go

# Start in a specific directory
go run main.go C:\path\to\folder

# Specify a custom preferred port
go run main.go -p 8080
```

## 🔨 Building for Each Platform

Go provides native cross-compilation without needing third-party tools or C compilers. Set the `GOOS` (target OS) and `GOARCH` (target architecture) environment variables before building.

### 🪟 Windows

| Target Architecture | PowerShell Command | Bash / Linux / macOS Command |
| :--- | :--- | :--- |
| **64-bit (x86_64 / amd64)** | `$env:GOOS="windows"; $env:GOARCH="amd64"; go build -o dist/file-server-windows-amd64.exe main.go` | `GOOS=windows GOARCH=amd64 go build -o dist/file-server-windows-amd64.exe main.go` |
| **ARM64 (Surface / Windows on ARM)** | `$env:GOOS="windows"; $env:GOARCH="arm64"; go build -o dist/file-server-windows-arm64.exe main.go` | `GOOS=windows GOARCH=arm64 go build -o dist/file-server-windows-arm64.exe main.go` |
| **32-bit (x86 / 386)** | `$env:GOOS="windows"; $env:GOARCH="386"; go build -o dist/file-server-windows-386.exe main.go` | `GOOS=windows GOARCH=386 go build -o dist/file-server-windows-386.exe main.go` |

---

### 🐧 Linux

| Target Architecture | PowerShell Command | Bash / Linux / macOS Command |
| :--- | :--- | :--- |
| **64-bit (x86_64 / amd64)** | `$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o dist/file-server-linux-amd64 main.go` | `GOOS=linux GOARCH=amd64 go build -o dist/file-server-linux-amd64 main.go` |
| **ARM64 (Raspberry Pi 4/5, AWS Graviton)** | `$env:GOOS="linux"; $env:GOARCH="arm64"; go build -o dist/file-server-linux-arm64 main.go` | `GOOS=linux GOARCH=arm64 go build -o dist/file-server-linux-arm64 main.go` |
| **ARMv7 (Raspberry Pi 2/3, 32-bit)** | `$env:GOOS="linux"; $env:GOARCH="arm"; $env:GOARM="7"; go build -o dist/file-server-linux-armv7 main.go` | `GOOS=linux GOARCH=arm GOARM=7 go build -o dist/file-server-linux-armv7 main.go` |
| **32-bit (x86 / 386)** | `$env:GOOS="linux"; $env:GOARCH="386"; go build -o dist/file-server-linux-386 main.go` | `GOOS=linux GOARCH=386 go build -o dist/file-server-linux-386 main.go` |

---

### 🍎 macOS (Darwin)

| Target Architecture | PowerShell Command | Bash / Linux / macOS Command |
| :--- | :--- | :--- |
| **Apple Silicon (M1 / M2 / M3 / M4 - ARM64)** | `$env:GOOS="darwin"; $env:GOARCH="arm64"; go build -o dist/file-server-darwin-arm64 main.go` | `GOOS=darwin GOARCH=arm64 go build -o dist/file-server-darwin-arm64 main.go` |
| **Intel (x86_64 / amd64)** | `$env:GOOS="darwin"; $env:GOARCH="amd64"; go build -o dist/file-server-darwin-amd64 main.go` | `GOOS=darwin GOARCH=amd64 go build -o dist/file-server-darwin-amd64 main.go` |
| **Universal Binary (macOS)** *(requires lipo on macOS)* | — | `lipo -create -output dist/file-server-darwin-universal dist/file-server-darwin-amd64 dist/file-server-darwin-arm64` |

---

### 📦 Build All Platforms at Once

#### In PowerShell (Windows)
```powershell
New-Item -ItemType Directory -Force -Path dist
$targets = @(
    @{OS="windows"; Arch="amd64"; Out="dist/file-server-windows-amd64.exe"},
    @{OS="windows"; Arch="arm64"; Out="dist/file-server-windows-arm64.exe"},
    @{OS="linux";   Arch="amd64"; Out="dist/file-server-linux-amd64"},
    @{OS="linux";   Arch="arm64"; Out="dist/file-server-linux-arm64"},
    @{OS="darwin";  Arch="amd64"; Out="dist/file-server-darwin-amd64"},
    @{OS="darwin";  Arch="arm64"; Out="dist/file-server-darwin-arm64"}
)
foreach ($t in $targets) {
    Write-Host "Building for $($t.OS)/$($t.Arch)..."
    $env:GOOS = $t.OS; $env:GOARCH = $t.Arch
    go build -ldflags="-s -w" -o $t.Out main.go
}
Remove-Item Env:\GOOS, Env:\GOARCH
```

#### In Bash (Linux / macOS)
```bash
mkdir -p dist
platforms=("windows/amd64/.exe" "windows/arm64/.exe" "linux/amd64/" "linux/arm64/" "darwin/amd64/" "darwin/arm64/")

for platform in "${platforms[@]}"; do
    IFS="/" read -r GOOS GOARCH EXT <<< "$platform"
    echo "Building for $GOOS/$GOARCH..."
    GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w" -o "dist/file-server-${GOOS}-${GOARCH}${EXT}" main.go
done
```

---

## ⌨️ Controls & Keybindings

### Navigation Mode
| Key | Action |
| --- | --- |
| `↑` / `k` | Move cursor up |
| `↓` / `j` | Move cursor down |
| `←` / `h` | Go to parent directory |
| `→` / `l` / `Enter` | Enter directory / Serve selected file |
| `PgUp` / `b` | Page up |
| `PgDown` / `f` | Page down |
| `Home` / `g` | Jump to top |
| `End` / `G` | Jump to bottom |
| `r` | Refresh current directory |
| `q` / `Ctrl+C` | Quit application |

### Serving Mode
| Key | Action |
| --- | --- |
| `↑` / `k` | Scroll request logs up |
| `↓` / `j` | Scroll request logs down |
| `Home` / `g` | Scroll to oldest logs |
| `End` / `G` | Scroll to newest logs (auto-scroll) |
| `Esc` | Stop server and return to navigation |
| `q` / `Ctrl+C` | Quit application |

---

## 🧪 Testing

Run all unit and integration tests:
```sh
go test ./...
```
