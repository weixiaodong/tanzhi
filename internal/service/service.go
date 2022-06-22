package service

import (
	"context"
	"encoding/json"

	"github.com/weixiaodong/tanzhi/config"
	"github.com/weixiaodong/tanzhi/ecode"
	"github.com/weixiaodong/tanzhi/internal/notifych"
	"github.com/weixiaodong/tanzhi/internal/store/db"
)

type Service struct{}

// 创建任务请求参数
type CreateJobRequest struct {
	Name    string               `json:"name"`
	Expr    string               `json:"expr"`
	Command config.CommandConfig `json:"command"`
}

// 创建任务响应
type CreateJobResponse struct {
}

func (*Service) CreateJob(ctx context.Context, in *CreateJobRequest) (out *CreateJobResponse, err error) {
	// 参数检查
	if in.Name == "" || in.Expr == "" || in.Command.Type == "" || in.Command.Target == "" {
		return nil, ecode.ParameterError
	}

	// 向通道写入任务信息
	job := config.Job{
		Name:    in.Name,
		Expr:    in.Expr,
		Command: in.Command,
	}

	notifych.NotifyJobCreate(ctx, job)
	return
}

// 获取任务执行结果
type ListJobRecordRequest struct {
	Page int `json:"page"`
	Size int `json:"size"`
}

func (*Service) ListJobRecord(ctx context.Context, in *ListJobRecordRequest) (out []*db.JobRecord, err error) {

	// 默认显示最新3条记录
	if in.Size == 0 {
		in.Size = 3
	}
	l, err := db.ListJobRecord(ctx, in.Page, in.Size)
	return l, err
}

// 获取任务执行结果
type GetJobRecordResultRequest struct {
	RecordId int `json:"recordId"`
}

// 获取任务执行结果详情
func (*Service) GetJobRecordResult(ctx context.Context, in *GetJobRecordResultRequest) (out map[string]interface{}, err error) {

	s, err := db.GetJobRecordResult(ctx, in.RecordId)
	if err == nil {
		err = json.Unmarshal([]byte(s), &out)
	}
	return out, err
}
