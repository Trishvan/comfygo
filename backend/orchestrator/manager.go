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
	queue        *JobQueue
	runSignal    chan struct{}
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
		state:     StateIdle,
		queue:     NewJobQueue(""),
		runSignal: make(chan struct{}, 1),
	}
	m.AssetHandler = NewImageAssetHandler()
	go m.worker()
	return m
}

func (m *Manager) Start(ctx context.Context) {
	m.ctx = ctx

	m.emitQueueUpdate()

	if m.queue.HasQueued() {
		select {
		case m.runSignal <- struct{}{}:
		default:
		}
	}
}

func (m *Manager) worker() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	for {
		item := m.queue.NextQueued()
		if item == nil {
			m.setState(StateIdle)
			<-m.runSignal
			continue
		}

		params := item.Params
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

		m.setState(StateLoading)
		m.emitQueueUpdate()

		handle, err := bridge.LoadModel(params.ModelPath, params.VaePath)
		if err != nil {
			m.queue.FailJob(item.ID, err.Error())
			m.emitQueueUpdate()
			m.setError(err)
			m.setState(StateIdle)
			continue
		}

		if m.isCancelled(item) {
			bridge.FreeModel(handle)
			continue
		}

		m.setState(StateGenerating)
		m.emitQueueUpdate()

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

		jobID := item.ID
		cb := func(step, total int) {
			p := float64(step) / float64(total)
			m.queue.SetProgress(jobID, p)
			m.emitQueueUpdate()
			m.setProgress(p)
			m.emit("progress", step, total)
		}

		result, err := bridge.Txt2Img(handle, cfg, cb)
		bridge.FreeModel(handle)

		if bridge.IsCancelInFlight() {
			bridge.ClearCancelInFlight()
			m.queue.FailJob(jobID, "cancelled")
			m.queue.SetProgress(jobID, 0)
			m.AssetHandler.ClearImage()
			m.emitQueueUpdate()
			m.setState(StateIdle)
			continue
		}

		if err != nil {
			m.queue.FailJob(jobID, err.Error())
			m.emitQueueUpdate()
			m.setError(err)
			m.setState(StateIdle)
			continue
		}

		w := result.Width
		h := result.Height
		ch := result.Channels

		raw := make([]byte, len(result.Data))
		copy(raw, result.Data)
		bridge.FreeImage(&result)

		pngBytes, err := encodePNG(raw, w, h, ch)
		if err != nil {
			m.queue.FailJob(jobID, fmt.Sprintf("encode PNG: %v", err))
			m.emitQueueUpdate()
			m.setError(fmt.Errorf("encode PNG: %w", err))
			m.setState(StateIdle)
			continue
		}

		savePath := ""
		if p, err := saveImage(pngBytes, params); err != nil {
			fmt.Printf("WARN: failed to save image: %v\n", err)
		} else {
			savePath = p
			fmt.Printf("INFO: image saved to %s\n", savePath)
		}

		m.mu.Lock()
		m.resultData = pngBytes
		m.resultW = w
		m.resultH = h
		m.mu.Unlock()

		m.AssetHandler.SetImage(pngBytes, w, h)

		m.queue.CompleteJob(jobID, savePath)
		m.emitQueueUpdate()

		m.setState(StateComplete)
		m.setProgress(1.0)
		m.emit("generation-complete", map[string]int{
			"width":  w,
			"height": h,
			"jobID":  jobID,
		})
	}
}

func (m *Manager) isCancelled(item *QueueItem) bool {
	if bridge.IsCancelInFlight() {
		bridge.ClearCancelInFlight()
		m.queue.FailJob(item.ID, "cancelled")
		m.AssetHandler.ClearImage()
		m.emitQueueUpdate()
		m.setState(StateIdle)
		return true
	}
	return false
}

func (m *Manager) EnqueueJob(params GenerationParams) {
	m.queue.Enqueue(params)
	m.emitQueueUpdate()

	select {
	case m.runSignal <- struct{}{}:
	default:
	}
}

func (m *Manager) CancelRunningJob() {
	if m.queue.RunningCount() > 0 {
		bridge.SetCancelInFlight()
	}
}

func (m *Manager) CancelJob(id int) {
	m.queue.CancelJob(id)
	m.emitQueueUpdate()
}

func (m *Manager) RetryJob(id int) {
	m.queue.RetryJob(id)
	m.emitQueueUpdate()
	select {
	case m.runSignal <- struct{}{}:
	default:
	}
}

func (m *Manager) GetQueue() []*QueueItem {
	return m.queue.Snapshot()
}

func (m *Manager) ReorderQueue(from, to int) {
	m.queue.Reorder(from, to)
	m.emitQueueUpdate()
}

func (m *Manager) ClearCompleted() {
	m.queue.ClearCompleted()
	m.emitQueueUpdate()
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

func (m *Manager) emitQueueUpdate() {
	m.emit("queue-update", m.queue.Snapshot())
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
