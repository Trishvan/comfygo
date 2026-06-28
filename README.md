# ComfyGo

Desktop application for Stable Diffusion 1.5 image generation. Go backend with Svelte frontend, powered by [stable-diffusion.cpp](https://github.com/leejet/stable-diffusion.cpp).

## Quick Start

```bash
git clone https://github.com/Trishvan/comfygo.git
cd comfygo
make setup              # one command — installs everything
cp model.safetensors ~/.comfygo/models/
make dev                # start in development mode
```

`make setup` auto-detects your OS, installs system dependencies, downloads the prebuilt stable-diffusion.cpp library, installs frontend packages, and generates Wails bindings.

## Prerequisites

- **Go 1.23+** — [go.dev/dl](https://go.dev/dl/)
- **Node.js 20+** — [nodejs.org](https://nodejs.org/)

`make setup` will verify these and report if anything is missing.

## Makefile Targets

| Target | Description |
|--------|-------------|
| `make setup` | Full one-time setup (system deps, library, npm, bindings) |
| `make dev` | Start `wails dev` server |
| `make build` | Production build to `build/bin/comfygo` |
| `make clean` | Remove build artifacts |
| `make bindings` | Regenerate Wails Go→TypeScript bindings |
| `make lint` | Run Svelte type-check |

## Manual Setup

If you prefer to do things step by step:

```bash
# 1. Install system dependencies (Linux)
sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.1-dev unzip

# 2. Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0

# 3. Download stable-diffusion.cpp library
#    Pick the right archive for your platform from:
#    https://github.com/leejet/stable-diffusion.cpp/releases/tag/master-700-c2df4e1
mkdir -p Sdcpp/sd-master-c2df4e1-bin-<your-platform>
cd Sdcpp/sd-master-c2df4e1-bin-<your-platform>
unzip /path/to/downloaded.zip
cd ../..

# 4. Install frontend dependencies
npm install --prefix frontend

# 5. Generate bindings
GOFLAGS="-tags=webkit2_41" wails generate module

# 6. Run
wails dev -tags webkit2_41
```

## Project Structure

```
├── backend/
│   ├── bridge/              # cgo interface to stable-diffusion.cpp
│   │   ├── bridge.go        # Go bindings + cgo directives
│   │   ├── bridge.h         # C-compatible struct definitions
│   │   └── bridge.cpp       # C++ wrapper for stable-diffusion.cpp API
│   └── orchestrator/
│       ├── manager.go       # State machine, worker, bound methods
│       ├── queue.go         # Ephemeral job queue (in-memory)
│       ├── history.go       # Persistent history store (~/.comfygo/history.json)
│       └── asset_handler.go # Custom Wails asset handler for image serving
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   │   ├── TopBar.svelte        # System stats (RAM/VRAM)
│   │   │   ├── LeftNav.svelte       # Navigation sidebar
│   │   │   ├── PreviewPanel.svelte  # Image preview with zoom + thumbnails
│   │   │   ├── InspectorPanel.svelte # Generation params + LoRA picker
│   │   │   ├── WorkflowStages.svelte # Real-time pipeline stages
│   │   │   ├── BottomPanel.svelte   # Queue, logs, system monitor
│   │   │   └── GalleryView.svelte   # Outputs gallery screen
│   │   └── App.svelte
├── scripts/
│   └── setup.sh              # Automated setup script
├── Sdcpp/                    # Prebuilt stable-diffusion.cpp library (gitignored)
├── .github/workflows/build.yml  # CI: builds on Linux, validates on macOS/Windows
├── Makefile                  # Build automation targets
├── main.go                   # App entry point
└── wails.json                # Wails configuration
```

## Model Placement

Place `.safetensors`, `.ckpt`, or `.gguf` model files in `~/.comfygo/models/`. LoRA files go in `~/.comfygo/loras/`. The UI dropdowns list available files on startup.

## Output

Generated images are saved to `~/.comfygo/generation/`. Generation history is persisted to `~/.comfygo/history.json`. The outputs gallery in the left nav shows all past outputs.

## CI

The GitHub Actions workflow (`.github/workflows/build.yml`) builds the full binary on Linux and validates Go code on macOS/Windows. Prebuilt stable-diffusion.cpp libraries are downloaded automatically from the upstream release.
