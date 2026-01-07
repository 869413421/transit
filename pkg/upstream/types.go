package upstream

// ChatCompletionRequest 文本对话请求
type ChatCompletionRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// Message 消息
type Message struct {
	Role    string `json:"role"` // system, user, assistant
	Content string `json:"content"`
}

// ChatCompletionResponse 文本对话响应
type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice 选择项
type Choice struct {
	Index   int     `json:"index"`
	Message Message `json:"message"`
}

// Usage Token使用量
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ImageGenerationRequest 图片生成请求
type ImageGenerationRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	N      int    `json:"n,omitempty"`    // 生成数量
	Size   string `json:"size,omitempty"` // 图片尺寸
}

// ImageGenerationResponse 图片生成响应(异步)
type ImageGenerationResponse struct {
	TaskID string `json:"task_id"`
	Status string `json:"status"` // pending, processing, completed, failed
}

// VideoGenerationRequest 视频生成请求
type VideoGenerationRequest struct {
	Model    string `json:"model"`
	Prompt   string `json:"prompt"`
	Duration int    `json:"duration,omitempty"` // 视频时长(秒)
}

// VideoGenerationResponse 视频生成响应(异步)
type VideoGenerationResponse struct {
	TaskID string `json:"task_id"`
	Status string `json:"status"` // pending, processing, completed, failed
}

// TaskStatusRequest 任务状态查询请求
type TaskStatusRequest struct {
	TaskID string `json:"task_id"`
}

// TaskStatusResponse 任务状态响应
type TaskStatusResponse struct {
	TaskID    string     `json:"id"`
	Status    string     `json:"status"` // pending, processing, completed, failed, cancelled
	Progress  int        `json:"progress"`
	Result    TaskResult `json:"result,omitempty"`
	ResultURL string     `json:"result_url,omitempty"`
	Error     TaskError  `json:"error,omitempty"`
	Created   int64      `json:"created"`
	Completed int64      `json:"completed,omitempty"`
}

// TaskResult 任务结果
type TaskResult struct {
	Images []string `json:"images,omitempty"`
	Videos []string `json:"videos,omitempty"`
}

// TaskError 任务错误
type TaskError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail 错误详情
type ErrorDetail struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}
