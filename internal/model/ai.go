package model

import (
	"context"
	"time"
)

// AIRequest represents a request to the AI service
type AIRequest struct {
	ID          string            `json:"id"`
	Model       string            `json:"model"`
	Prompt      string            `json:"prompt"`
	MaxTokens   int               `json:"max_tokens,omitempty"`
	Temperature float32           `json:"temperature,omitempty"`
	TopP        float32           `json:"top_p,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
}

// AIResponse represents a response from the AI service
type AIResponse struct {
	ID           string                 `json:"id"`
	Model        string                 `json:"model"`
	Choices      []Choice               `json:"choices"`
	Usage        Usage                  `json:"usage"`
	GeneratedAt  time.Time              `json:"generated_at"`
	ProcessingMs int64                  `json:"processing_ms"`
}

// Choice represents a single generated choice
type Choice struct {
	Index        int    `json:"index"`
	Text         string `json:"text"`
	FinishReason string `json:"finish_reason"`
	LogProbs     *float64 `json:"log_probs,omitempty"`
}

// Usage represents token usage statistics
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// AIService defines the interface for AI operations
type AIService interface {
	GenerateText(ctx context.Context, req *AIRequest) (*AIResponse, error)
	GenerateCompletion(ctx context.Context, req *AIRequest) (*AIResponse, error)
	AnalyzeSentiment(ctx context.Context, text string) (*SentimentResponse, error)
	SummarizeText(ctx context.Context, text string, maxLength int) (*SummaryResponse, error)
	ValidateModel(model string) error
}

// SentimentResponse represents sentiment analysis result
type SentimentResponse struct {
	Text      string  `json:"text"`
	Sentiment string  `json:"sentiment"`
	Score     float64 `json:"score"`
	Confidence float64 `json:"confidence"`
}

// SummaryResponse represents text summarization result
type SummaryResponse struct {
	OriginalText string `json:"original_text"`
	Summary      string `json:"summary"`
	Compression  float64 `json:"compression"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type"`
	Details interface{} `json:"details,omitempty"`
}

// Validate validates the AI request
func (r *AIRequest) Validate() error {
	if r.Prompt == "" {
		return &ErrorResponse{
			Code:    400,
			Message: "prompt is required",
			Type:    "validation_error",
		}
	}
	if r.Model == "" {
		return &ErrorResponse{
			Code:    400,
			Message: "model is required",
			Type:    "validation_error",
		}
	}
	if r.MaxTokens < 0 {
		return &ErrorResponse{
			Code:    400,
			Message: "max_tokens must be positive",
			Type:    "validation_error",
		}
	}
	if r.Temperature < 0 || r.Temperature > 1 {
		return &ErrorResponse{
			Code:    400,
			Message: "temperature must be between 0 and 1",
			Type:    "validation_error",
		}
	}
	return nil
}

// Error implements the error interface for ErrorResponse
func (e *ErrorResponse) Error() string {
	return e.Message
}