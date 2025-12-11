package grabtask

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGrabTask(t *testing.T) {
	// initialize SQLite
	tmpDir := t.TempDir()
	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_busy_timeout=5000", filepath.Join(tmpDir, "tasks.db"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	if err := db.AutoMigrate(&Task{}); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}
	if err := seedTasks(db); err != nil {
		t.Fatalf("failed to seed tasks: %v", err)
	}

	// concurrent workers grabbing tasks
	var (
		ctx = t.Context()
		svc = NewTaskService(db)
		wg  = sync.WaitGroup{}

		mu      = sync.Mutex{}
		taskSet = make(map[int]struct{})
	)

	workerFunc := func(ctx context.Context, svc *TaskService, workerID int) {
		defer wg.Done()
		for {
			task, err := svc.GrabTask(ctx, workerID)
			if errors.Is(err, ErrNoTaskAvailable) {
				return
			}
			if err != nil {
				t.Errorf("worker %d: failed to grab task: %v", workerID, err)
				return
			}

			mu.Lock()
			if _, exists := taskSet[int(task.ID)]; exists {
				// grabbing duplicated task is not allowed
				t.Errorf("worker %d: grabbed a duplicated task ID %d", workerID, task.ID)
			} else {
				taskSet[int(task.ID)] = struct{}{}
			}
			mu.Unlock()
		}
	}

	numWorkers := 5
	wg.Add(numWorkers)
	for i := range numWorkers {
		go workerFunc(ctx, svc, i+1)
	}
	wg.Wait()

	if len(taskSet) != 100 {
		t.Fatalf("expected to process 100 unique tasks, but got %d", len(taskSet))
	}
}

func seedTasks(db *gorm.DB) error {
	tasks := make([]*Task, 0, 100)
	for range 100 {
		tasks = append(tasks, &Task{
			Status:   TaskStatusPending,
			WorkerID: 0,
		})
	}
	return db.Create(&tasks).Error
}
