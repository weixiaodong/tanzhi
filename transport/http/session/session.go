package session

import (
	"context"
	"net/http"
)

type Session interface {
	GetRawRequest() *http.Request
}

// 请求参数
type session struct {
	rawRequest *http.Request
}

func (s *session) GetRawRequest() *http.Request {
	return s.rawRequest
}

func New() Session {
	return &session{}
}

type sessionKey struct{}

// BeforeRequestFunc before request decode for save common info ctx
func BeforeRequestFunc(ctx context.Context, r *http.Request) context.Context {
	s := session{}
	s.rawRequest = r
	ctx = NewContext(ctx, &s)
	return ctx
}

func NewContext(ctx context.Context, session Session) context.Context {
	return context.WithValue(ctx, sessionKey{}, session)
}

func FromContext(ctx context.Context) Session {
	s, ok := ctx.Value(sessionKey{}).(Session)
	if !ok {
		return nil
	}

	return s
}
