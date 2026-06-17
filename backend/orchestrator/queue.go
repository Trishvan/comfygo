package orchestrator

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type JobStatus string

const (
	JobQueued     JobStatus = "queued"
	JobRunning    JobStatus = "running"
	JobCompleted  JobStatus = "completed"
	JobFailed     JobStatus = "failed"
	JobCancelled  JobStatus = "cancelled"
	JobInterrupted JobStatus = "interrupted"
)

type QueueItem struct {
	ID        int             `json:"id"`
	Params    GenerationParams `json:"params"`
	Status    JobStatus        `json:"status"`
	Progress  float64          `json:"progress"`
	OutputPath string          `json:"outputPath"`
	Error     string           `json:"error"`
	CreatedAt string           `json:"createdAt"`
}

type JobQueue struct {
	mu          sync.Mutex
	Items       []*QueueItem `json:"items"`
	nextID      int          `json:"nextID"`
	persistPath string
	dirty       bool
}

func NewJobQueue(persistPath string) *JobQueue {
	if persistPath == "" {
		home, _ := os.UserHomeDir()
		persistPath = filepath.Join(home, ".comfygo", "queue.json")
	}
	q := &JobQueue{
		persistPath: persistPath,
	}
	q.load()
	if q.Items == nil {
		q.Items = []*QueueItem{}
	}
	return q
}

func (q *JobQueue) Enqueue(params GenerationParams) *QueueItem {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.nextID++
	item := &QueueItem{
		ID:        q.nextID,
		Params:    params,
		Status:    JobQueued,
		Progress:  0,
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	q.Items = append(q.Items, item)
	q.dirty = true
	q.persist()
	return item
}

func (q *JobQueue) NextQueued() *QueueItem {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, item := range q.Items {
		if item.Status == JobQueued {
			item.Status = JobRunning
			item.Progress = 0
			q.dirty = true
			q.persist()
			return item
		}
	}
	return nil
}

func (q *JobQueue) CompleteJob(id int, outputPath string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, item := range q.Items {
		if item.ID == id {
			item.Status = JobCompleted
			item.Progress = 1.0
			item.OutputPath = outputPath
			q.dirty = true
			q.persist()
			return
		}
	}
}

func (q *JobQueue) FailJob(id int, errMsg string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, item := range q.Items {
		if item.ID == id {
			item.Status = JobFailed
			item.Error = errMsg
			q.dirty = true
			q.persist()
			return
		}
	}
}

func (q *JobQueue) CancelJob(id int) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, item := range q.Items {
		if item.ID == id {
			if item.Status == JobQueued {
				item.Status = JobCancelled
				q.dirty = true
				q.persist()
			}
			return
		}
	}
}

func (q *JobQueue) HasRunning() bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, item := range q.Items {
		if item.Status == JobRunning {
			return true
		}
	}
	return false
}

func (q *JobQueue) RunningItem() *QueueItem {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, item := range q.Items {
		if item.Status == JobRunning {
			return item
		}
	}
	return nil
}

func (q *JobQueue) Reorder(from, to int) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if from < 0 || from >= len(q.Items) || to < 0 || to >= len(q.Items) || from == to {
		return
	}
	item := q.Items[from]
	q.Items = append(q.Items[:from], q.Items[from+1:]...)
	q.Items = append(q.Items[:to], append([]*QueueItem{item}, q.Items[to:]...)...)
	q.dirty = true
	q.persist()
}

func (q *JobQueue) ClearCompleted() {
	q.mu.Lock()
	defer q.mu.Unlock()

	filtered := make([]*QueueItem, 0, len(q.Items))
	for _, item := range q.Items {
		if item.Status == JobQueued || item.Status == JobRunning {
			filtered = append(filtered, item)
		}
	}
	q.Items = filtered
	q.dirty = true
	q.persist()
}

func (q *JobQueue) RetryJob(id int) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, item := range q.Items {
		if item.ID == id {
			switch item.Status {
			case JobCompleted, JobFailed, JobCancelled, JobInterrupted:
				item.Status = JobQueued
				item.Progress = 0
				item.Error = ""
				q.dirty = true
				q.persist()
			}
			return
		}
	}
}

func (q *JobQueue) Snapshot() []*QueueItem {
	q.mu.Lock()
	defer q.mu.Unlock()

	snapshot := make([]*QueueItem, len(q.Items))
	for i, item := range q.Items {
		copy := *item
		snapshot[i] = &copy
	}
	return snapshot
}

func (q *JobQueue) HasQueued() bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, item := range q.Items {
		if item.Status == JobQueued {
			return true
		}
	}
	return false
}

func (q *JobQueue) QueueLen() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.Items)
}

func (q *JobQueue) RunningCount() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	count := 0
	for _, item := range q.Items {
		if item.Status == JobRunning {
			count++
		}
	}
	return count
}

func (q *JobQueue) Progress() float64 {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, item := range q.Items {
		if item.Status == JobRunning {
			return item.Progress
		}
	}
	return 0
}

func (q *JobQueue) SetProgress(id int, p float64) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, item := range q.Items {
		if item.ID == id {
			item.Progress = p
			return
		}
	}
}

func (q *JobQueue) MarkInterrupted() {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, item := range q.Items {
		if item.Status == JobRunning {
			item.Status = JobInterrupted
		}
	}
	q.dirty = true
	q.persist()
}

func (q *JobQueue) persist() {
	if !q.dirty {
		return
	}
	dir := filepath.Dir(q.persistPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return
	}
	data, err := json.MarshalIndent(q, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(q.persistPath, data, 0644)
	q.dirty = false
}

func (q *JobQueue) load() {
	data, err := os.ReadFile(q.persistPath)
	if err != nil {
		return
	}
	json.Unmarshal(data, q)
	q.MarkInterrupted()
}
