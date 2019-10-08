package mhttp

import (
	"context"
	"github.com/owenliang/myf-go-cat/cat"
	"github.com/owenliang/myf-go-cat/message"
	"github.com/owenliang/myf-go/app/middlewares/metadata"
	cat2 "github.com/owenliang/myf-go/client/cat"
	"github.com/owenliang/myf-go/util"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

//默认超时时间
const DEFAULT_TIMEOUT = 3000

// 请求配置
type HttpConfig struct {
	Timeout int `toml: "timeout"` // 连接超时时间
}

// 默认http配置
var defaultHttpConfig = &HttpConfig{
	Timeout: DEFAULT_TIMEOUT,
}

//初始化http配置
func InitHttpClient(config *HttpConfig) (err error) {
	if config != nil && config.Timeout > 0 {
		defaultHttpConfig.Timeout = config.Timeout
	}
	return
}

// 设置请求配置默认值
func setDefaultHttpConfig(userConfig *HttpConfig) (mergedConfig *HttpConfig) {
	// 用户没传, 采用默认配置
	if userConfig == nil {
		mergedConfig = defaultHttpConfig
		return
	}

	// 复制用户参数, 进行默认值覆盖
	mergedConfig = &HttpConfig{}
	*mergedConfig = *userConfig
	if mergedConfig.Timeout <= 0 {
		mergedConfig.Timeout = defaultHttpConfig.Timeout
	}
	return
}

// 判断站内站外URL
func getURLType(host string) (urlType string) {
	urlType = "out"
	if strings.Contains(host, ".smmyf.com") {
		urlType = "in"
	}
	return urlType
}

// GET请求
func Get(ctx context.Context, reqUrl string, query *url.Values, config *HttpConfig) (response *Response, err error) {
	var req *Request
	if req, err = NewGetRequest(reqUrl, query, config); err != nil {
		return
	}
	return doRequest(ctx, req)
}

// POST表单
func Post(ctx context.Context, reqUrl string, query *url.Values, form *url.Values, config *HttpConfig) (response *Response, err error) {
	var req *Request
	if req, err = NewPostRequest(reqUrl, query, form, config); err != nil {
		return
	}
	return doRequest(ctx, req)
}

// POST JSON
func PostJson(ctx context.Context, reqUrl string, json []byte, config *HttpConfig) (response *Response, err error) {
	var req *Request
	if req, err = NewPostJsonRequest(reqUrl, json, config); err != nil {
		return
	}
	return doRequest(ctx, req)
}

// 并发调用
func MultiRequest(ctx context.Context, reqList []*Request) (respList []*Response, err error) {
	wg := sync.WaitGroup{}
	respList = make([]*Response, len(reqList))

	for i, _ := range respList {
		respList[i] = &Response{}
		// 关闭追踪
		reqList[i].isMulti = true
		// 协程并发
		wg.Add(1)
		go func(c context.Context, wg *sync.WaitGroup, idx int) {
			respList[idx], _ = doRequest(c, reqList[idx])
			wg.Done()
		}(ctx, &wg, i)
	}
	wg.Wait()

	// cat埋点
	{
		if myfCat, ok := ctx.Value("cat").(*cat2.MyfCat); ok {
			for _, req := range reqList {
				if req.tran != nil {
					myfCat.Append(req.tran)
					myfCat.Pop()
				}
			}
		}
	}
	return
}

// 封装网络请求(该函数返回值resp永不为空)
func doRequest(context context.Context, req *Request) (resp *Response, err error) {
	resp = &Response{}
	defer func() {
		resp.Err = err
	}()

	config := setDefaultHttpConfig(req.Config)

	client := &http.Client{
		Timeout: time.Duration(config.Timeout) * time.Millisecond, //设置默认超时时间
	}

	rawReq := req.RawReq.WithContext(context)

	{	// 请求携带链路header
		if metadata, ok := context.Value("metadata").(*metadata.MetaData); ok {
			rawReq.Header.Set("_catCallerDomain", metadata.Gin.Request.Host)
			rawReq.Header.Set("_catCallerMethod", metadata.Gin.Request.RequestURI)
		}
	}

	// Cat埋点
	{
		if err = catTraceBegin(context, req); err != nil {
			return
		}
		defer func() {
			catTraceEnd(context, req, err)
		}()
	}

	var rawResp *http.Response
	if rawResp, err = client.Do(rawReq); err != nil {
		return
	}
	defer rawResp.Body.Close()

	var respBody []byte
	if respBody, err = ioutil.ReadAll(rawResp.Body); err != nil {
		return
	}

	resp.RawResp = rawResp
	resp.Body = respBody
	return
}

// 封装cat埋点
func catTraceBegin(ctx context.Context, req *Request) (err error) {
	if _, ok := ctx.Value("cat").(*cat2.MyfCat); ok {
		reqUrl := req.RawReq.URL.String()
		var urlInfo *url.URL
		if urlInfo, err  = url.ParseRequestURI(reqUrl); err != nil {
			return
		}

		urlType := getURLType(urlInfo.Host)

		// cat上报干净链接
		parts := strings.Split(reqUrl, "?")
		cleanUrl := util.ReplaceURI(parts[0])

		var tran message.Transactor
		if urlType == "out" {
			tran = cat.NewTransaction("Curl.Outter", cleanUrl)
		} else {
			tran = cat.NewTransaction("Call", cleanUrl)
		}

		tran.LogEvent("PigeonCall.app", urlInfo.Host)
		req.tran = tran
	}
	return
}

func catTraceEnd(ctx context.Context, req *Request, err error) {
	if myfCat, ok := ctx.Value("cat").(*cat2.MyfCat); ok {
		if err != nil {
			req.tran.SetStatus(err.Error())
		}

		if req.isMulti {	// 为了并发安全, 并发调用的埋点complet延迟到主协程完成
			return
		}
		myfCat.Append(req.tran)
		myfCat.Pop()
	}
}