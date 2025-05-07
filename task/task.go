package task

import (
	"context"
)

// Task 表示一个可执行的任务单元
type Task interface {
	// ID 返回任务唯一标识
	ID() string

	// Execute 执行任务并返回结果或错误
	Execute(ctx context.Context) (interface{}, error)
}

// SingleTask 任务实现
type SingleTask struct {
	Id     string
	Action func() (interface{}, error)
}

// ID 返回任务唯一标识
func (t *SingleTask) ID() string {
	return t.Id
}

// Execute 执行任务并返回结果或错误
func (t *SingleTask) Execute(ctx context.Context) (interface{}, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		if t.Action == nil {
			return nil, nil
		}
		return t.Action()
	}
}

// BatchTask 任务组
type BatchTask struct {
	tasks []Task
}

// ID 返回任务组唯一标识
func (b *BatchTask) Execute(ctx context.Context) (interface{}, error) {
	results := make([]interface{}, len(b.tasks))
	for i, task := range b.tasks {
		out, err := task.Execute(ctx)
		if err != nil {
			return nil, err
		}
		results[i] = out
	}
	return results, nil
}
