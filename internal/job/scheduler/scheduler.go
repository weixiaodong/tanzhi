package scheduler

import (
	"context"
	"time"

	"github.com/weixiaodong/tanzhi/component/log"
	job "github.com/weixiaodong/tanzhi/internal/job/queue"
)

// 启动10个goroutine获取任务执行
func Start() {
	ctx := context.Background()
	go startQloop(ctx)
}

func startQloop(ctx context.Context) {
	for {
		for !job.IsEmpty() {
			t := job.Popq()
			log.Info(ctx, "execed_job", "job", t)

			// 执行任务
			err := t.Execed(ctx)
			if err != nil {
				log.Error(ctx, "execed_job_failed", "err", err)
				t.Failed()
				if t.IsRetry() {
					job.Putq(t)
				}
				continue
			}
			// 判断是否需要重试
			if t.IsRetry() {
				job.Putq(t)
			}

		}
		// 暂时没有任务了，休息一会
		time.Sleep(1 * time.Second)
	}
}
