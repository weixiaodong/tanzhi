package handler

import (
	"context"
	"time"

	"github.com/weixiaodong/tanzhi/component/log"
	"github.com/weixiaodong/tanzhi/internal/job/executor"
	"github.com/weixiaodong/tanzhi/internal/job/queue"
)

type cornJob struct {
	Name   string
	Type   string
	Method string
	Target string

	Ctx context.Context
}

func (s cornJob) Run() {
	log.Print(s.Ctx, "arrived_job", "name", s.Name, "time", time.Now().Format("2006-01-02 15:04:05"), "cornJob", s)

	switch s.Type {
	case executor.JobTypeHTTP:
		t := executor.NewHTTPJob(s.Name, s.Type, s.Method, s.Target)
		queue.Putq(t)
	case executor.JobTypeShell:
		t := executor.NewShellJob(s.Name, s.Type, s.Target)
		queue.Putq(t)
	default:
	}
}

func NewCornJob(ctx context.Context, name, typ, method, target string) *cornJob {
	return &cornJob{
		Name:   name,
		Type:   typ,
		Method: method,
		Target: target,
		Ctx:    ctx,
	}
}
