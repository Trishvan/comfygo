package orchestrator

import (
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
)

type QueueItem struct {
	ID         int              `json:"id"`
	Params     GenerationParams `json:"params"`
	Status     JobStatus        `json:"status"`
	Progress   float64          `json:"progress"`
	OutputPath string           `json:"outputPath"`
	Error      string           `json:"error"`
	CreatedAt  string           `json:"createdAt"`
}

type JobQueue struct {
	mu     sync.Mutex
	Items  []*QueueItem `json:"items"`
	nextID int
}

func NewJobQueue() *JobQueue {
	return &JobQueue{
		Items: []*QueueItem{},
	}
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
	return item
}

func (q *JobQueue) NextQueued() *QueueItem {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, item := range q.Items {
		if item.Status == JobQueued {
			item.Status = JobRunning
			item.Progress = 0
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
}

func (q *JobQueue) RetryJob(id int) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, item := range q.Items {
		if item.ID == id {
			switch item.Status {
			case JobCompleted, JobFailed, JobCancelled:
				item.Status = JobQueued
				item.Progress = 0
				item.Error = ""
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
