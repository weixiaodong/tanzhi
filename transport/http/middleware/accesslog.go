package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/weixiaodong/tanzhi/component/log"
)

type (
	additionInfoKey int
	AdditionInfo    map[string]interface{}
)

func (a AdditionInfo) String() string {
	if len(a) == 0 {
		return ""
	}
	b, _ := json.Marshal(a)
	return string(b)
}

var (
	CtxAdditionInfoKey additionInfoKey
)

func AccessLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		iw := w.(*interceptingWriter)

		bodyBytes, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

		defer func(begin time.Time) {
			log.Info(r.Context(),
				"accesslog_http",
				"host", strings.Split(r.Host, ":")[0],
				"clientip", strings.Split(r.RemoteAddr, ":")[0],
				"request_method", r.Method,
				"request_url", r.RequestURI,
				"status", iw.code,
				"http_user_agent", r.UserAgent(),
				"request_time", float64(time.Since(begin))/float64(time.Second),
				"request_param", string(bodyBytes),
				"addition_info", r.Context().Value(CtxAdditionInfoKey))
		}(time.Now())
		ctx := context.WithValue(r.Context(), CtxAdditionInfoKey, AdditionInfo{})
		r = r.WithContext(ctx)
		next.ServeHTTP(iw, r)
	})
}
