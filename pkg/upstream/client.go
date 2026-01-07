package upstream

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/869413421/transit/pkg/logger"
	"go.uber.org/zap"
)

// Client 上游API客户端
type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

// NewClient 创建上游API客户端
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		baseURL: baseURL,
		apiKey:  apiKey,
	}
}

// Request 通用请求结构
type Request struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    interface{}
}

// Response 通用响应结构
type Response struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

// Do 执行HTTP请求
func (c *Client) Do(ctx context.Context, req *Request) (*Response, error) {
	// 构建请求URL
	url := c.baseURL + req.Path

	// 序列化请求体
	var bodyReader io.Reader
	if req.Body != nil {
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create http request: %w", err)
	}

	// 设置默认请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	// 设置自定义请求头
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// 发送请求
	logger.Debug("Sending upstream request",
		zap.String("method", req.Method),
		zap.String("url", url),
	)

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send http request: %w", err)
	}
	defer httpResp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	logger.Debug("Received upstream response",
		zap.Int("status_code", httpResp.StatusCode),
		zap.Int("body_size", len(respBody)),
	)

	return &Response{
		StatusCode: httpResp.StatusCode,
		Body:       respBody,
		Headers:    httpResp.Header,
	}, nil
}
