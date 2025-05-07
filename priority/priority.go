package priority

import "github.com/hellobchain/taskscheduler/task"

// PriorityTask 带优先级的任务
type PriorityTask struct {
	Task     task.Task
	Priority int // 优先级，数字越大优先级越高
}

// PriorityQueue 优先队列实现
type PriorityQueue []*PriorityTask

func (pq PriorityQueue) Len() int { return len(pq) }

// Less 比较两个任务优先级
func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].Priority > pq[j].Priority
}

// Swap 交换两个任务
func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

// Push 添加任务
func (pq *PriorityQueue) Push(x interface{}) {
	*pq = append(*pq, x.(*PriorityTask))
}

// Pop 删除任务
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}
