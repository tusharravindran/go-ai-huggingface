package mocks

import (
	"context"
	"fmt"
	"time"

	"github.com/tusharr/go-ai-huggingface/internal/model"
)

// MockAIService is a mock implementation of model.AIService
type MockAIService struct {
	GenerateTextFunc       func(ctx context.Context, req *model.AIRequest) (*model.AIResponse, error)
	GenerateCompletionFunc func(ctx context.Context, req *model.AIRequest) (*model.AIResponse, error)
	AnalyzeSentimentFunc   func(ctx context.Context, text string) (*model.SentimentResponse, error)
	SummarizeTextFunc      func(ctx context.Context, text string, maxLength int) (*model.SummaryResponse, error)
	ValidateModelFunc      func(model string) error

	// Call tracking
	GenerateTextCalls       int
	GenerateCompletionCalls int
	AnalyzeSentimentCalls   int
	SummarizeTextCalls      int
	ValidateModelCalls      int
}

// NewMockAIService creates a new mock AI service with default implementations
func NewMockAIService() *MockAIService {
	return &MockAIService{
		GenerateTextFunc: func(ctx context.Context, req *model.AIRequest) (*model.AIResponse, error) {
			return &model.AIResponse{
				ID:           req.ID,
				Model:        req.Model,
				GeneratedAt:  time.Now(),
				ProcessingMs: 100,
				Choices: []model.Choice{
					{
						Index:        0,
						Text:         "Generated text for: " + req.Prompt,
						FinishReason: "stop",
					},
				},
				Usage: model.Usage{
					PromptTokens:     len(req.Prompt) / 4,
					CompletionTokens: 20,
					TotalTokens:      len(req.Prompt)/4 + 20,
				},
			}, nil
		},
		GenerateCompletionFunc: func(ctx context.Context, req *model.AIRequest) (*model.AIResponse, error) {
			return &model.AIResponse{
				ID:           req.ID,
				Model:        req.Model,
				GeneratedAt:  time.Now(),
				ProcessingMs: 150,
				Choices: []model.Choice{
					{
						Index:        0,
						Text:         "Completed text for: " + req.Prompt,
						FinishReason: "stop",
					},
				},
				Usage: model.Usage{
					PromptTokens:     len(req.Prompt) / 4,
					CompletionTokens: 25,
					TotalTokens:      len(req.Prompt)/4 + 25,
				},
			}, nil
		},
		AnalyzeSentimentFunc: func(ctx context.Context, text string) (*model.SentimentResponse, error) {
			sentiment := "positive"
			score := 0.8
			if len(text) < 10 {
				sentiment = "negative"
				score = 0.3
			}
			return &model.SentimentResponse{
				Text:       text,
				Sentiment:  sentiment,
				Score:      score,
				Confidence: score,
			}, nil
		},
		SummarizeTextFunc: func(ctx context.Context, text string, maxLength int) (*model.SummaryResponse, error) {
			summary := "Summary of: " + text[:min(len(text), 50)] + "..."
			if len(summary) > maxLength {
				summary = summary[:maxLength]
			}
			return &model.SummaryResponse{
				OriginalText: text,
				Summary:      summary,
				Compression:  float64(len(summary)) / float64(len(text)),
			}, nil
		},
		ValidateModelFunc: func(modelName string) error {
			if modelName == "" {
				return fmt.Errorf("model name cannot be empty")
			}
			if modelName == "invalid-model" {
				return fmt.Errorf("model not supported")
			}
			return nil
		},
	}
}

// GenerateText implements model.AIService
func (m *MockAIService) GenerateText(ctx context.Context, req *model.AIRequest) (*model.AIResponse, error) {
	m.GenerateTextCalls++
	return m.GenerateTextFunc(ctx, req)
}

// GenerateCompletion implements model.AIService
func (m *MockAIService) GenerateCompletion(ctx context.Context, req *model.AIRequest) (*model.AIResponse, error) {
	m.GenerateCompletionCalls++
	return m.GenerateCompletionFunc(ctx, req)
}

// AnalyzeSentiment implements model.AIService
func (m *MockAIService) AnalyzeSentiment(ctx context.Context, text string) (*model.SentimentResponse, error) {
	m.AnalyzeSentimentCalls++
	return m.AnalyzeSentimentFunc(ctx, text)
}

// SummarizeText implements model.AIService
func (m *MockAIService) SummarizeText(ctx context.Context, text string, maxLength int) (*model.SummaryResponse, error) {
	m.SummarizeTextCalls++
	return m.SummarizeTextFunc(ctx, text, maxLength)
}

// ValidateModel implements model.AIService
func (m *MockAIService) ValidateModel(modelName string) error {
	m.ValidateModelCalls++
	return m.ValidateModelFunc(modelName)
}

// SetGenerateTextError makes GenerateText return an error
func (m *MockAIService) SetGenerateTextError(err error) {
	m.GenerateTextFunc = func(ctx context.Context, req *model.AIRequest) (*model.AIResponse, error) {
		return nil, err
	}
}

// SetAnalyzeSentimentError makes AnalyzeSentiment return an error
func (m *MockAIService) SetAnalyzeSentimentError(err error) {
	m.AnalyzeSentimentFunc = func(ctx context.Context, text string) (*model.SentimentResponse, error) {
		return nil, err
	}
}

// SetSummarizeTextError makes SummarizeText return an error
func (m *MockAIService) SetSummarizeTextError(err error) {
	m.SummarizeTextFunc = func(ctx context.Context, text string, maxLength int) (*model.SummaryResponse, error) {
		return nil, err
	}
}

// SetValidateModelError makes ValidateModel return an error
func (m *MockAIService) SetValidateModelError(err error) {
	m.ValidateModelFunc = func(modelName string) error {
		return err
	}
}

// Reset resets all call counters
func (m *MockAIService) Reset() {
	m.GenerateTextCalls = 0
	m.GenerateCompletionCalls = 0
	m.AnalyzeSentimentCalls = 0
	m.SummarizeTextCalls = 0
	m.ValidateModelCalls = 0
}

// Helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}