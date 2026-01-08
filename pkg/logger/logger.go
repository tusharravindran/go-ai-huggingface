package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

// LogLevel represents the log level
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger defines the logging interface
type Logger interface {
	Debug(ctx context.Context, message string, fields map[string]interface{})
	Info(ctx context.Context, message string, fields map[string]interface{})
	Warn(ctx context.Context, message string, fields map[string]interface{})
	Error(ctx context.Context, message string, fields map[string]interface{})
	WithFields(fields map[string]interface{}) Logger
	SetLevel(level LogLevel)
}

// StructuredLogger implements the Logger interface with structured logging
type StructuredLogger struct {
	level      LogLevel
	fields     map[string]interface{}
	structured bool
	output     *log.Logger
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
	TraceID   string                 `json:"trace_id,omitempty"`
}

// NewLogger creates a new structured logger
func NewLogger(level LogLevel, structured bool) Logger {
	return &StructuredLogger{
		level:      level,
		fields:     make(map[string]interface{}),
		structured: structured,
		output:     log.New(os.Stdout, "", 0),
	}
}

// Debug logs a debug message
func (l *StructuredLogger) Debug(ctx context.Context, message string, fields map[string]interface{}) {
	if l.level <= DebugLevel {
		l.log(ctx, DebugLevel, message, fields)
	}
}

// Info logs an info message
func (l *StructuredLogger) Info(ctx context.Context, message string, fields map[string]interface{}) {
	if l.level <= InfoLevel {
		l.log(ctx, InfoLevel, message, fields)
	}
}

// Warn logs a warning message
func (l *StructuredLogger) Warn(ctx context.Context, message string, fields map[string]interface{}) {
	if l.level <= WarnLevel {
		l.log(ctx, WarnLevel, message, fields)
	}
}

// Error logs an error message
func (l *StructuredLogger) Error(ctx context.Context, message string, fields map[string]interface{}) {
	if l.level <= ErrorLevel {
		l.log(ctx, ErrorLevel, message, fields)
	}
}

// WithFields returns a new logger with the given fields
func (l *StructuredLogger) WithFields(fields map[string]interface{}) Logger {
	newFields := make(map[string]interface{})
	
	// Copy existing fields
	for k, v := range l.fields {
		newFields[k] = v
	}
	
	// Add new fields
	for k, v := range fields {
		newFields[k] = v
	}

	return &StructuredLogger{
		level:      l.level,
		fields:     newFields,
		structured: l.structured,
		output:     l.output,
	}
}

// SetLevel sets the log level
func (l *StructuredLogger) SetLevel(level LogLevel) {
	l.level = level
}

// log performs the actual logging
func (l *StructuredLogger) log(ctx context.Context, level LogLevel, message string, fields map[string]interface{}) {
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level.String(),
		Message:   message,
		Fields:    l.mergeFields(fields),
	}

	// Extract context values
	if ctx != nil {
		if requestID := getStringFromContext(ctx, "request_id"); requestID != "" {
			entry.RequestID = requestID
		}
		if traceID := getStringFromContext(ctx, "trace_id"); traceID != "" {
			entry.TraceID = traceID
		}
	}

	if l.structured {
		l.logStructured(entry)
	} else {
		l.logPlain(entry)
	}
}

// logStructured logs in JSON format
func (l *StructuredLogger) logStructured(entry LogEntry) {
	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		l.output.Printf("Error marshaling log entry: %v", err)
		return
	}
	l.output.Println(string(jsonBytes))
}

// logPlain logs in plain text format
func (l *StructuredLogger) logPlain(entry LogEntry) {
	output := fmt.Sprintf("[%s] %s: %s", entry.Timestamp, entry.Level, entry.Message)
	
	if entry.RequestID != "" {
		output += fmt.Sprintf(" [request_id=%s]", entry.RequestID)
	}
	
	if entry.TraceID != "" {
		output += fmt.Sprintf(" [trace_id=%s]", entry.TraceID)
	}

	if len(entry.Fields) > 0 {
		fieldsStr, _ := json.Marshal(entry.Fields)
		output += fmt.Sprintf(" %s", string(fieldsStr))
	}

	l.output.Println(output)
}

// mergeFields merges logger fields with provided fields
func (l *StructuredLogger) mergeFields(fields map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	
	// Copy logger fields first
	for k, v := range l.fields {
		merged[k] = v
	}
	
	// Override with provided fields
	for k, v := range fields {
		merged[k] = v
	}

	return merged
}

// getStringFromContext extracts a string value from context
func getStringFromContext(ctx context.Context, key string) string {
	if ctx == nil {
		return ""
	}
	if value := ctx.Value(key); value != nil {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// ParseLogLevel parses a log level from string
func ParseLogLevel(level string) LogLevel {
	switch level {
	case "debug", "DEBUG":
		return DebugLevel
	case "info", "INFO":
		return InfoLevel
	case "warn", "WARN", "warning", "WARNING":
		return WarnLevel
	case "error", "ERROR":
		return ErrorLevel
	default:
		return InfoLevel
	}
}

// NoopLogger is a logger that does nothing (useful for testing)
type NoopLogger struct{}

// NewNoopLogger creates a new noop logger
func NewNoopLogger() Logger {
	return &NoopLogger{}
}

func (l *NoopLogger) Debug(ctx context.Context, message string, fields map[string]interface{}) {}
func (l *NoopLogger) Info(ctx context.Context, message string, fields map[string]interface{})  {}
func (l *NoopLogger) Warn(ctx context.Context, message string, fields map[string]interface{})  {}
func (l *NoopLogger) Error(ctx context.Context, message string, fields map[string]interface{}) {}
func (l *NoopLogger) WithFields(fields map[string]interface{}) Logger                          { return l }
func (l *NoopLogger) SetLevel(level LogLevel)                                                  {}