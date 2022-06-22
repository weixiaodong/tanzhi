package db

import (
	"context"
	"time"
)

type JobRecord struct {
	Id           int       `db:"id" json:"id"`
	Name         string    `db:"name" json:"name"`
	Type         string    `db:"type" json:"type"`
	Command      string    `db:"command" json:"command"`
	Result       string    `db:"result" json:"result"`               // 记录任务结果
	FailedCnt    uint32    `db:"failed_cnt" json:"failed_cnt"`       // 保存任务失败次数
	CreateTime   time.Time `db:"create_time" json:"create_time"`     // 任务创建时间
	StartedTime  time.Time `db:"started_time" json:"started_time"`   // 任务开始执行时间
	FinishedTime time.Time `db:"finished_time" json:"finished_time"` // 任务结束时间
}

func InsertJobRecord(ctx context.Context, t *JobRecord) (err error) {

	sql1 := `INSERT INTO t_job_record (name, type, command, result, failed_cnt, create_time, started_time, finished_time)
			VALUES (:name, :type, :command, :result, :failed_cnt, :create_time, :started_time, :finished_time)`
	_, err = db.NamedExec(sql1, t)

	return
}

func ListJobRecord(ctx context.Context, page, size int) (res []*JobRecord, err error) {
	sql1 := `
		SELECT *
		FROM t_job_record
		ORDER BY create_time desc
		Limit $1, $2
	`
	err = db.Select(&res, sql1, page, size)
	return
}

func GetJobRecordResult(ctx context.Context, id int) (res string, err error) {
	sql1 := `
		SELECT result
		FROM t_job_record
		WHERE id = $1
	`
	err = db.Get(&res, sql1, id)
	return
}
