package api

import (
	"github.com/weixiaodong/tanzhi/internal/service"
	"github.com/weixiaodong/tanzhi/transport/http/endpoint"
)

func init() {
	s := &service.Service{}
	endpoint.RegisterHttpEndpoint("/createJob", s.CreateJob)
	endpoint.RegisterHttpEndpoint("/listJobRecord", s.ListJobRecord)
	endpoint.RegisterHttpEndpoint("/getJobRecordResult", s.GetJobRecordResult)

}
