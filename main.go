package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"math/rand"

	"github.com/hellobchain/taskscheduler/scheduler"
	"github.com/hellobchain/taskscheduler/task"
	"github.com/hellobchain/taskscheduler/withretry"
	"github.com/hellobchain/wswlog/wlogging"
)

var logger = wlogging.MustGetFileLoggerWithoutName(nil)

func main() {
	// 创建调度器：3个工作协程，队列大小100
	scheduler := scheduler.NewScheduler(3, 100)
	// 设置每秒最多处理5个任务
	scheduler.SetRateLimit(5)
	// 启动调度器
	scheduler.Start()
	// 收集结果
	go func() {
		for result := range scheduler.ResultChan {
			logger.Infof("Task %s completed: %v\n", result.TaskID, result.Output)
		}
	}()
	// 处理错误
	go func() {
		for err := range scheduler.ErrorChan {
			logger.Errorf("Task %s failed: %v (attempts: %d)\n",
				err.TaskID, err.Err, err.Attempts)
		}
	}()
	// 提交任务
	for i := 0; i < 20; i++ {
		taskID := fmt.Sprintf("task-%d", i)
		task := &task.SingleTask{
			Id: taskID,
			Action: func() (interface{}, error) {
				// 模拟任务执行
				time.Sleep(time.Millisecond * 100)
				if rand.Intn(10) == 0 { // 10%失败率
					return nil, fmt.Errorf("random error")
				}
				return fmt.Sprintf("result of %s", taskID), nil
			},
		}
		// 包装为带重试的任务
		retryTask := &withretry.WithRetry{
			Task:    task,
			Max:     3,
			Backoff: time.Millisecond * 200,
		}
		scheduler.Submit(retryTask)
	}
	// 信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	// 优雅停机
	scheduler.Stop()
	scheduler.Wait()
}
