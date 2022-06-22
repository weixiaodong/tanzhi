package endpoint

import (
	"context"
	"errors"
	"reflect"

	"github.com/go-kit/kit/endpoint"
	"github.com/mitchellh/mapstructure"

	"github.com/weixiaodong/tanzhi/transport/http/middleware"
	"github.com/weixiaodong/tanzhi/transport/http/session"
)

type beforeEndpointFunc func(next endpoint.Endpoint) endpoint.Endpoint

var (
	Endpoints = make(map[string]endpoint.Endpoint)
)

func RegisterHttpEndpoint(pattern string, handler interface{}) {
	// 从 handler 获取信息
	fn := makeHttpEndpoint(handler)
	fn = WithHttpCloseEndpoint(fn)
	fn = WithResponseEndpoint(fn)
	Endpoints[pattern] = fn
}

func makeHttpEndpoint(handler interface{}) endpoint.Endpoint {
	fv := reflect.ValueOf(handler)
	fn := func(ctx context.Context, request interface{}) (response interface{}, err error) {
		params := make([]reflect.Value, 2)
		params[0] = reflect.ValueOf(ctx)
		// mapstructure 解码map值成Go结构体
		v := reflect.New(reflect.TypeOf(handler).In(1)).Interface()

		var (
			decoder *mapstructure.Decoder
			cfg     = &mapstructure.DecoderConfig{
				WeaklyTypedInput: true,
				Result:           v,
				TagName:          "json",
			}
		)
		decoder, err = mapstructure.NewDecoder(cfg)
		if err == nil {
			err = decoder.Decode(request)
		}
		if err != nil {
			panic("decode_request_error")
		}
		params[1] = reflect.ValueOf(v).Elem()
		rs := fv.Call(params)
		response = rs[0].Interface()
		err, _ = rs[1].Interface().(error)
		return
	}
	return fn
}

var (
	ClientHasDisconnected = errors.New("client disconnected")
)

func WithHttpCloseEndpoint(next endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (rsp interface{}, err error) {
		rsp, err = next(ctx, request)
		select {
		case <-ctx.Done():
			return nil, ClientHasDisconnected
		default:
			return
		}
	}
}

func WithResponseEndpoint(next endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (rsp interface{}, err error) {
		rsp, err = next(ctx, request)
		ss := session.FromContext(ctx)
		rawRequest := ss.GetRawRequest()
		if l, ok := rawRequest.Context().Value(middleware.CtxAdditionInfoKey).(middleware.AdditionInfo); ok {
			l["data"] = rsp
			if err != nil {
				l["err"] = err.Error()
			}
		}
		return
	}
}
