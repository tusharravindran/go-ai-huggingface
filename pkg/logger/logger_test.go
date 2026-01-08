package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"strings"
	"testing"
)

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		level LogLevel
		want  string
	}{
		{DebugLevel, "DEBUG"},
		{InfoLevel, "INFO"},
		{WarnLevel, "WARN"},
		{ErrorLevel, "ERROR"},
		{LogLevel(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.level.String(); got != tt.want {
				t.Errorf("LogLevel.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input string
		want  LogLevel
	}{
		{"debug", DebugLevel},
		{"DEBUG", DebugLevel},
		{"info", InfoLevel},
		{"INFO", InfoLevel},
		{"warn", WarnLevel},
		{"WARN", WarnLevel},
		{"warning", WarnLevel},
		{"WARNING", WarnLevel},
		{"error", ErrorLevel},
		{"ERROR", ErrorLevel},
		{"invalid", InfoLevel}, // default
		{"", InfoLevel},        // default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ParseLogLevel(tt.input); got != tt.want {
				t.Errorf("ParseLogLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewLogger(t *testing.T) {
	logger := NewLogger(InfoLevel, true)
	
	if logger == nil {
		t.Error("NewLogger() returned nil")
	}

	structuredLogger, ok := logger.(*StructuredLogger)
	if !ok {
		t.Error("NewLogger() did not return *StructuredLogger")
	}

	if structuredLogger.level != InfoLevel {
		t.Errorf("NewLogger() level = %v, want %v", structuredLogger.level, InfoLevel)
	}

	if !structuredLogger.structured {
		t.Error("NewLogger() structured = false, want true")
	}
}

func TestStructuredLogger_LogLevels(t *testing.T) {
	var buf bytes.Buffer
	logger := &StructuredLogger{
		level:      DebugLevel,
		fields:     make(map[string]interface{}),
		structured: true,
		output:     log.New(&buf, "", 0),
	}

	ctx := context.Background()
	testMessage := "test message"
	testFields := map[string]interface{}{"key": "value"}

	tests := []struct {
		name     string
		logFunc  func(context.Context, string, map[string]interface{})
		minLevel LogLevel
		want     string
	}{
		{"Debug", logger.Debug, DebugLevel, "DEBUG"},
		{"Info", logger.Info, InfoLevel, "INFO"},
		{"Warn", logger.Warn, WarnLevel, "WARN"},
		{"Error", logger.Error, ErrorLevel, "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			logger.level = tt.minLevel
			
			tt.logFunc(ctx, testMessage, testFields)
			
			output := buf.String()
			if output == "" {
				t.Errorf("%s() produced no output", tt.name)
				return
			}

			var entry LogEntry
			if err := json.Unmarshal([]byte(output), &entry); err != nil {
				t.Errorf("%s() output is not valid JSON: %v", tt.name, err)
				return
			}

			if entry.Level != tt.want {
				t.Errorf("%s() level = %v, want %v", tt.name, entry.Level, tt.want)
			}

			if entry.Message != testMessage {
				t.Errorf("%s() message = %v, want %v", tt.name, entry.Message, testMessage)
			}

			if entry.Fields["key"] != "value" {
				t.Errorf("%s() fields[key] = %v, want %v", tt.name, entry.Fields["key"], "value")
			}
		})
	}
}

func TestStructuredLogger_LevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := &StructuredLogger{
		level:      WarnLevel,
		fields:     make(map[string]interface{}),
		structured: true,
		output:     log.New(&buf, "", 0),
	}

	ctx := context.Background()

	// Debug and Info should be filtered out
	logger.Debug(ctx, "debug message", nil)
	logger.Info(ctx, "info message", nil)
	
	if buf.Len() > 0 {
		t.Error("Debug/Info messages should be filtered out at WARN level")
	}

	// Warn and Error should pass through
	logger.Warn(ctx, "warn message", nil)
	logger.Error(ctx, "error message", nil)
	
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	
	if len(lines) != 2 {
		t.Errorf("Expected 2 log entries, got %d", len(lines))
	}
}

func TestStructuredLogger_WithFields(t *testing.T) {
	var buf bytes.Buffer
	baseLogger := &StructuredLogger{
		level:      InfoLevel,
		fields:     map[string]interface{}{"base": "value"},
		structured: true,
		output:     log.New(&buf, "", 0),
	}

	// Test WithFields creates new logger with merged fields
	newLogger := baseLogger.WithFields(map[string]interface{}{"new": "field"})
	
	if newLogger == baseLogger {
		t.Error("WithFields() should return a new logger instance")
	}

	newLogger.Info(context.Background(), "test", map[string]interface{}{"extra": "data"})
	
	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Errorf("Failed to parse JSON output: %v", err)
		return
	}

	// Check all fields are present
	expectedFields := map[string]interface{}{
		"base":  "value",
		"new":   "field",
		"extra": "data",
	}

	for k, v := range expectedFields {
		if entry.Fields[k] != v {
			t.Errorf("Field %s = %v, want %v", k, entry.Fields[k], v)
		}
	}
}

func TestStructuredLogger_SetLevel(t *testing.T) {
	logger := &StructuredLogger{
		level: InfoLevel,
	}

	logger.SetLevel(ErrorLevel)
	
	if logger.level != ErrorLevel {
		t.Errorf("SetLevel() level = %v, want %v", logger.level, ErrorLevel)
	}
}

func TestStructuredLogger_PlainFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := &StructuredLogger{
		level:      InfoLevel,
		fields:     make(map[string]interface{}),
		structured: false, // Plain format
		output:     log.New(&buf, "", 0),
	}

	ctx := context.Background()
	logger.Info(ctx, "test message", map[string]interface{}{"key": "value"})
	
	output := buf.String()
	
	// Should contain level, message, and JSON fields
	if !strings.Contains(output, "INFO") {
		t.Error("Plain output should contain log level")
	}
	if !strings.Contains(output, "test message") {
		t.Error("Plain output should contain message")
	}
	if !strings.Contains(output, `"key":"value"`) {
		t.Error("Plain output should contain fields as JSON")
	}
}

func TestStructuredLogger_WithContext(t *testing.T) {
	var buf bytes.Buffer
	logger := &StructuredLogger{
		level:      InfoLevel,
		fields:     make(map[string]interface{}),
		structured: true,
		output:     log.New(&buf, "", 0),
	}

	// Create context with values
	ctx := context.WithValue(context.Background(), "request_id", "req-123")
	ctx = context.WithValue(ctx, "trace_id", "trace-456")
	
	logger.Info(ctx, "test message", nil)
	
	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Errorf("Failed to parse JSON output: %v", err)
		return
	}

	if entry.RequestID != "req-123" {
		t.Errorf("RequestID = %v, want %v", entry.RequestID, "req-123")
	}

	if entry.TraceID != "trace-456" {
		t.Errorf("TraceID = %v, want %v", entry.TraceID, "trace-456")
	}
}

func TestStructuredLogger_InvalidJSONMarshaling(t *testing.T) {
	var buf bytes.Buffer
	logger := &StructuredLogger{
		level:      InfoLevel,
		fields:     make(map[string]interface{}),
		structured: true,
		output:     log.New(&buf, "", 0),
	}

	// Create a field that can't be marshaled to JSON
	invalidField := map[string]interface{}{
		"func": func() {}, // functions can't be marshaled to JSON
	}

	logger.Info(context.Background(), "test message", invalidField)
	
	output := buf.String()
	if !strings.Contains(output, "Error marshaling log entry") {
		t.Error("Should handle JSON marshaling errors gracefully")
	}
}

func TestNoopLogger(t *testing.T) {
	logger := NewNoopLogger()
	ctx := context.Background()
	fields := map[string]interface{}{"key": "value"}

	// All methods should do nothing and not panic
	logger.Debug(ctx, "debug", fields)
	logger.Info(ctx, "info", fields)
	logger.Warn(ctx, "warn", fields)
	logger.Error(ctx, "error", fields)
	
	newLogger := logger.WithFields(fields)
	if newLogger != logger {
		t.Error("NoopLogger.WithFields() should return itself")
	}
	
	logger.SetLevel(ErrorLevel) // Should not panic
}

func TestGetStringFromContext(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		key  string
		want string
	}{
		{
			name: "string value exists",
			ctx:  context.WithValue(context.Background(), "test_key", "test_value"),
			key:  "test_key",
			want: "test_value",
		},
		{
			name: "key does not exist",
			ctx:  context.Background(),
			key:  "nonexistent",
			want: "",
		},
		{
			name: "value is not string",
			ctx:  context.WithValue(context.Background(), "int_key", 123),
			key:  "int_key",
			want: "",
		},
		{
			name: "nil context",
			ctx:  nil,
			key:  "test_key",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getStringFromContext(tt.ctx, tt.key)
			if got != tt.want {
				t.Errorf("getStringFromContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLogEntry_AllFields(t *testing.T) {
	var buf bytes.Buffer
	logger := &StructuredLogger{
		level:      InfoLevel,
		fields:     map[string]interface{}{"persistent": "field"},
		structured: true,
		output:     log.New(&buf, "", 0),
	}

	ctx := context.WithValue(context.Background(), "request_id", "req-123")
	ctx = context.WithValue(ctx, "trace_id", "trace-456")

	logger.Info(ctx, "test message", map[string]interface{}{"temp": "field"})

	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Errorf("Failed to parse JSON output: %v", err)
		return
	}

	if entry.Timestamp == "" {
		t.Error("LogEntry should have timestamp")
	}
	if entry.Level != "INFO" {
		t.Errorf("LogEntry.Level = %v, want INFO", entry.Level)
	}
	if entry.Message != "test message" {
		t.Errorf("LogEntry.Message = %v, want 'test message'", entry.Message)
	}
	if entry.RequestID != "req-123" {
		t.Errorf("LogEntry.RequestID = %v, want 'req-123'", entry.RequestID)
	}
	if entry.TraceID != "trace-456" {
		t.Errorf("LogEntry.TraceID = %v, want 'trace-456'", entry.TraceID)
	}
	if entry.Fields["persistent"] != "field" {
		t.Error("LogEntry should contain persistent fields")
	}
	if entry.Fields["temp"] != "field" {
		t.Error("LogEntry should contain temporary fields")
	}
}

// Benchmark tests
func BenchmarkStructuredLogger_Info(b *testing.B) {
	var buf bytes.Buffer
	logger := &StructuredLogger{
		level:      InfoLevel,
		fields:     make(map[string]interface{}),
		structured: true,
		output:     log.New(&buf, "", 0),
	}

	ctx := context.Background()
	fields := map[string]interface{}{"key": "value", "number": 42}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		logger.Info(ctx, "benchmark message", fields)
	}
}

func BenchmarkParseLogLevel(b *testing.B) {
	levels := []string{"debug", "info", "warn", "error", "invalid"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseLogLevel(levels[i%len(levels)])
	}
}