package timeout

import (
	"context"
	"time"

	"github.com/hellobchain/taskscheduler/task"
)

// WithTimeout 带超时控制的任务包装器
type WithTimeout struct {
	Task    task.Task
	Timeout time.Duration
}

func (t *WithTimeout) ID() string {
	return t.Task.ID()
}

func (t *WithTimeout) Execute(ctx context.Context) (interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, t.Timeout)
	defer cancel()
	return t.Task.Execute(ctx)
}
