package orchestrator

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type HistoryEntry struct {
	ID      int              `json:"id"`
	Params  GenerationParams `json:"params"`
	Status  JobStatus        `json:"status"`
	Filename string          `json:"filename"`
	Error   string           `json:"error"`
	Width   int              `json:"width"`
	Height  int              `json:"height"`
	CreatedAt string         `json:"createdAt"`
}

type HistoryStore struct {
	mu     sync.Mutex
	Items  []HistoryEntry `json:"items"`
	nextID int
	path   string
}

func NewHistoryStore(path string) *HistoryStore {
	s := &HistoryStore{
		Items: []HistoryEntry{},
		path:  path,
	}
	s.load()
	return s
}

func (s *HistoryStore) Add(params GenerationParams, status JobStatus, filename string, width, height int, errMsg string) HistoryEntry {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	entry := HistoryEntry{
		ID:        s.nextID,
		Params:    params,
		Status:    status,
		Filename:  filename,
		Error:     errMsg,
		Width:     width,
		Height:    height,
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	s.Items = append(s.Items, entry)
	s.persist()
	return entry
}

func (s *HistoryStore) Snapshot() []HistoryEntry {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]HistoryEntry, len(s.Items))
	copy(result, s.Items)
	return result
}

func (s *HistoryStore) persist() {
	if s.path == "" {
		return
	}
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(s.path, data, 0644)
}

func (s *HistoryStore) load() {
	if s.path == "" {
		return
	}
	data, err := os.ReadFile(s.path)
	if err != nil {
		return
	}
	json.Unmarshal(data, s)
	if s.Items == nil {
		s.Items = []HistoryEntry{}
	}
}
