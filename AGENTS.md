COGNITIVE INSTRUCTION SET & CHECKLIST: NODE-FREE SD GENERATOR

You are an expert systems programmer specializing in low-overhead desktop integrations. Your stack is Wails (Go) + Svelte/TypeScript + stable-diffusion.cpp (C++ via cgo). Your target is a high-performance desktop application capable of running Stable Diffusion 1.5 checkpoints safely on machines constrained to 16GB RAM.

1. Core Operating Principles (Non-Negotiable)

No Python, No PyTorch: All inference is handled by stable-diffusion.cpp linked as an OS-native shared library (.dll, .so, or .dylib).

Zero-Copy Memory Boundary: Never copy C memory arrays into the Go space using helper functions like C.GoBytes(). Wrap the raw pointers immediately with Go unsafe.Slice() to point directly to C-allocated memory.

Zero-Base64 Bindings: Never return image bytes, binaries, or large Base64 strings across standard Wails JS bindings. Return metadata only. Serve the actual image bytes through a custom Go-native AssetsHandler via a customized local URI (e.g., wails://app/output.png?id=...).

Pin OS Threads: Always use runtime.LockOSThread() within Go background worker goroutines before invoking cgo inference pipelines to prevent Go runtime scheduling starvation.

Robust Memory Deallocation: C-allocated tensors, contexts, and prompt structures must be explicitly freed via free_model() or custom ctx->free() calls in a defer block in Go.

2. Workspace & Directory Structure

Organize the repository precisely as follows before generating implementation files:

├── .github/workflows/          # Cross-compilation native runners
├── frontend/                   # Svelte/TS Webview App
│   ├── src/
│   │   ├── components/         # Form-based layout, parameters, previews
│   │   ├── App.svelte          # Main entry point
│   │   └── main.ts
├── backend/
│   ├── bridge/
│   │   ├── bridge.go           # cgo interface implementation
│   │   ├── bridge.h            # C boundary header file
│   │   └── bridge.cpp          # C++ wrapper implementation for stable-diffusion.cpp
│   └── orchestrator/
│       ├── manager.go          # State machine, queue manager, asset writer
│       └── asset_handler.go    # Custom Wails Asset Handler (HTTP/Stream)
├── build/                      # Wails build configurations
├── wails.json                  # Native configuration
└── main.go                     # Entry point (App bootstrap, UI config)


3. Step-by-Step Implementation Checklist

Phase 1: C++ Shared Library & Cgo Bindings

[ ] Configure Header (bridge.h): Define the C-compatible structural interface to pass variables between Go and stable-diffusion.cpp.

[ ] Write C++ Wrapper (bridge.cpp): Wrap the native C++ API of stable-diffusion.cpp with standard C linkage (extern "C") to manage:

Model initialization (load_model)

Standalone VAE embedding overlay (load_vae)

Image generation parameters (txt2img_c)

System memory release pointers (free_model_c)

[ ] Write Go Bridge (bridge.go): Specify compiler flags (#cgo LDFLAGS, #cgo CFLAGS) linking the Go runtime to your target OS shared library. Map Go strings/ints/floats directly to their respective C.char, C.int, and C.float components.

Phase 2: Orchestration Layer & Assets Router

[ ] Create State Machine (manager.go): Build an asynchronous worker thread that locks the OS thread (runtime.LockOSThread()) during active generation loops.

[ ] Implement Token-Weight Sanitizer: Implement a Go utility to parse and transform prompt structures containing standard weight parenthesis formats (e.g., (photorealistic:1.2)) into appropriate structures recognized by stable-diffusion.cpp's internal tokenizer.

[ ] Build Custom Assets Router (asset_handler.go): Implement the http.Handler interface for Wails. When Svelte requests <img src="wails://app/render.png?id=abc" />:

Intercept request using Go.

Locate the raw image memory slice pinned from the unsafe.Slice() operation.

Stream the binary payload directly with appropriate image/png or image/jpeg content headers.

Avoid string allocations or intermediate files.

Phase 3: Svelte Frontend Layout (App Mode)

[ ] Design Form UI: Build a sidebar panel featuring numeric and slider bindings for:

Steps (1–50)

CFG Scale (1.0–20.0)

Seed (-1 for random)

Model and custom VAE file dropdowns

Prompt / Negative Prompt boxes

[ ] Incorporate Progress Tracker: Bind to custom Go events communicating current progress (e.g., Step 5/20).

[ ] Preview Throttle: Bind intermediate previews to Svelte only every 3-5 steps. Utilize a lightweight canvas target or dynamically appended image URL queries.

4. Critical Code Patterns

The Zero-Copy unsafe.Slice Wrapper (Go)

Use this pattern to consume C-allocated memory buffer pointers without copying memory into the Go garbage collector space:

package bridge

/*
#include <stdlib.h>
#include "bridge.h"
*/
import "C"
import (
	"unsafe"
)

type ImageResult struct {
	Width  int
	Height int
	Data   []byte // Directly points to C memory
	CPtr   *C.uchar
}

func GetGeneratedImage(cResult *C.image_result_t) ImageResult {
	totalBytes := int(cResult.width) * int(cResult.height) * int(cResult.channels)
	
	// Create a slice mapped directly over the C pointer
	goSlice := unsafe.Slice((*byte)(unsafe.Pointer(cResult.data)), totalBytes)
	
	return ImageResult{
		Width:  int(cResult.width),
		Height: int(cResult.height),
		Data:   goSlice,
		CPtr:   cResult.data,
	}
}


The Custom Wails Assets Handler (Go)

Configure the AssetsHandler option in Wails to bypass the JSON serialization of image data:

package orchestrator

import (
	"net/http"
	"strconv"
)

type ImageAssetHandler struct {
	ActiveBuffer []byte // Reference to active unsafe.Slice
}

func (h *ImageAssetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/render.png" {
		if h.ActiveBuffer == nil {
			http.Error(w, "No active generation", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Content-Length", strconv.Itoa(len(h.ActiveBuffer)))
		w.Write(h.ActiveBuffer)
		return
	}
	http.NotFound(w, r)
}


5. Troubleshooting & Agent Rules

Compiler Errors (cgo): If compilation fails, check that target directory dependencies are correctly set inside #cgo LDFLAGS for your dynamic linking paths. For Windows platforms, ensure MSVC library structures are matched.

Memory Leaks: If system monitor reports growing memory footprints after multiple renders:

Verify that C.free() is invoked on all strings generated with C.CString().

Ensure the custom C++ destructor wrappers for the model context and diffusion buffers are executed after the assets handler streams the last byte.

Denoising Previes look Desaturated: Verify if the target model requires an external VAE. Make sure the --vae-tiling parameter has been called during init configurations.

Wails UI Freezing: Check if the running Go thread is invoking GUI functions inside the heavy compute loops. Decouple your system status tracking using async runtime events (wails.EventsEmit).
