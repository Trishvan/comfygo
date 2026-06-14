package orchestrator

import (
	"net/http"
	"strconv"
	"sync"
)

type ImageAssetHandler struct {
	mu           sync.RWMutex
	ActiveBuffer []byte
	Width        int
	Height       int
}

func NewImageAssetHandler() *ImageAssetHandler {
	return &ImageAssetHandler{}
}

func (h *ImageAssetHandler) SetImage(buf []byte, width, height int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.ActiveBuffer = buf
	h.Width = width
	h.Height = height
}

func (h *ImageAssetHandler) GetImage() ([]byte, int, int) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.ActiveBuffer, h.Width, h.Height
}

func (h *ImageAssetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/render.png":
		h.serveImage(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *ImageAssetHandler) serveImage(w http.ResponseWriter, r *http.Request) {
	buf, wd, ht := h.GetImage()
	if buf == nil {
		http.Error(w, "No active generation", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(len(buf)))
	w.Header().Set("X-Image-Width", strconv.Itoa(wd))
	w.Header().Set("X-Image-Height", strconv.Itoa(ht))
	w.Write(buf)
}
