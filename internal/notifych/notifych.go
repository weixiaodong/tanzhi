package notifych

import (
	"context"

	"github.com/weixiaodong/tanzhi/component/log"
	"github.com/weixiaodong/tanzhi/config"
)

var (

	// 新建任务通道
	JobCreateCh = make(chan config.Job, 100)
)

func NotifyJobCreate(ctx context.Context, job config.Job) {
	select {
	case JobCreateCh <- job:
	default:
		log.Warning(ctx, "JobCreateChFull", "ignore job", job)
	}
}
