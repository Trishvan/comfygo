package orchestrator

import (
	"context"
	"fmt"
	"regexp"
	"runtime"
	"strings"
	"sync"

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
	Prompt         string  `json:"prompt"`
	NegativePrompt string  `json:"negativePrompt"`
	ModelPath      string  `json:"modelPath"`
	VaePath        string  `json:"vaePath"`
	Steps          int     `json:"steps"`
	CfgScale       float64 `json:"cfgScale"`
	Seed           int     `json:"seed"`
	Width          int     `json:"width"`
	Height         int     `json:"height"`
	SamplerName    string  `json:"samplerName"`
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
		handle, err := bridge.LoadModel(params.ModelPath)
		if err != nil {
			m.setError(err)
			continue
		}

		if params.VaePath != "" {
			if err := bridge.LoadVae(handle, params.VaePath); err != nil {
				bridge.FreeModel(handle)
				m.setError(err)
				continue
			}
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

		m.mu.Lock()
		oldData := m.resultData
		m.resultData = result.Data
		m.resultW = result.Width
		m.resultH = result.Height
		m.mu.Unlock()

		if oldData != nil {
			bridge.FreeImage(&bridge.ImageResult{Data: oldData})
		}

		m.AssetHandler.SetImage(result.Data, result.Width, result.Height)

		m.setState(StateComplete)
		m.setProgress(1.0)
		m.emit("generation-complete", map[string]int{
			"width":  result.Width,
			"height": result.Height,
		})
	}
}

func (m *Manager) Generate(params GenerationParams) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.state != StateIdle {
		return fmt.Errorf("generation already in progress (state: %s)", m.state)
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
