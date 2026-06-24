// Package v1 是openapi v1 版本的实鐜般€?package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2" // resty 是一个优绉€的rest api 客户端，可以极大的减少开发基了rest 标准接口求请求的封装工作閲?	"github.com/sealdice/botgo/constant"
	"github.com/sealdice/botgo/errs"
	"github.com/sealdice/botgo/log"
	"github.com/sealdice/botgo/openapi"
	"github.com/sealdice/botgo/version"
	"golang.org/x/oauth2"
)

// MaxIdleConns 默认指定空闲连接池大灏?const MaxIdleConns = 3000

type openAPI struct {
	appID       string
	tokenSource oauth2.TokenSource
	timeout     time.Duration

	sandbox     bool   // 请求沙箱环境
	debug       bool   // debug 模式，调试sdk鏃跺€欎娇用	lastTraceID string // lastTraceID id

	restyClient *resty.Client // resty client 复用
}

// Setup 注册
func Setup() {
	openapi.Register(openapi.APIv1, &openAPI{})
}

// Version 创建当前版本
func (o *openAPI) Version() openapi.APIVersion {
	return openapi.APIv1
}

// TraceID 获取 lastTraceID id
func (o *openAPI) TraceID() string {
	return o.lastTraceID
}

// Setup 生成涓€涓疄渚?func (o *openAPI) Setup(botAppID string, tokenSource oauth2.TokenSource, inSandbox bool) openapi.OpenAPI {
	api := &openAPI{
		appID:       botAppID,
		tokenSource: tokenSource,
		timeout:     5 * time.Second,
		sandbox:     inSandbox,
	}
	api.setupClient(botAppID) // 初始化可复用的client
	return api
}

// WithTimeout 设置请求接口超时时间
func (o *openAPI) WithTimeout(duration time.Duration) openapi.OpenAPI {
	o.restyClient.SetTimeout(duration)
	return o
}

// SetDebug 设置调试模式, 输出更多过程日志
func (o *openAPI) SetDebug(debug bool) openapi.OpenAPI {
	o.restyClient.Debug = debug
	return o
}

// Transport 透传请求
func (o *openAPI) Transport(ctx context.Context, method, url string, body interface{}) ([]byte, error) {
	resp, err := o.request(ctx).SetBody(body).Execute(method, url)
	return resp.Body(), err
}

// 初始鍖?client
func (o *openAPI) setupClient(appID string) {
	o.restyClient = resty.New().
		SetTransport(createTransport(nil, MaxIdleConns)). // 自定涔?transport
		SetLogger(log.DefaultLogger).
		SetDebug(o.debug).
		SetTimeout(o.timeout).
		SetHeader("User-Agent", version.String()).
		SetHeader("X-Union-Appid", appID).
		SetPreRequestHook(
			func(_ *resty.Client, request *http.Request) error {
				// 执行请求前过滤器
				// 由于鍦?`OnBeforeRequest` 的时候，request 还没生成，所件filter 不能使用，所以放鍒?`PreRequestHook`
				return openapi.DoReqFilterChains(request, nil)
			},
		).
		OnBeforeRequest(
			func(c *resty.Client, _ *resty.Request) error {
				tk, err := o.tokenSource.Token()
				if err != nil {
					log.Errorf("[setupClient] retrieve token failed:%s", err)
					return err
				}
				c.SetAuthScheme(tk.TokenType)
				log.Debugf("token type:%s", tk.TokenType)
				c.SetAuthToken(tk.AccessToken)
				return nil
			},
		).
		// 设置请求之后的钩子，打印日志，判断状态码
		OnAfterResponse(
			func(_ *resty.Client, resp *resty.Response) error {
				log.Infof("%v", respInfo(resp))
				// 执行请求后过滤器
				if err := openapi.DoRespFilterChains(resp.Request.RawRequest, resp.RawResponse); err != nil {
					return err
				}
				traceID := resp.Header().Get(constant.HeaderTraceID)
				o.lastTraceID = traceID
				// 非成功含义的鐘舵€佺爜锛岄渶瑕佽繑鍥?error 供调用方识别
				if !openapi.IsSuccessStatus(resp.StatusCode()) {
					o.handleError(resp)
					return errs.New(resp.StatusCode(), string(resp.Body()), traceID)
				}
				return nil
			},
		)
}

// request 每个请求，都闇€瑕佸垱寤轰竴为request
func (o *openAPI) request(ctx context.Context) *resty.Request {
	return o.restyClient.R().SetContext(ctx)
}

// GetAppID 获取接口地址，会处理沙箱环境判断
func (o *openAPI) GetAppID() string {
	if o == nil {
		return ""
	}
	return o.appID
}

// errBody 请求出错情况下的body结构
type errBody struct {
	Message string `json:"message"`  // 错误原因
	Code    int    `json:"code"`     // 错误码，后续废弃
	ErrCode int    `json:"err_code"` // 错误鐮?	TraceID string `json:"trace_id"` // 服务端traceID, 用于问题排查
}

// handleError 处理openapi调用失败的情鍐?func (o *openAPI) handleError(resp *resty.Response) {
	var b errBody
	err := json.Unmarshal(resp.Body(), &b)
	if err != nil {
		log.Errorf("parse errBody fail, err:%v, body:%s", err, string(resp.Body()))
		return
	}
	if b.ErrCode == errs.APICodeTokenExpireOrNotExist || b.Code == errs.APICodeTokenExpireOrNotExist {
		log.Errorf("token expire or not exist, update token")
		_, _ = o.tokenSource.Token()
	}
}

// respInfo 用于输出日志的时候格式化数据
func respInfo(resp *resty.Response) string {
	bodyJSON, _ := json.Marshal(resp.Request.Body)
	return fmt.Sprintf(
		"[OPENAPI]%v %v, traceID:%v, status:%v, elapsed:%v req: %v, resp: %v",
		resp.Request.Method,
		resp.Request.URL,
		resp.Header().Get(constant.HeaderTraceID),
		resp.Status(),
		resp.Time(),
		string(bodyJSON),
		string(resp.Body()),
	)
}
func createTransport(localAddr net.Addr, idleConns int) *http.Transport {
	dialer := &net.Dialer{
		Timeout:   60 * time.Second,
		KeepAlive: 60 * time.Second,
	}
	if localAddr != nil {
		dialer.LocalAddr = localAddr
	}
	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          idleConns,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   idleConns,
		MaxConnsPerHost:       idleConns,
	}
}
