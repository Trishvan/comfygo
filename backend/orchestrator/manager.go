package orchestrator

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/comfygo/backend/bridge"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type GenerationState int

const (
	StateIdle      GenerationState = iota
	StateLoading   GenerationState = 1
	StateGenerating                 = 2
	StateComplete                   = 3
	StateError                      = 4
)

func (s GenerationState) String() string {
	switch s {
	case StateIdle:
		return "idle"
	case StateLoading:
		return "loading"
	case StateGenerating:
		return "generating"
	case StateComplete:
		return "complete"
	case StateError:
		return "error"
	default:
		return "unknown"
	}
}

type GenerationParams struct {
	Prompt         string   `json:"prompt"`
	NegativePrompt string   `json:"negativePrompt"`
	ModelPath      string   `json:"modelPath"`
	VaePath        string   `json:"vaePath"`
	Steps          int      `json:"steps"`
	CfgScale       float64  `json:"cfgScale"`
	Seed           int      `json:"seed"`
	Width          int      `json:"width"`
	Height         int      `json:"height"`
	SamplerName    string   `json:"samplerName"`
	LoraPaths      []string `json:"loraPaths"`
	LoraScales     []float64 `json:"loraScales"`
}

type Manager struct {
	ctx          context.Context
	mu           sync.Mutex
	state        GenerationState
	progress     float64
	jobChan      chan GenerationParams
	cancelChan   chan struct{}
	AssetHandler *ImageAssetHandler

	resultData []byte
	resultW    int
	resultH    int
}

type SystemStats struct {
	RAMTotalGB  float64 `json:"ramTotalGB"`
	RAMUsedGB   float64 `json:"ramUsedGB"`
	RAMPercent  float64 `json:"ramPercent"`
	VRAMTotalGB float64 `json:"vramTotalGB"`
	VRAMUsedGB  float64 `json:"vramUsedGB"`
	VRAMPercent float64 `json:"vramPercent"`
}

func getSystemStats() SystemStats {
	var stats SystemStats

	// RAM from /proc/meminfo
	if data, err := os.ReadFile("/proc/meminfo"); err == nil {
		re := regexp.MustCompile(`(?m)^MemTotal:\s+(\d+)\s+kB`)
		if m := re.FindStringSubmatch(string(data)); len(m) > 1 {
			if v, err := strconv.ParseFloat(m[1], 64); err == nil {
				stats.RAMTotalGB = v / 1024 / 1024
			}
		}
		re = regexp.MustCompile(`(?m)^MemAvailable:\s+(\d+)\s+kB`)
		if m := re.FindStringSubmatch(string(data)); len(m) > 1 {
			if v, err := strconv.ParseFloat(m[1], 64); err == nil {
				stats.RAMUsedGB = stats.RAMTotalGB - v/1024/1024
			}
		}
		if stats.RAMTotalGB > 0 {
			stats.RAMPercent = (stats.RAMUsedGB / stats.RAMTotalGB) * 100
		}
	}

	// VRAM from nvidia-smi
	if out, err := exec.Command("nvidia-smi", "--query-gpu=memory.used,memory.total", "--format=csv,noheader,nounits").Output(); err == nil {
		parts := strings.Fields(string(out))
		if len(parts) >= 2 {
			if used, err := strconv.ParseFloat(parts[0], 64); err == nil {
				stats.VRAMUsedGB = used / 1024
			}
			if total, err := strconv.ParseFloat(parts[1], 64); err == nil {
				stats.VRAMTotalGB = total / 1024
			}
			if stats.VRAMTotalGB > 0 {
				stats.VRAMPercent = (stats.VRAMUsedGB / stats.VRAMTotalGB) * 100
			}
		}
	}

	return stats
}

func NewManager() *Manager {
	m := &Manager{
		state:      StateIdle,
		jobChan:    make(chan GenerationParams, 1),
		cancelChan: make(chan struct{}, 1),
	}
	m.AssetHandler = NewImageAssetHandler()
	go m.worker()
	return m
}

func (m *Manager) Start(ctx context.Context) {
	m.ctx = ctx
}

func (m *Manager) worker() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	for params := range m.jobChan {
		select {
		case <-m.cancelChan:
			m.setState(StateIdle)
			continue
		default:
		}

		m.setState(StateLoading)
		handle, err := bridge.LoadModel(params.ModelPath, params.VaePath)
		if err != nil {
			m.setError(err)
			continue
		}

		select {
		case <-m.cancelChan:
			bridge.FreeModel(handle)
			m.setState(StateIdle)
			continue
		default:
		}

		m.setState(StateGenerating)

		sanitizedPrompt := SanitizePromptWeights(params.Prompt)
		sanitizedNeg := SanitizePromptWeights(params.NegativePrompt)

		cfg := bridge.GenerationConfig{
			Prompt:         sanitizedPrompt,
			NegativePrompt: sanitizedNeg,
			ModelPath:      params.ModelPath,
			VaePath:        params.VaePath,
			Steps:          params.Steps,
			CfgScale:       params.CfgScale,
			Seed:           params.Seed,
			Width:          params.Width,
			Height:         params.Height,
			SamplerName:    params.SamplerName,
			LoraPaths:      params.LoraPaths,
			LoraScales:     params.LoraScales,
		}

		cb := func(step, total int) {
			m.setProgress(float64(step) / float64(total))
			m.emit("progress", step, total)
		}

		result, err := bridge.Txt2Img(handle, cfg, cb)
		if err != nil {
			bridge.FreeModel(handle)
			m.setError(err)
			continue
		}

		bridge.FreeModel(handle)

		w := result.Width
		h := result.Height
		ch := result.Channels

		// Copy raw pixel data into Go-managed memory
		raw := make([]byte, len(result.Data))
		copy(raw, result.Data)
		bridge.FreeImage(&result)

		// Encode raw pixels to PNG
		pngBytes, err := encodePNG(raw, w, h, ch)
		if err != nil {
			m.setError(fmt.Errorf("encode PNG: %w", err))
			continue
		}

		// Sanity check: save to ~/.comfygo/generation/
		if savePath, err := saveImage(pngBytes, params); err != nil {
			fmt.Printf("WARN: failed to save image: %v\n", err)
		} else {
			fmt.Printf("INFO: image saved to %s\n", savePath)
		}

		m.mu.Lock()
		m.resultData = pngBytes
		m.resultW = w
		m.resultH = h
		m.mu.Unlock()

		m.AssetHandler.SetImage(pngBytes, w, h)

		m.setState(StateComplete)
		m.setProgress(1.0)
		m.emit("generation-complete", map[string]int{
			"width":  w,
			"height": h,
		})
	}
}

func (m *Manager) Generate(params GenerationParams) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.state != StateIdle && m.state != StateComplete {
		return fmt.Errorf("cannot generate in state: %s", m.state)
	}

	if params.Width == 0 {
		params.Width = 512
	}
	if params.Height == 0 {
		params.Height = 512
	}
	if params.Steps <= 0 {
		params.Steps = 20
	}
	if params.CfgScale <= 0 {
		params.CfgScale = 7.0
	}
	if params.SamplerName == "" {
		params.SamplerName = "euler_a"
	}

	go func() {
		m.jobChan <- params
	}()
	return nil
}

func (m *Manager) Cancel() {
	select {
	case m.cancelChan <- struct{}{}:
	default:
	}
}

func (m *Manager) GetState() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.state.String()
}

func (m *Manager) GetProgress() float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.progress
}

func (m *Manager) GetImageData() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.resultData) == 0 {
		return ""
	}
	return base64.StdEncoding.EncodeToString(m.resultData)
}

func (m *Manager) GetSystemStats() SystemStats {
	return getSystemStats()
}

func (m *Manager) ListLoras() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return []string{}
	}
	dir := filepath.Join(home, ".comfygo", "loras")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return []string{}
	}
	var loras []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if ext == ".safetensors" || ext == ".ckpt" || ext == ".pt" || ext == ".pth" {
			loras = append(loras, filepath.Join(dir, e.Name()))
		}
	}
	return loras
}

func (m *Manager) ListModels() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return []string{}
	}
	dir := filepath.Join(home, ".comfygo", "models")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return []string{}
	}
	var models []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if ext == ".safetensors" || ext == ".ckpt" || ext == ".pt" || ext == ".pth" || ext == ".gguf" {
			models = append(models, filepath.Join(dir, e.Name()))
		}
	}
	return models
}

func (m *Manager) emit(event string, data ...interface{}) {
	if m.ctx != nil {
		wailsRuntime.EventsEmit(m.ctx, event, data...)
	}
}

func (m *Manager) setState(s GenerationState) {
	m.mu.Lock()
	m.state = s
	m.mu.Unlock()
	m.emit("state-change", s.String())
}

func (m *Manager) setProgress(p float64) {
	m.mu.Lock()
	m.progress = p
	m.mu.Unlock()
}

func (m *Manager) setError(err error) {
	m.mu.Lock()
	m.state = StateError
	m.mu.Unlock()
	m.emit("error", err.Error())
}

func encodePNG(data []byte, w, h, ch int) ([]byte, error) {
	if ch == 4 {
		img := &image.RGBA{
			Pix:    data,
			Stride: 4 * w,
			Rect:   image.Rect(0, 0, w, h),
		}
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}

	// Assume 3 channels (RGB) — convert to RGBA
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			srcOff := (y*w + x) * 3
			dstOff := (y*w + x) * 4
			img.Pix[dstOff+0] = data[srcOff+0]
			img.Pix[dstOff+1] = data[srcOff+1]
			img.Pix[dstOff+2] = data[srcOff+2]
			img.Pix[dstOff+3] = 255
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func saveImage(data []byte, params GenerationParams) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".comfygo", "generation")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	ts := time.Now().Format("20060102_150405")
	path := filepath.Join(dir, fmt.Sprintf("%s_%dx%d.png", ts, params.Width, params.Height))
	return path, os.WriteFile(path, data, 0644)
}

func SanitizePromptWeights(prompt string) string {
	if prompt == "" {
		return prompt
	}

	re := regexp.MustCompile(`\(([^:()]+):([0-9.]+)\)`)
	sanitized := re.ReplaceAllStringFunc(prompt, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}
		token := strings.TrimSpace(parts[1])
		return fmt.Sprintf("(%s)", token)
	})

	return sanitized
}
