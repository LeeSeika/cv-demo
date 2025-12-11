## Grab Task

#### 背景

在分布式任务处理系统中，多个工作节点（Worker）需要从数据库中抓取任务进行处理。为了避免多个节点同时抓取同一个任务，通常需要一种机制来确保任务的唯一性分配。

#### 实现

通过 `子查询` 和 `RETURNING` 子句，我们可以在一条 UPDATE 语句中实现空闲任务的抓取和状态更新，由单条 SQL 语句的原子性确保任务的唯一性分配。

``` go
func (s *TaskService) GrabTask(ctx context.Context, workerID int) (*Task, error) {
	db := s.db.WithContext(ctx)

	grabbedTask := Task{
		WorkerID: workerID,
		Status:   TaskStatusRunning,
	}

	subQuery := db.Model(&Task{}).
		Where("status = ? AND worker_id = ?", TaskStatusPending, 0).
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
```

#### 索引设计

``` go
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
```

为了优化查询性能，我们为 Task 表创建了联合索引 `idx_status_worker`，涵盖 `status` 和 `worker_id` 两个字段。<br>
对于联合索引先后顺序的设计，我们选择将 `status` 字段放在前面。虽然 `status` 和 `worker_id` 的基数都较低，但是 `status` 的数据分布有一个特点，那就是随着时间的推移，`pending` 状态的任务数量占比会逐渐减少，所有任务最终都会趋向于 `completed` 状态。但是 `worker_id` 字段的分布则相对均匀。因此，将 `status` 放在前面可以更有效地过滤掉大部分非 `pending` 状态的任务，从而提高查询效率。