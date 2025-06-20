package models

import "time"

type TaskStatus string

const (
	StatusPending    TaskStatus = "pending"
	StatusProcessing TaskStatus = "processing"
	StatusCompleted  TaskStatus = "completed"
	StatusFailed     TaskStatus = "failed"
	StatusCancelled  TaskStatus = "cancelled"
)

type Task struct {
	ID          string      `json:"id"`
	Status      TaskStatus  `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
	StartedAt   *time.Time  `json:"started_at,omitempty"`
	CompletedAt *time.Time  `json:"completed_at,omitempty"`
	CancelledAt *time.Time  `json:"cancelled_at,omitempty"`
	Duration    float64     `json:"duration_seconds,omitempty"`
	Result      interface{} `json:"result,omitempty"`
	Error       string      `json:"error,omitempty"`
	Description string      `json:"description,omitempty"`
}

type TaskUpdate struct {
	Description *string `json:"description,omitempty"`
}

type TaskList struct {
	Tasks []Task `json:"tasks"`
	Total int    `json:"total"`
}
