package middleware

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/weixiaodong/tanzhi/component/log"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 每个请求到来时增加recover处理
		defer func() {
			if err := recover(); err != nil {

				w.WriteHeader(http.StatusInternalServerError)
				rsp := make(map[string]interface{}, 4)
				rsp["code"] = -10001
				rsp["message"] = "服务器开小差了，请稍后再试"
				rsp["currentTime"] = time.Now().Unix()
				rsp["data"] = make(map[string]interface{})
				json.NewEncoder(w).Encode(rsp)

				log.Stack(r.Context(), "panic", "err", err)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
