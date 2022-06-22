package middleware

import (
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/weixiaodong/tanzhi/component/log"
)

func Trace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 每个请求到来时增加trace id
		traceId := strings.ReplaceAll(uuid.New().String(), "-", "")
		ctx := r.Context()
		ctx = log.WithTraceId(ctx, traceId)
		ctx = log.WithModule(ctx, "http_service")

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
