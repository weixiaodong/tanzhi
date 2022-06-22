package executor

import (
	"context"
	"fmt"
	"time"
)

type job struct {
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	CreateAt   time.Time `json:"createAt"`
	StartedAt  time.Time `json:"startedAt"`
	FinishedAt time.Time `json:"finishedAt"`
	FailedCnt  uint32    `json:"failedCnt"`
	Retry      uint32    `json:"retry"`
	err        error
}

func (t *job) GetJobName() string {
	return t.Name
}

func (t *job) GetJobType() string {
	return t.Type
}

func (t *job) GetFailedCnt() uint32 {
	return t.FailedCnt
}

func (t *job) GetRetry() uint32 {
	return t.Retry
}

func (t *job) String() string {
	return fmt.Sprintf("job{name=%s,type=%s}", t.Name, t.Type)
}

func (t *job) Execed(ctx context.Context) error {
	return nil
}

func (t *job) Failed() {
	t.FailedCnt++
}

func (t *job) IsRetry() bool {
	if t.err == nil {
		return false
	}
	// 失败次数大于0并且小于等于重试次数时去重试
	if t.FailedCnt > 0 && t.FailedCnt < t.Retry {
		return true
	}
	return false
}
