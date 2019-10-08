package mhttp

import (
	"bytes"
	"github.com/owenliang/myf-go-cat/message"
	"io"
	"net/http"
	"net/url"
)

// 请求封装
type Request struct {
	RawReq *http.Request	// 原始请求
	Config *HttpConfig	// 请求级配置
	isMulti bool // 是否为并发请求
	tran message.Transactor // cat埋点
}

// 应答封装
type Response struct {
	RawResp *http.Response	// 原始应答
	Body []byte	// 应答体
	Err error	// 错误
}

// 创建GET请求
func NewGetRequest(reqUrl string, query *url.Values, config *HttpConfig) (req *Request, err error) {
	// 构造请求
	var rawReq *http.Request
	if rawReq, err = http.NewRequest("GET", reqUrl, nil); err != nil {
		return
	}
	if query != nil {
		rawReq.URL.RawQuery = query.Encode()
	}
	req = &Request{
		RawReq: rawReq,
		Config: config,
	}
	return
}

// 创建POST表单请求
func NewPostRequest(reqUrl string, query *url.Values, form *url.Values, config *HttpConfig) (req *Request, err error) {
	// POST表单
	var formData io.Reader = nil
	if form != nil {
		formData = bytes.NewBufferString(form.Encode())
	}

	// 构造请求
	var rawReq *http.Request
	if rawReq, err = http.NewRequest("POST", reqUrl, formData); err != nil {
		return
	}
	if query != nil {
		rawReq.URL.RawQuery = query.Encode()
	}
	rawReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	req = &Request{
		RawReq: rawReq,
		Config: config,
	}
	return
}

// 创建POST JSON请求
func NewPostJsonRequest(reqUrl string, json []byte, config *HttpConfig) (req *Request, err error) {
	// 构造请求
	var rawReq *http.Request
	if rawReq, err = http.NewRequest("POST", reqUrl, bytes.NewBuffer(json)); err != nil {
		return
	}
	rawReq.Header.Add("Content-Type", "application/json")
	req = &Request{
		RawReq: rawReq,
		Config: config,
	}
	return
}