package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/tusharr/go-ai-huggingface/internal/model"
	"github.com/tusharr/go-ai-huggingface/pkg/logger"
	"github.com/google/uuid"
)

// AIHandler handles AI-related HTTP requests
type AIHandler struct {
	aiService model.AIService
	logger    logger.Logger
}

// NewAIHandler creates a new AI handler
func NewAIHandler(aiService model.AIService, logger logger.Logger) *AIHandler {
	return &AIHandler{
		aiService: aiService,
		logger:    logger,
	}
}

// GenerateText handles text generation requests
func (h *AIHandler) GenerateText(w http.ResponseWriter, r *http.Request) {
	ctx := h.setRequestID(r.Context())
	h.logger.Info(ctx, "Received text generation request", nil)

	var req model.AIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.handleError(ctx, w, &model.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Invalid JSON: %v", err),
			Type:    "validation_error",
		})
		return
	}

	// Set request ID and timestamp if not provided
	if req.ID == "" {
		req.ID = uuid.New().String()
	}
	req.CreatedAt = time.Now()

	// Validate request
	if err := req.Validate(); err != nil {
		if errResp, ok := err.(*model.ErrorResponse); ok {
			h.handleError(ctx, w, errResp)
		} else {
			h.handleError(ctx, w, &model.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
				Type:    "validation_error",
			})
		}
		return
	}

	// Generate text
	response, err := h.aiService.GenerateText(ctx, &req)
	if err != nil {
		h.logger.Error(ctx, "Failed to generate text", map[string]interface{}{
			"error": err.Error(),
		})
		h.handleError(ctx, w, &model.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to generate text",
			Type:    "service_error",
			Details: err.Error(),
		})
		return
	}

	h.sendJSONResponse(ctx, w, http.StatusOK, response)
}

// GenerateCompletion handles text completion requests
func (h *AIHandler) GenerateCompletion(w http.ResponseWriter, r *http.Request) {
	ctx := h.setRequestID(r.Context())
	h.logger.Info(ctx, "Received completion request", nil)

	var req model.AIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.handleError(ctx, w, &model.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Invalid JSON: %v", err),
			Type:    "validation_error",
		})
		return
	}

	if req.ID == "" {
		req.ID = uuid.New().String()
	}
	req.CreatedAt = time.Now()

	if err := req.Validate(); err != nil {
		if errResp, ok := err.(*model.ErrorResponse); ok {
			h.handleError(ctx, w, errResp)
		} else {
			h.handleError(ctx, w, &model.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
				Type:    "validation_error",
			})
		}
		return
	}

	response, err := h.aiService.GenerateCompletion(ctx, &req)
	if err != nil {
		h.logger.Error(ctx, "Failed to generate completion", map[string]interface{}{
			"error": err.Error(),
		})
		h.handleError(ctx, w, &model.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to generate completion",
			Type:    "service_error",
			Details: err.Error(),
		})
		return
	}

	h.sendJSONResponse(ctx, w, http.StatusOK, response)
}

// AnalyzeSentiment handles sentiment analysis requests
func (h *AIHandler) AnalyzeSentiment(w http.ResponseWriter, r *http.Request) {
	ctx := h.setRequestID(r.Context())
	h.logger.Info(ctx, "Received sentiment analysis request", nil)

	var req struct {
		Text string `json:"text"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.handleError(ctx, w, &model.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Invalid JSON: %v", err),
			Type:    "validation_error",
		})
		return
	}

	if req.Text == "" {
		h.handleError(ctx, w, &model.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Text is required",
			Type:    "validation_error",
		})
		return
	}

	response, err := h.aiService.AnalyzeSentiment(ctx, req.Text)
	if err != nil {
		h.logger.Error(ctx, "Failed to analyze sentiment", map[string]interface{}{
			"error": err.Error(),
		})
		h.handleError(ctx, w, &model.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to analyze sentiment",
			Type:    "service_error",
			Details: err.Error(),
		})
		return
	}

	h.sendJSONResponse(ctx, w, http.StatusOK, response)
}

// SummarizeText handles text summarization requests
func (h *AIHandler) SummarizeText(w http.ResponseWriter, r *http.Request) {
	ctx := h.setRequestID(r.Context())
	h.logger.Info(ctx, "Received summarization request", nil)

	var req struct {
		Text      string `json:"text"`
		MaxLength int    `json:"max_length,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.handleError(ctx, w, &model.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Invalid JSON: %v", err),
			Type:    "validation_error",
		})
		return
	}

	if req.Text == "" {
		h.handleError(ctx, w, &model.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Text is required",
			Type:    "validation_error",
		})
		return
	}

	// Default max length
	if req.MaxLength <= 0 {
		req.MaxLength = 130
	}

	response, err := h.aiService.SummarizeText(ctx, req.Text, req.MaxLength)
	if err != nil {
		h.logger.Error(ctx, "Failed to summarize text", map[string]interface{}{
			"error": err.Error(),
		})
		h.handleError(ctx, w, &model.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to summarize text",
			Type:    "service_error",
			Details: err.Error(),
		})
		return
	}

	h.sendJSONResponse(ctx, w, http.StatusOK, response)
}

// ValidateModel handles model validation requests
func (h *AIHandler) ValidateModel(w http.ResponseWriter, r *http.Request) {
	ctx := h.setRequestID(r.Context())
	modelName := r.URL.Query().Get("model")
	
	if modelName == "" {
		h.handleError(ctx, w, &model.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Model parameter is required",
			Type:    "validation_error",
		})
		return
	}

	err := h.aiService.ValidateModel(modelName)
	if err != nil {
		h.handleError(ctx, w, &model.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Invalid model: %v", err),
			Type:    "validation_error",
		})
		return
	}

	response := map[string]interface{}{
		"model": modelName,
		"valid": true,
	}

	h.sendJSONResponse(ctx, w, http.StatusOK, response)
}

// Health handles health check requests
func (h *AIHandler) Health(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "go-ai-huggingface",
		"version":   "1.0.0",
	}

	h.sendJSONResponse(r.Context(), w, http.StatusOK, response)
}

// Metrics handles metrics requests (basic implementation)
func (h *AIHandler) Metrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// In a production system, you'd collect real metrics
	metrics := map[string]interface{}{
		"requests_total":     1000, // Example counter
		"requests_duration": "150ms", // Example histogram
		"error_rate":        "0.01",  // Example gauge
		"timestamp":         time.Now().UTC().Format(time.RFC3339),
	}

	h.sendJSONResponse(ctx, w, http.StatusOK, metrics)
}

// setRequestID adds a request ID to the context if not already present
func (h *AIHandler) setRequestID(ctx context.Context) context.Context {
	if id, ok := ctx.Value("request_id").(string); ok && id != "" {
		return ctx
	}
	requestID := uuid.New().String()
	return context.WithValue(ctx, "request_id", requestID)
}

// handleError handles error responses
func (h *AIHandler) handleError(ctx context.Context, w http.ResponseWriter, errResp *model.ErrorResponse) {
	h.logger.Error(ctx, "Request error", map[string]interface{}{
		"error_code":    errResp.Code,
		"error_message": errResp.Message,
		"error_type":    errResp.Type,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errResp.Code)
	json.NewEncoder(w).Encode(errResp)
}

// sendJSONResponse sends a JSON response
func (h *AIHandler) sendJSONResponse(ctx context.Context, w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error(ctx, "Failed to encode JSON response", map[string]interface{}{
			"error": err.Error(),
		})
	}
}

// CORS middleware
func (h *AIHandler) EnableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequestLogger middleware
func (h *AIHandler) RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ctx := h.setRequestID(r.Context())
		
		h.logger.Info(ctx, "Request started", map[string]interface{}{
			"method": r.Method,
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
		})

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		next.ServeHTTP(wrapped, r.WithContext(ctx))

		duration := time.Since(start)
		h.logger.Info(ctx, "Request completed", map[string]interface{}{
			"method":      r.Method,
			"path":        r.URL.Path,
			"status_code": wrapped.statusCode,
			"duration_ms": duration.Milliseconds(),
		})
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// RateLimiter middleware (basic implementation)
func (h *AIHandler) RateLimiter(requestsPerMinute int) func(http.Handler) http.Handler {
	// Simple in-memory rate limiter - in production use Redis or similar
	var mu sync.Mutex
	requests := make(map[string][]time.Time)
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := r.RemoteAddr
			now := time.Now()
			
			mu.Lock()
			// Clean old requests
			if times, exists := requests[clientIP]; exists {
				validTimes := make([]time.Time, 0)
				for _, t := range times {
					if now.Sub(t) < time.Minute {
						validTimes = append(validTimes, t)
					}
				}
				requests[clientIP] = validTimes
			}
			
			// Check rate limit
			if len(requests[clientIP]) >= requestsPerMinute {
				mu.Unlock()
				h.handleError(r.Context(), w, &model.ErrorResponse{
					Code:    http.StatusTooManyRequests,
					Message: "Rate limit exceeded",
					Type:    "rate_limit_error",
				})
				return
			}
			
			// Add current request
			requests[clientIP] = append(requests[clientIP], now)
			mu.Unlock()
			
			next.ServeHTTP(w, r)
		})
	}
}