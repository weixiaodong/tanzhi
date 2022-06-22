package http

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"time"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"

	"github.com/weixiaodong/tanzhi/ecode"
	"github.com/weixiaodong/tanzhi/transport/http/endpoint"
	"github.com/weixiaodong/tanzhi/transport/http/innerhandler"
	"github.com/weixiaodong/tanzhi/transport/http/middleware"
	"github.com/weixiaodong/tanzhi/transport/http/session"
)

func NewHTTPHandler() http.Handler {
	// route register
	r := mux.NewRouter()
	r.HandleFunc("/ping", innerhandler.HandlePing)
	r.NotFoundHandler = innerhandler.Handle404()

	// use middleware
	r.Use(middleware.Recovery)
	r.Use(middleware.Mix)
	r.Use(middleware.Instrument)
	r.Use(middleware.Trace)
	r.Use(middleware.AccessLogging)

	opts := []kithttp.ServerOption{
		kithttp.ServerBefore(session.BeforeRequestFunc),
		kithttp.ServerErrorEncoder(encodeError),
	}

	for pattern := range endpoint.Endpoints {
		r.Handle(pattern, kithttp.NewServer(
			endpoint.Endpoints[pattern],
			decodeHTTPGenericRequest,
			encodeResponse,
			opts...,
		))
	}

	return r
}

// 解析请求参数，返回map[string]interface{}, endpoint层从map中解析参数
func decodeHTTPGenericRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	obj := map[string]interface{}{}
	if r.ContentLength > 0 {
		err = json.NewDecoder(r.Body).Decode(&obj)
	}

	return obj, err
}

type replyRsp struct {
	Code        int         `json:"code"`
	Message     string      `json:"message"`
	CurrentTime int64       `json:"currentTime"`
	Data        interface{} `json:"data"`
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	var r replyRsp
	r.Code = ecode.Success.Code()
	r.Message = ecode.Success.Message()
	r.Data = response
	if isValueNil(response) {
		r.Data = make(map[string]interface{})
	}
	r.CurrentTime = time.Now().Unix()
	return json.NewEncoder(w).Encode(r)
}

func encodeError(ctx context.Context, err error, w http.ResponseWriter) {
	if err == endpoint.ClientHasDisconnected {
		w.WriteHeader(499)
	} else {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		var r replyRsp
		switch e := err.(type) {
		case ecode.ECode:
			r.Code = e.Code()
			r.Message = e.Message()
		default:
			r.Code = 500
			r.Message = err.Error()
		}

		r.CurrentTime = time.Now().Unix()
		r.Data = make(map[string]interface{})
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(r)
	}
}

func isValueNil(response interface{}) bool {
	if response == nil {
		return true
	}
	v := reflect.ValueOf(response)
	return v.Kind() == reflect.Ptr && v.IsNil()
}
