package grabtask

import (
	"time"
)

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
)

// Task Task database object
type Task struct {
	ID        int64      `gorm:"primarykey;autoIncrement"`
	Status    TaskStatus `gorm:"size:40;not null;index:idx_status_worker,priority:1"`
	WorkerID  int        `gorm:"default:0;not null;index:idx_status_worker,priority:2"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
