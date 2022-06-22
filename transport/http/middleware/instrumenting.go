package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpRequestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "application_sync",
			Name:      "http_request_totals",
			Help:      "http request count",
		},
		[]string{"url", "status"})
	httpRequestDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  "application_sync",
			Name:       "request_duration_milliseconds",
			Help:       "http request duration",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"url", "status"})
)

func init() {
	prometheus.MustRegister(httpRequestCount)
	prometheus.MustRegister(httpRequestDuration)
}

func Instrument(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func(begin time.Time) {
			iw := w.(*interceptingWriter)

			values := []string{r.URL.Path, strconv.Itoa(iw.code)}
			httpRequestCount.WithLabelValues(values...).Inc()
			timeElapsed := float64(time.Since(begin)) / float64(time.Millisecond)
			httpRequestDuration.WithLabelValues(values...).Observe(timeElapsed)
		}(time.Now())

		next.ServeHTTP(w, r)
	})
}
