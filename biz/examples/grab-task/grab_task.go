package grabtask

import (
	"context"
	"errors"

	"github.com/leeseika/cv-demo/pkg/model/object"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrNoTaskAvailable = errors.New("no task available")

type TaskService struct {
	db *gorm.DB
}

func NewTaskService(db *gorm.DB) *TaskService {
	return &TaskService{db: db}
}

func (s *TaskService) GrabTask(ctx context.Context, workerID int) (*object.Task, error) {
	db := s.db.WithContext(ctx)

	grabbedTask := object.Task{
		WorkerID: workerID,
		Status:   object.TaskStatusRunning,
	}

	subQuery := db.Model(&object.Task{}).
		Where("status = ? AND worker_id = ?", object.TaskStatusPending, 0).
		Limit(1).
		Select("id")

	result := db.Model(&grabbedTask).
		Clauses(clause.Returning{}).
		Where("id IN (?)", subQuery).
		Updates(&grabbedTask)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, ErrNoTaskAvailable
	}

	return &grabbedTask, nil
}
