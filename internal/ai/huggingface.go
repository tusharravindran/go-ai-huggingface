package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/tusharr/go-ai-huggingface/internal/config"
	"github.com/tusharr/go-ai-huggingface/internal/model"
	"github.com/tusharr/go-ai-huggingface/pkg/logger"
)

// HuggingFaceService implements the AIService interface using Hugging Face API
type HuggingFaceService struct {
	config     *config.HuggingFaceConfig
	httpClient *http.Client
	logger     logger.Logger
}

// HuggingFaceRequest represents a request to Hugging Face API
type HuggingFaceRequest struct {
	Inputs     string                 `json:"inputs"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Options    map[string]interface{} `json:"options,omitempty"`
}

// HuggingFaceResponse represents a response from Hugging Face API
type HuggingFaceResponse struct {
	GeneratedText string  `json:"generated_text,omitempty"`
	Score         float64 `json:"score,omitempty"`
	Label         string  `json:"label,omitempty"`
	SummaryText   string  `json:"summary_text,omitempty"`
}

// HuggingFaceError represents an error response from Hugging Face API
type HuggingFaceError struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// NewHuggingFaceService creates a new Hugging Face service instance
func NewHuggingFaceService(config *config.HuggingFaceConfig, logger logger.Logger) *HuggingFaceService {
	return &HuggingFaceService{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		logger: logger,
	}
}

// GenerateText generates text using the specified model
func (s *HuggingFaceService) GenerateText(ctx context.Context, req *model.AIRequest) (*model.AIResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	startTime := time.Now()
	s.logger.Info(ctx, "Starting text generation", map[string]interface{}{
		"request_id": req.ID,
		"model":      req.Model,
		"prompt":     req.Prompt[:min(len(req.Prompt), 100)] + "...", // Log first 100 chars
	})

	hfReq := &HuggingFaceRequest{
		Inputs: req.Prompt,
		Parameters: map[string]interface{}{
			"max_new_tokens": req.MaxTokens,
			"temperature":    req.Temperature,
			"top_p":          req.TopP,
		},
		Options: map[string]interface{}{
			"wait_for_model": true,
		},
	}

	// Merge additional parameters
	for k, v := range req.Parameters {
		hfReq.Parameters[k] = v
	}

	response, err := s.makeRequest(ctx, req.Model, hfReq)
	if err != nil {
		s.logger.Error(ctx, "Failed to generate text", map[string]interface{}{
			"request_id": req.ID,
			"error":      err.Error(),
		})
		return nil, err
	}

	processingTime := time.Since(startTime)

	// Parse response
	var hfResponses []HuggingFaceResponse
	if err := json.Unmarshal(response, &hfResponses); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(hfResponses) == 0 {
		return nil, fmt.Errorf("no response generated")
	}

	// Build AI response
	aiResponse := &model.AIResponse{
		ID:           req.ID,
		Model:        req.Model,
		GeneratedAt:  time.Now(),
		ProcessingMs: processingTime.Milliseconds(),
		Choices:      make([]model.Choice, len(hfResponses)),
	}

	totalTokens := 0
	for i, hfResp := range hfResponses {
		generatedText := hfResp.GeneratedText
		// Remove original prompt from generated text if it's included
		generatedText = strings.TrimPrefix(generatedText, req.Prompt)

		aiResponse.Choices[i] = model.Choice{
			Index:        i,
			Text:         generatedText,
			FinishReason: "stop",
		}

		// Rough token estimation (1 token ≈ 4 characters)
		totalTokens += (len(req.Prompt) + len(generatedText)) / 4
	}

	aiResponse.Usage = model.Usage{
		PromptTokens:     len(req.Prompt) / 4,
		CompletionTokens: totalTokens - (len(req.Prompt) / 4),
		TotalTokens:      totalTokens,
	}

	s.logger.Info(ctx, "Text generation completed", map[string]interface{}{
		"request_id":    req.ID,
		"processing_ms": processingTime.Milliseconds(),
		"total_tokens":  totalTokens,
	})

	return aiResponse, nil
}

// GenerateCompletion is an alias for GenerateText for compatibility
func (s *HuggingFaceService) GenerateCompletion(ctx context.Context, req *model.AIRequest) (*model.AIResponse, error) {
	return s.GenerateText(ctx, req)
}

// AnalyzeSentiment analyzes sentiment of the given text
func (s *HuggingFaceService) AnalyzeSentiment(ctx context.Context, text string) (*model.SentimentResponse, error) {
	s.logger.Info(ctx, "Starting sentiment analysis", map[string]interface{}{
		"text_length": len(text),
	})

	hfReq := &HuggingFaceRequest{
		Inputs: text,
	}

	// Use a sentiment analysis model
	modelName := "cardiffnlp/twitter-roberta-base-sentiment-latest"
	response, err := s.makeRequest(ctx, modelName, hfReq)
	if err != nil {
		return nil, err
	}

	var hfResponses [][]HuggingFaceResponse
	if err := json.Unmarshal(response, &hfResponses); err != nil {
		return nil, fmt.Errorf("failed to parse sentiment response: %w", err)
	}

	if len(hfResponses) == 0 || len(hfResponses[0]) == 0 {
		return nil, fmt.Errorf("no sentiment analysis result")
	}

	// Find the highest score sentiment
	bestResult := hfResponses[0][0]
	for _, result := range hfResponses[0] {
		if result.Score > bestResult.Score {
			bestResult = result
		}
	}

	return &model.SentimentResponse{
		Text:       text,
		Sentiment:  bestResult.Label,
		Score:      bestResult.Score,
		Confidence: bestResult.Score,
	}, nil
}

// SummarizeText summarizes the given text
func (s *HuggingFaceService) SummarizeText(ctx context.Context, text string, maxLength int) (*model.SummaryResponse, error) {
	s.logger.Info(ctx, "Starting text summarization", map[string]interface{}{
		"text_length": len(text),
		"max_length":  maxLength,
	})

	hfReq := &HuggingFaceRequest{
		Inputs: text,
		Parameters: map[string]interface{}{
			"max_length": maxLength,
			"min_length": maxLength / 4, // Min length is 25% of max
		},
	}

	// Use a summarization model
	modelName := "facebook/bart-large-cnn"
	response, err := s.makeRequest(ctx, modelName, hfReq)
	if err != nil {
		return nil, err
	}

	var hfResponses []HuggingFaceResponse
	if err := json.Unmarshal(response, &hfResponses); err != nil {
		return nil, fmt.Errorf("failed to parse summarization response: %w", err)
	}

	if len(hfResponses) == 0 {
		return nil, fmt.Errorf("no summarization result")
	}

	summary := hfResponses[0].SummaryText
	compression := float64(len(summary)) / float64(len(text))

	return &model.SummaryResponse{
		OriginalText: text,
		Summary:      summary,
		Compression:  compression,
	}, nil
}

// ValidateModel validates if the model is supported
func (s *HuggingFaceService) ValidateModel(modelName string) error {
	if modelName == "" {
		return fmt.Errorf("model name cannot be empty")
	}

	// Basic validation - in a real implementation, you might want to
	// make a request to check if the model exists
	supportedModels := []string{
		"gpt2", "gpt2-medium", "gpt2-large", "gpt2-xl",
		"microsoft/DialoGPT-medium", "microsoft/DialoGPT-large",
		"facebook/bart-large-cnn", "cardiffnlp/twitter-roberta-base-sentiment-latest",
	}

	for _, supported := range supportedModels {
		if modelName == supported {
			return nil
		}
	}

	// Allow any model name for flexibility
	s.logger.Warn(context.Background(), "Using unvalidated model", map[string]interface{}{
		"model": modelName,
	})

	return nil
}

// makeRequest makes an HTTP request to Hugging Face API
func (s *HuggingFaceService) makeRequest(ctx context.Context, modelName string, req *HuggingFaceRequest) ([]byte, error) {
	requestBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/models/%s", s.config.BaseURL, modelName)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+s.config.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "go-ai-huggingface/1.0")

	// Retry logic
	var lastErr error
	for attempt := 0; attempt <= s.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(s.config.RetryDelay):
			}
			s.logger.Info(ctx, "Retrying request", map[string]interface{}{
				"attempt": attempt,
				"model":   modelName,
			})
		}

		// Recreate the request body for each attempt as it's consumed by Do()
		httpReq, err = http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		httpReq.Header.Set("Authorization", "Bearer "+s.config.APIKey)
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("User-Agent", "go-ai-huggingface/1.0")

		resp, err := s.httpClient.Do(httpReq)
		if err != nil {
			lastErr = fmt.Errorf("HTTP request failed: %w", err)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close() // Close immediately to avoid leaks
		if err != nil {
			lastErr = fmt.Errorf("failed to read response: %w", err)
			continue
		}

		if resp.StatusCode == http.StatusOK {
			return body, nil
		}

		// Handle error response
		var hfError HuggingFaceError
		if json.Unmarshal(body, &hfError) == nil {
			lastErr = fmt.Errorf("API error (%d): %s - %s", resp.StatusCode, hfError.Error, hfError.Message)
		} else {
			lastErr = fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
		}

		// Don't retry on client errors (4xx)
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			break
		}
	}

	return nil, lastErr
}
