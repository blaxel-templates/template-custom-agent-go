package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"go.opentelemetry.io/otel/trace"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	TRACE LogLevel = iota
	DEBUG
	INFO
	WARNING
	ERROR
	FATAL
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case TRACE:
		return "TRACE"
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// JsonFormatter handles JSON log formatting with OpenTelemetry context
type JsonFormatter struct {
	TraceIdName     string
	SpanIdName      string
	LabelsName      string
	TraceIdPrefix   string
	SpanIdPrefix    string
	TaskIndex       string
	TaskPrefix      string
	ExecutionKey    string
	ExecutionPrefix string
}

// NewJsonFormatter creates a new JSON formatter with environment variable configuration
func NewJsonFormatter() *JsonFormatter {
	return &JsonFormatter{
		TraceIdName:     getEnvOrDefault("BL_LOGGER_TRACE_ID", "trace_id"),
		SpanIdName:      getEnvOrDefault("BL_LOGGER_SPAN_ID", "span_id"),
		LabelsName:      getEnvOrDefault("BL_LOGGER_LABELS", "labels"),
		TraceIdPrefix:   getEnvOrDefault("BL_LOGGER_TRACE_ID_PREFIX", ""),
		SpanIdPrefix:    getEnvOrDefault("BL_LOGGER_SPAN_ID_PREFIX", ""),
		TaskIndex:       getEnvOrDefault("BL_TASK_KEY", "TASK_INDEX"),
		TaskPrefix:      getEnvOrDefault("BL_TASK_PREFIX", ""),
		ExecutionKey:    getEnvOrDefault("BL_EXECUTION_KEY", "BL_EXECUTION_ID"),
		ExecutionPrefix: getEnvOrDefault("BL_EXECUTION_PREFIX", ""),
	}
}

// Format formats a log entry as JSON with trace context
func (jf *JsonFormatter) Format(ctx context.Context, level LogLevel, message string) string {
	logEntry := map[string]interface{}{
		"message":     message,
		"severity":    level.String(),
		jf.LabelsName: map[string]string{},
	}

	// Get current active span from context
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		spanContext := span.SpanContext()
		traceIdHex := spanContext.TraceID().String()
		spanIdHex := spanContext.SpanID().String()

		logEntry[jf.TraceIdName] = jf.TraceIdPrefix + traceIdHex
		logEntry[jf.SpanIdName] = jf.SpanIdPrefix + spanIdHex
	}

	// Add task ID if available
	if taskId := os.Getenv(jf.TaskIndex); taskId != "" {
		labels := logEntry[jf.LabelsName].(map[string]string)
		labels["blaxel-task"] = jf.TaskPrefix + taskId
	}

	// Add execution ID if available
	if executionId := os.Getenv(jf.ExecutionKey); executionId != "" {
		labels := logEntry[jf.LabelsName].(map[string]string)
		parts := strings.Split(executionId, "-")
		if len(parts) > 0 {
			labels["blaxel-execution"] = jf.ExecutionPrefix + parts[len(parts)-1]
		}
	}

	jsonBytes, _ := json.Marshal(logEntry)
	return string(jsonBytes)
}

// ColoredFormatter handles colored log formatting
type ColoredFormatter struct {
	Colors map[string]string
}

// NewColoredFormatter creates a new colored formatter
func NewColoredFormatter() *ColoredFormatter {
	return &ColoredFormatter{
		Colors: map[string]string{
			"TRACE":   "\033[1;35m", // Magenta
			"DEBUG":   "\033[1;36m", // Cyan
			"INFO":    "\033[1;32m", // Green
			"WARNING": "\033[1;33m", // Yellow
			"ERROR":   "\033[1;31m", // Red
			"FATAL":   "\033[1;41m", // Red background
		},
	}
}

// Format formats a log entry with colors
func (cf *ColoredFormatter) Format(ctx context.Context, level LogLevel, message string) string {
	levelStr := level.String()
	color := cf.Colors[levelStr]
	if color == "" {
		color = "\033[0m"
	}

	// Calculate spacing to align log levels
	maxLevelLen := 7 // Length of "WARNING"
	spaces := strings.Repeat(" ", maxLevelLen-len(levelStr))

	return fmt.Sprintf("%s%s\033[0m:%s %s", color, levelStr, spaces, message)
}

// Formatter interface for different log formatters
type Formatter interface {
	Format(ctx context.Context, level LogLevel, message string) string
}

// Logger represents our custom logger
type Logger struct {
	level     LogLevel
	formatter Formatter
	logger    *log.Logger
}

// Global logger instance
var globalLogger *Logger

// init initializes the global logger
func init() {
	globalLogger = New()
}

// New creates a new logger instance
func New() *Logger {
	level := getLogLevelFromEnv()
	formatter := getFormatterFromEnv()

	return &Logger{
		level:     level,
		formatter: formatter,
		logger:    log.New(os.Stdout, "", 0), // No default formatting
	}
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getFormatterFromEnv returns the appropriate formatter based on BL_LOGGER env var
func getFormatterFromEnv() Formatter {
	loggerType := getEnvOrDefault("BL_LOGGER", "colored")
	if loggerType == "json" {
		return NewJsonFormatter()
	}
	return NewColoredFormatter()
}

// getLogLevelFromEnv reads the log level from environment variable
func getLogLevelFromEnv() LogLevel {
	envLevel := strings.ToUpper(os.Getenv("LOG_LEVEL"))
	switch envLevel {
	case "TRACE":
		return TRACE
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARNING":
		return WARNING
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	default:
		return DEBUG // Default to DEBUG level
	}
}

// SetLevel sets the minimum log level
func SetLevel(level LogLevel) {
	globalLogger.level = level
}

// SetLevelFromString sets the log level from a string
func SetLevelFromString(levelStr string) {
	switch strings.ToUpper(levelStr) {
	case "TRACE":
		SetLevel(TRACE)
	case "DEBUG":
		SetLevel(DEBUG)
	case "INFO":
		SetLevel(INFO)
	case "WARNING":
		SetLevel(WARNING)
	case "ERROR":
		SetLevel(ERROR)
	case "FATAL":
		SetLevel(FATAL)
	}
}

// InitLogger initializes the logging configuration
func InitLogger(logLevel string) {
	SetLevelFromString(logLevel)
	// You can add additional initialization logic here
}

// shouldLog checks if a message should be logged based on the current level
func (l *Logger) shouldLog(level LogLevel) bool {
	return level >= l.level
}

// logf formats and logs a message if the level is appropriate
func (l *Logger) logf(level LogLevel, format string, args ...interface{}) {
	if !l.shouldLog(level) {
		return
	}

	message := fmt.Sprintf(format, args...)
	ctx := context.Background() // You can pass context from calling functions for trace context
	formattedMessage := l.formatter.Format(ctx, level, message)
	l.logger.Print(formattedMessage)

	// Exit the program for FATAL logs
	if level == FATAL {
		os.Exit(1)
	}
}

// Global logger functions
func Trace(message string) {
	globalLogger.logf(TRACE, "%s", message)
}

func Tracef(format string, args ...interface{}) {
	globalLogger.logf(TRACE, format, args...)
}

func Debug(message string) {
	globalLogger.logf(DEBUG, "%s", message)
}

func Debugf(format string, args ...interface{}) {
	globalLogger.logf(DEBUG, format, args...)
}

func Info(message string) {
	globalLogger.logf(INFO, "%s", message)
}

func Infof(format string, args ...interface{}) {
	globalLogger.logf(INFO, format, args...)
}

func Warning(message string) {
	globalLogger.logf(WARNING, "%s", message)
}

func Warningf(format string, args ...interface{}) {
	globalLogger.logf(WARNING, format, args...)
}

func Error(message string) {
	globalLogger.logf(ERROR, "%s", message)
}

func Errorf(format string, args ...interface{}) {
	globalLogger.logf(ERROR, format, args...)
}

func Fatal(message string) {
	globalLogger.logf(FATAL, "%s", message)
}

func Fatalf(format string, args ...interface{}) {
	globalLogger.logf(FATAL, format, args...)
}

// GetLevel returns the current log level
func GetLevel() LogLevel {
	return globalLogger.level
}
