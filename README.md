# ComfyGo

Desktop application for Stable Diffusion 1.5 image generation. Go backend with Svelte frontend, powered by [stable-diffusion.cpp](https://github.com/leejet/stable-diffusion.cpp).

## Prerequisites

- Go 1.23+
- Node 20+
- **Linux:** `sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.1-dev`
- **macOS:** Xcode Command Line Tools (`xcode-select --install`)
- **Windows:** WebView2 runtime (included in Windows 10+)

## Setup

### 1. Get the prebuilt stable-diffusion.cpp library

Download the appropriate prebuilt binary for your platform from the [releases page](https://github.com/leejet/stable-diffusion.cpp/releases/tag/master-700-c2df4e1):

| Platform | Download |
|----------|----------|
| Linux (x86_64, Vulkan) | `sd-master-c2df4e1-bin-Linux-Ubuntu-24.04-x86_64-vulkan.zip` |
| Linux (x86_64, CPU) | `sd-master-c2df4e1-bin-Linux-Ubuntu-24.04-x86_64.zip` |
| macOS (ARM64) | `sd-master-c2df4e1-bin-Darwin-macOS-15.7.7-arm64.zip` |
| Windows (x86_64, Vulkan) | `sd-master-c2df4e1-bin-win-vulkan-x64.zip` |

Extract into `Sdcpp/`:

```bash
mkdir -p Sdcpp/sd-master-c2df4e1-bin-Linux-Ubuntu-24.04-x86_64-vulkan
cd Sdcpp/sd-master-c2df4e1-bin-Linux-Ubuntu-24.04-x86_64-vulkan
unzip /path/to/downloaded.zip
```

> The directory name encodes the branch, commit, OS, arch, and backend. CI downloads these automatically from the release.

### 2. Install dependencies

```bash
npm install --prefix frontend
```

### 3. Copy your model

```bash
mkdir -p ~/.comfygo/models
cp your-model.safetensors ~/.comfygo/models/
```

### 4. Run in dev mode

```bash
wails dev -tags webkit2_41
```

## Build

```bash
wails build -tags webkit2_41 -o comfygo
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
├── Sdcpp/                   # Prebuilt stable-diffusion.cpp library (gitignored)
├── .github/workflows/build.yml  # CI: builds on all 3 platforms
├── main.go                  # App entry point
└── wails.json               # Wails configuration
```

## CI

The GitHub Actions workflow builds on Linux (full binary) and validates on macOS/Windows (go vet). Prebuilt stable-diffusion.cpp libraries are downloaded automatically from the upstream release.

## Model Placement

Place `.safetensors`, `.ckpt`, or `.gguf` model files in `~/.comfygo/models/`. LoRA files go in `~/.comfygo/loras/`. The UI dropdowns will list available files on startup.

## Output

Generated images are saved to `~/.comfygo/generation/`. Generation history is persisted to `~/.comfygo/history.json`. The gallery tab in the left nav shows all past outputs.
