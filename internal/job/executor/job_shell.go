package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/weixiaodong/tanzhi/component/log"
	"github.com/weixiaodong/tanzhi/internal/store/db"
)

type shellTask struct {
	job
	Target string `json:"target"`
}

func NewShellJob(name string, typ string, target string) *shellTask {
	t := &shellTask{}
	t.Name = name
	t.Type = typ
	t.Target = target
	t.CreateAt = time.Now()
	t.Retry = 3 // 真实场景中根据配置设置重试次数
	return t
}

func (t *shellTask) String() string {
	return fmt.Sprintf("shelljob{name=%s,type=%s}", t.Name, t.Type)
}

func (t *shellTask) Execed(ctx context.Context) error {
	t.StartedAt = time.Now()

	cmd := exec.Command("/bin/sh", "-c", t.Target)
	out, err := cmd.Output()
	t.err = err
	if err != nil {
		log.Error(ctx, "execed_job_failed", "err", err, "job", t)
		t.Failed()
	}

	// 命令执行结束时间
	t.FinishedAt = time.Now()
	// 保存结果到db中
	command, _ := json.Marshal(map[string]interface{}{
		"target": t.Target,
	})

	result, _ := json.Marshal(map[string]interface{}{
		"out": string(out),
		"err": fmt.Sprint(err),
	})

	job := &db.JobRecord{
		Name:         t.Name,
		Type:         t.Type,
		Command:      string(command),
		Result:       string(result),
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
