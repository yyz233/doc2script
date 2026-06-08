package pkg

import (
	"sync"
	"time"
)

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
)

type Task struct {
	ID        string     `json:"id"`
	Status    TaskStatus `json:"status"`
	Progress  int        `json:"progress"`
	Total     int        `json:"total"`
	Message   string     `json:"message,omitempty"`
	SaveDir   string     `json:"save_dir,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type TaskStore struct {
	mu    sync.RWMutex
	tasks map[string]*Task
}

func NewTaskStore() *TaskStore {
	return &TaskStore{tasks: make(map[string]*Task)}
}

func (ts *TaskStore) Create(id string, total int) *Task {
	t := &Task{
		ID:        id,
		Status:    TaskStatusPending,
		Total:     total,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	ts.mu.Lock()
	ts.tasks[id] = t
	ts.mu.Unlock()
	return t
}

func (ts *TaskStore) Get(id string) *Task {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.tasks[id]
}

func (ts *TaskStore) UpdateStatus(id string, status TaskStatus, progress int, message string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if t, ok := ts.tasks[id]; ok {
		t.Status = status
		t.Progress = progress
		if message != "" {
			t.Message = message
		}
		t.UpdatedAt = time.Now()
	}
}

func (ts *TaskStore) SetSaveDir(id string, saveDir string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if t, ok := ts.tasks[id]; ok {
		t.SaveDir = saveDir
		t.UpdatedAt = time.Now()
	}
}
