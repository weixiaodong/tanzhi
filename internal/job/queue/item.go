package queue

import (
	"context"
	"sync"
)

type Executor interface {
	// 获取任务名称
	GetJobName() string
	// 获取任务类型
	GetJobType() string
	String() string
	// 执行任务
	Execed(context.Context) error
	// 获取失败次数
	GetFailedCnt() uint32
	// 获取重试次数
	GetRetry() uint32
	// 设置任务失败处理
	Failed()
	// 判断任务是否需要重试
	IsRetry() bool
}

var (
	Jobq      = NewQueue(10)
	jobqMutex = sync.RWMutex{}
)

func IsEmpty() bool {
	jobqMutex.RLock()
	ok := len(Jobq) == 0
	jobqMutex.RUnlock()
	return ok
}

func Putq(t Executor) {
	jobqMutex.Lock()
	Jobq.Push(t)
	jobqMutex.Unlock()
}

func Popq() Executor {
	jobqMutex.Lock()
	t := Jobq.Pop()
	jobqMutex.Unlock()
	return t
}

func Clear() {
	jobqMutex.Lock()
	Jobq = Jobq[:0]
	jobqMutex.Unlock()
}

func Getq() Queue {
	jobqMutex.RLock()
	pq := make(Queue, len(Jobq))
	copy(pq, Jobq)
	jobqMutex.RUnlock()

	return pq
}
