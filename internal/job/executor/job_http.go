package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/weixiaodong/tanzhi/component/httpclient"
	"github.com/weixiaodong/tanzhi/component/log"
	"github.com/weixiaodong/tanzhi/internal/store/db"
)

// 处理http任务
type httpJob struct {
	job
	Method string `json:"method"`
	Target string `json:"target"`
}

func NewHTTPJob(name string, typ string, method string, target string) *httpJob {
	t := &httpJob{}
	t.Name = name
	t.Type = typ
	t.Method = method
	t.Target = target
	t.CreateAt = time.Now()
	return t
}

func (t *httpJob) String() string {
	return fmt.Sprintf("httpjob{name=%s,type=%s}", t.Name, t.Type)
}

// 发起http请求
// 不管成功或失败将结果保存到db中
// 失败后记录失败次数
func (t *httpJob) Execed(ctx context.Context) error {
	t.StartedAt = time.Now()
	result, err := httpclient.Send(ctx, t.Method, t.Target, nil)
	t.err = err // 保存执行err信息，用于后续重试判断，什么错误可用重试，什么错误不可以重试
	if err != nil {
		log.Error(ctx, "execed_job_failed", "err", err, "job", t)
		t.Failed()
	}
	// 命令执行结束时间
	t.FinishedAt = time.Now()
	// 保存结果到db中
	command, _ := json.Marshal(map[string]interface{}{
		"method": t.Method,
		"target": t.Target,
	})
	job := &db.JobRecord{
		Name:         t.Name,
		Type:         t.Type,
		Command:      string(command),
		Result:       result.Encode(),
		FailedCnt:    t.FailedCnt,
		CreateTime:   t.CreateAt,
		StartedTime:  t.StartedAt,
		FinishedTime: t.FinishedAt,
	}
	err = db.InsertJobRecord(ctx, job)
	if err != nil {
		log.Error(ctx, "insert_db_job_failed", "err", err, "job", t)
	}
	return err
}
