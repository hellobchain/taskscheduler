package scheduler

import (
	"context"
	"fmt"
	"sync"

	"github.com/hellobchain/taskscheduler/task"
	"golang.org/x/time/rate"
)

// Scheduler 任务调度器
type Scheduler struct {
	workerNum   int            // 工作协程数
	rateLimiter *rate.Limiter  // 速率限制器
	taskQueue   chan task.Task // 任务队列
	ResultChan  chan *Result   // 结果通道
	ErrorChan   chan *Error    // 错误通道
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

// Result 任务执行结果
type Result struct {
	TaskID   string
	Output   interface{}
	Attempts int
}

// Error 任务执行错误
type Error struct {
	TaskID   string
	Err      error
	Attempts int
}

// NewScheduler 创建新调度器
func NewScheduler(workerNum, queueSize int) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		workerNum:  workerNum,
		taskQueue:  make(chan task.Task, queueSize),
		ResultChan: make(chan *Result, queueSize),
		ErrorChan:  make(chan *Error, queueSize),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// SetRateLimit 设置每秒最大任务数
func (s *Scheduler) SetRateLimit(perSecond int) {
	s.rateLimiter = rate.NewLimiter(rate.Limit(perSecond), perSecond)
}

// Submit 提交任务
func (s *Scheduler) Submit(task task.Task) {
	s.taskQueue <- task
}

// Start 启动调度器
func (s *Scheduler) Start() {
	for i := 0; i < s.workerNum; i++ {
		s.wg.Add(1)
		go s.worker()
	}
}

// worker 工作协程
func (s *Scheduler) worker() {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		case task := <-s.taskQueue:
			// 速率限制
			if s.rateLimiter != nil {
				if err := s.rateLimiter.Wait(s.ctx); err != nil {
					s.ErrorChan <- &Error{
						TaskID: task.ID(),
						Err:    fmt.Errorf("rate limit wait failed: %w", err),
					}
					continue
				}
			}
			// 执行任务
			output, err := task.Execute(s.ctx)
			if err != nil {
				s.ErrorChan <- &Error{
					TaskID: task.ID(),
					Err:    err,
				}
			} else {
				s.ResultChan <- &Result{
					TaskID: task.ID(),
					Output: output,
				}
			}
		}
	}
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	s.cancel()
	s.wg.Wait()
	close(s.taskQueue)
	close(s.ResultChan)
	close(s.ErrorChan)
}

// AdjustWorkers 调整worker数量
func (s *Scheduler) AdjustWorkers(newNum int) {
	if newNum > s.workerNum {
		// 增加worker
		for i := s.workerNum; i < newNum; i++ {
			s.wg.Add(1)
			go s.worker()
		}
	} else if newNum < s.workerNum {
		// 减少worker (通过context取消)
		s.workerNum = newNum
	}
}

// Wait 等待所有任务完成
func (s *Scheduler) Wait() {
	s.wg.Wait()
}
