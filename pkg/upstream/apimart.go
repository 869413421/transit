package upstream

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/869413421/transit/pkg/logger"
	"go.uber.org/zap"
)

// APIMartAdapter APIMart适配器
type APIMartAdapter struct {
	client *Client
}

// NewAPIMartAdapter 创建APIMart适配器
func NewAPIMartAdapter(baseURL, apiKey string) *APIMartAdapter {
	return &APIMartAdapter{
		client: NewClient(baseURL, apiKey),
	}
}

// ChatCompletion 文本对话(同步)
func (a *APIMartAdapter) ChatCompletion(ctx context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	logger.Info("APIMart chat completion request",
		zap.String("model", req.Model),
		zap.Int("messages", len(req.Messages)),
	)

	// 发送请求到APIMart
	resp, err := a.client.Do(ctx, &Request{
		Method: http.MethodPost,
		Path:   "/v1/chat/completions",
		Body:   req,
	})
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(resp.Body, &errResp); err != nil {
			return nil, fmt.Errorf("http %d: %s", resp.StatusCode, string(resp.Body))
		}
		return nil, fmt.Errorf("api error: %s", errResp.Error.Message)
	}

	// 解析响应
	var chatResp ChatCompletionResponse
	if err := json.Unmarshal(resp.Body, &chatResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	logger.Info("APIMart chat completion success",
		zap.String("model", chatResp.Model),
		zap.Int("total_tokens", chatResp.Usage.TotalTokens),
	)

	return &chatResp, nil
}

// ImageGeneration 图片生成(异步)
func (a *APIMartAdapter) ImageGeneration(ctx context.Context, req *ImageGenerationRequest) (*ImageGenerationResponse, error) {
	logger.Info("APIMart image generation request",
		zap.String("model", req.Model),
		zap.String("prompt", req.Prompt[:min(50, len(req.Prompt))]),
	)

	// 发送请求到APIMart
	resp, err := a.client.Do(ctx, &Request{
		Method: http.MethodPost,
		Path:   "/v1/images/generations",
		Body:   req,
	})
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(resp.Body, &errResp); err != nil {
			return nil, fmt.Errorf("http %d: %s", resp.StatusCode, string(resp.Body))
		}
		return nil, fmt.Errorf("api error: %s", errResp.Error.Message)
	}

	// 解析响应(APIMart返回格式: {code: 0, data: {status: "submitted", task_id: "xxx"}})
	var apiResp struct {
		Code int `json:"code"`
		Data struct {
			Status string `json:"status"`
			TaskID string `json:"task_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body, &apiResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	logger.Info("APIMart image generation submitted",
		zap.String("task_id", apiResp.Data.TaskID),
		zap.String("status", apiResp.Data.Status),
	)

	return &ImageGenerationResponse{
		TaskID: apiResp.Data.TaskID,
		Status: apiResp.Data.Status,
	}, nil
}

// VideoGeneration 视频生成(异步)
func (a *APIMartAdapter) VideoGeneration(ctx context.Context, req *VideoGenerationRequest) (*VideoGenerationResponse, error) {
	logger.Info("APIMart video generation request",
		zap.String("model", req.Model),
		zap.String("prompt", req.Prompt[:min(50, len(req.Prompt))]),
	)

	// 发送请求到APIMart
	resp, err := a.client.Do(ctx, &Request{
		Method: http.MethodPost,
		Path:   "/v1/videos/generations",
		Body:   req,
	})
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(resp.Body, &errResp); err != nil {
			return nil, fmt.Errorf("http %d: %s", resp.StatusCode, string(resp.Body))
		}
		return nil, fmt.Errorf("api error: %s", errResp.Error.Message)
	}

	// 解析响应
	var apiResp struct {
		Code int `json:"code"`
		Data struct {
			Status string `json:"status"`
			TaskID string `json:"task_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body, &apiResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	logger.Info("APIMart video generation submitted",
		zap.String("task_id", apiResp.Data.TaskID),
		zap.String("status", apiResp.Data.Status),
	)

	return &VideoGenerationResponse{
		TaskID: apiResp.Data.TaskID,
		Status: apiResp.Data.Status,
	}, nil
}

// GetTaskStatus 查询任务状态
func (a *APIMartAdapter) GetTaskStatus(ctx context.Context, taskID string) (*TaskStatusResponse, error) {
	logger.Debug("APIMart get task status", zap.String("task_id", taskID))

	// 发送请求到APIMart
	resp, err := a.client.Do(ctx, &Request{
		Method: http.MethodGet,
		Path:   fmt.Sprintf("/v1/tasks/%s", taskID),
	})
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, string(resp.Body))
	}

	// 解析响应
	var taskResp TaskStatusResponse
	if err := json.Unmarshal(resp.Body, &taskResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	logger.Debug("APIMart task status",
		zap.String("task_id", taskID),
		zap.String("status", taskResp.Status),
	)

	return &taskResp, nil
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
