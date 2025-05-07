package withretry

import (
	"context"
	"fmt"
	"time"

	"github.com/hellobchain/taskscheduler/task"
)

// WithRetry 带重试的任务包装器
type WithRetry struct {
	Task    task.Task
	Max     int           // 最大重试次数
	Backoff time.Duration // 退避时间
}

func (r *WithRetry) ID() string {
	return r.Task.ID()
}

func (r *WithRetry) Execute(ctx context.Context) (interface{}, error) {
	var lastErr error
	for i := 0; i < r.Max; i++ {
		if i > 0 {
			select {
			case <-time.After(r.Backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		output, err := r.Task.Execute(ctx)
		if err == nil {
			return output, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("after %d attempts: %w", r.Max, lastErr)
}
