package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// 设置默认传输配置，保持长链
func setDefaultTransport() {
	t := http.DefaultTransport.(*http.Transport)
	// 最大2000
	t.MaxIdleConns = 2000

	// 每个ip 100
	t.MaxIdleConnsPerHost = 200

	// 空闲超时 30s
	t.IdleConnTimeout = 30 * time.Second

}

type httpResult struct {
	Status int    `json:"status"`
	Body   string `json:"body"`
}

func (h *httpResult) Encode() string {
	if h == nil {
		return ""
	}
	b, _ := json.Marshal(h)
	return string(b)
}

// 注入请求操作，如http 头注入
type InjectRequest func(*http.Request)

// Send 发送http请求
func Send(ctx context.Context, method string, url string, reqBody map[string]interface{}, irs ...InjectRequest) (result *httpResult, err error) {

	req, err := newRequest(ctx, method, url, reqBody)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// 请求头注入
	for _, o := range irs {
		o(req)
	}

	// 创建客户端并发送请求
	client := &http.Client{
		Timeout: 1 * time.Second, // 设置超时时间
	}
	var (
		bodyBytes []byte
	)
	rsp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	bodyBytes, err = ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}

	result = &httpResult{
		Status: rsp.StatusCode,
		Body:   string(bodyBytes),
	}
	return result, nil
}

func newRequest(ctx context.Context, method string, path string, in map[string]interface{}) (req *http.Request, err error) {
	// 编码请求
	if strings.ToUpper(method) == "GET" {
		req, err = http.NewRequest(method, path, nil)
		if err != nil {
			return
		}
		params := make(url.Values)
		for k, v := range in {
			params.Add(k, fmt.Sprint(v))
		}
		req.URL.RawQuery = params.Encode()
	} else {
		var b []byte
		if len(in) > 0 {
			b, _ = json.Marshal(in)
		}
		req, err = http.NewRequest(method, path, bytes.NewReader(b))
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}

	return
}
