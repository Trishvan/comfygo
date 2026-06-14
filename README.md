# ComfyGo

Desktop application for Stable Diffusion 1.5 image generation. Go backend with Svelte frontend, powered by [stable-diffusion.cpp](https://github.com/leejet/stable-diffusion.cpp).

## Prerequisites

- Go 1.23+
- Node 20+
- Linux: `sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.1-dev`
- Prebuilt `libstable-diffusion.so` in `Sdcpp/` (see below)

## Quick Start

```bash
# Install dependencies
npm install --prefix frontend

# Copy your model
mkdir -p ~/.comfygo/models
cp your-model.safetensors ~/.comfygo/models/

# Generate Wails bindings
GOFLAGS="-tags=webkit2_41" wails generate module

# Run in dev mode
wails dev -tags webkit2_41
```

## Build

```bash
wails build -tags webkit2_41
```

## Project Structure

```
├── backend/
│   ├── bridge/          # cgo interface to stable-diffusion.cpp
│   └── orchestrator/    # State machine, queue, asset serving
├── frontend/            # Svelte 4 + TypeScript UI
├── main.go              # App entry point
└── wails.json           # Wails configuration
```

## Model Placement

Place `.safetensors`, `.ckpt`, or `.gguf` model files in `~/.comfygo/models/`. The model selector dropdown will list available models on startup.
