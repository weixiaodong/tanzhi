package innerhandler

import (
	"net/http"

	"github.com/weixiaodong/tanzhi/transport/http/middleware"
)

// HandlePing 处理ping请求
func HandlePing(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func Handle404() http.Handler {
	// Handle 404
	h := http.NotFound
	return wrapHandler(http.HandlerFunc(h))
}

func wrapHandler(handler http.Handler) http.Handler {
	h := middleware.AccessLogging(handler)
	h = middleware.Instrument(h)
	h = middleware.Mix(h)

	return h
}
