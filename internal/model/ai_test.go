package model

import (
	"testing"
	"time"
)

func TestAIRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request AIRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: AIRequest{
				ID:          "test-id",
				Model:       "gpt2",
				Prompt:      "Hello world",
				MaxTokens:   100,
				Temperature: 0.7,
				TopP:        0.9,
				CreatedAt:   time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing prompt",
			request: AIRequest{
				ID:        "test-id",
				Model:     "gpt2",
				Prompt:    "",
				MaxTokens: 100,
			},
			wantErr: true,
			errMsg:  "prompt is required",
		},
		{
			name: "missing model",
			request: AIRequest{
				ID:        "test-id",
				Model:     "",
				Prompt:    "Hello world",
				MaxTokens: 100,
			},
			wantErr: true,
			errMsg:  "model is required",
		},
		{
			name: "negative max tokens",
			request: AIRequest{
				ID:        "test-id",
				Model:     "gpt2",
				Prompt:    "Hello world",
				MaxTokens: -1,
			},
			wantErr: true,
			errMsg:  "max_tokens must be positive",
		},
		{
			name: "temperature too low",
			request: AIRequest{
				ID:          "test-id",
				Model:       "gpt2",
				Prompt:      "Hello world",
				MaxTokens:   100,
				Temperature: -0.1,
			},
			wantErr: true,
			errMsg:  "temperature must be between 0 and 1",
		},
		{
			name: "temperature too high",
			request: AIRequest{
				ID:          "test-id",
				Model:       "gpt2",
				Prompt:      "Hello world",
				MaxTokens:   100,
				Temperature: 1.1,
			},
			wantErr: true,
			errMsg:  "temperature must be between 0 and 1",
		},
		{
			name: "zero max tokens (allowed)",
			request: AIRequest{
				ID:        "test-id",
				Model:     "gpt2",
				Prompt:    "Hello world",
				MaxTokens: 0,
			},
			wantErr: false,
		},
		{
			name: "temperature at boundaries",
			request: AIRequest{
				ID:          "test-id",
				Model:       "gpt2",
				Prompt:      "Hello world",
				Temperature: 0.0,
			},
			wantErr: false,
		},
		{
			name: "temperature at upper boundary",
			request: AIRequest{
				ID:          "test-id",
				Model:       "gpt2",
				Prompt:      "Hello world",
				Temperature: 1.0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("AIRequest.Validate() expected error but got nil")
					return
				}
				
				if errResp, ok := err.(*ErrorResponse); ok {
					if errResp.Message != tt.errMsg {
						t.Errorf("AIRequest.Validate() error message = %v, want %v", errResp.Message, tt.errMsg)
					}
					if errResp.Code != 400 {
						t.Errorf("AIRequest.Validate() error code = %v, want %v", errResp.Code, 400)
					}
					if errResp.Type != "validation_error" {
						t.Errorf("AIRequest.Validate() error type = %v, want %v", errResp.Type, "validation_error")
					}
				} else {
					t.Errorf("AIRequest.Validate() error type = %T, want *ErrorResponse", err)
				}
			} else {
				if err != nil {
					t.Errorf("AIRequest.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestErrorResponse_Error(t *testing.T) {
	tests := []struct {
		name string
		err  ErrorResponse
		want string
	}{
		{
			name: "simple error",
			err: ErrorResponse{
				Code:    400,
				Message: "test error message",
				Type:    "test_error",
			},
			want: "test error message",
		},
		{
			name: "empty message",
			err: ErrorResponse{
				Code:    500,
				Message: "",
				Type:    "internal_error",
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("ErrorResponse.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAIRequest_ValidateWithParameters(t *testing.T) {
	// Test with additional parameters
	request := AIRequest{
		ID:     "test-id",
		Model:  "gpt2",
		Prompt: "Hello world",
		Parameters: map[string]interface{}{
			"custom_param": "value",
			"number_param": 42,
		},
	}

	err := request.Validate()
	if err != nil {
		t.Errorf("AIRequest.Validate() with parameters unexpected error = %v", err)
	}
}

func TestChoiceStruct(t *testing.T) {
	choice := Choice{
		Index:        1,
		Text:         "Generated text",
		FinishReason: "length",
		LogProbs:     nil,
	}

	if choice.Index != 1 {
		t.Errorf("Choice.Index = %v, want %v", choice.Index, 1)
	}
	if choice.Text != "Generated text" {
		t.Errorf("Choice.Text = %v, want %v", choice.Text, "Generated text")
	}
	if choice.FinishReason != "length" {
		t.Errorf("Choice.FinishReason = %v, want %v", choice.FinishReason, "length")
	}
	if choice.LogProbs != nil {
		t.Errorf("Choice.LogProbs = %v, want nil", choice.LogProbs)
	}

	// Test with LogProbs
	logProb := -2.5
	choice.LogProbs = &logProb
	if choice.LogProbs == nil || *choice.LogProbs != -2.5 {
		t.Errorf("Choice.LogProbs = %v, want %v", choice.LogProbs, &logProb)
	}
}

func TestUsageStruct(t *testing.T) {
	usage := Usage{
		PromptTokens:     10,
		CompletionTokens: 20,
		TotalTokens:      30,
	}

	if usage.PromptTokens != 10 {
		t.Errorf("Usage.PromptTokens = %v, want %v", usage.PromptTokens, 10)
	}
	if usage.CompletionTokens != 20 {
		t.Errorf("Usage.CompletionTokens = %v, want %v", usage.CompletionTokens, 20)
	}
	if usage.TotalTokens != 30 {
		t.Errorf("Usage.TotalTokens = %v, want %v", usage.TotalTokens, 30)
	}
}

func TestSentimentResponse(t *testing.T) {
	sentiment := SentimentResponse{
		Text:       "I love this product!",
		Sentiment:  "positive",
		Score:      0.95,
		Confidence: 0.92,
	}

	if sentiment.Text != "I love this product!" {
		t.Errorf("SentimentResponse.Text = %v, want %v", sentiment.Text, "I love this product!")
	}
	if sentiment.Sentiment != "positive" {
		t.Errorf("SentimentResponse.Sentiment = %v, want %v", sentiment.Sentiment, "positive")
	}
	if sentiment.Score != 0.95 {
		t.Errorf("SentimentResponse.Score = %v, want %v", sentiment.Score, 0.95)
	}
	if sentiment.Confidence != 0.92 {
		t.Errorf("SentimentResponse.Confidence = %v, want %v", sentiment.Confidence, 0.92)
	}
}

func TestSummaryResponse(t *testing.T) {
	summary := SummaryResponse{
		OriginalText: "This is a very long text that needs to be summarized for better understanding.",
		Summary:      "Long text summarized.",
		Compression:  0.25,
	}

	if summary.OriginalText != "This is a very long text that needs to be summarized for better understanding." {
		t.Errorf("SummaryResponse.OriginalText = %v", summary.OriginalText)
	}
	if summary.Summary != "Long text summarized." {
		t.Errorf("SummaryResponse.Summary = %v, want %v", summary.Summary, "Long text summarized.")
	}
	if summary.Compression != 0.25 {
		t.Errorf("SummaryResponse.Compression = %v, want %v", summary.Compression, 0.25)
	}
}

func TestAIResponseStruct(t *testing.T) {
	now := time.Now()
	response := AIResponse{
		ID:           "response-id",
		Model:        "gpt2",
		GeneratedAt:  now,
		ProcessingMs: 150,
		Choices: []Choice{
			{
				Index:        0,
				Text:         "Generated response",
				FinishReason: "stop",
			},
		},
		Usage: Usage{
			PromptTokens:     5,
			CompletionTokens: 10,
			TotalTokens:      15,
		},
	}

	if response.ID != "response-id" {
		t.Errorf("AIResponse.ID = %v, want %v", response.ID, "response-id")
	}
	if response.Model != "gpt2" {
		t.Errorf("AIResponse.Model = %v, want %v", response.Model, "gpt2")
	}
	if !response.GeneratedAt.Equal(now) {
		t.Errorf("AIResponse.GeneratedAt = %v, want %v", response.GeneratedAt, now)
	}
	if response.ProcessingMs != 150 {
		t.Errorf("AIResponse.ProcessingMs = %v, want %v", response.ProcessingMs, 150)
	}
	if len(response.Choices) != 1 {
		t.Errorf("len(AIResponse.Choices) = %v, want %v", len(response.Choices), 1)
	}
	if response.Choices[0].Text != "Generated response" {
		t.Errorf("AIResponse.Choices[0].Text = %v, want %v", response.Choices[0].Text, "Generated response")
	}
}

// Benchmark tests
func BenchmarkAIRequest_Validate(b *testing.B) {
	request := AIRequest{
		ID:          "test-id",
		Model:       "gpt2",
		Prompt:      "Hello world",
		MaxTokens:   100,
		Temperature: 0.7,
		CreatedAt:   time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request.Validate()
	}
}

func BenchmarkErrorResponse_Error(b *testing.B) {
	err := ErrorResponse{
		Code:    400,
		Message: "test error message",
		Type:    "test_error",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err.Error()
	}
}