package ecode

import (
	"fmt"
)

type ECode struct {
	code int    `json:"code"`
	msg  string `json:"message"`
}

func (e ECode) Error() string {
	return fmt.Sprintf("code: %d, msg: %s", e.Code(), e.Message())
}

func (e ECode) Code() int { return int(e.code) }

func (e ECode) Message() string {
	return e.msg
}

func New(code int, msg string) ECode {
	return ECode{code: code, msg: msg}
}

var (
	// code 正向私有，负向共用
	Success = New(0, "success")

	ServerError             = New(-1, "服务器内部错误")
	ParameterError          = New(-2, "参数错误")
	ServerAcquireLockFailed = New(-3, "获取锁资源失败，请稍后再试")
)
