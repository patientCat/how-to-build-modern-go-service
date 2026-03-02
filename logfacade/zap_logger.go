package logfacade

import (
	"context"
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// ZapLogger 是基于 go.uber.org/zap 库的 Logger 接口实现。
// 它提供高性能的结构化日志记录功能，支持多种输出方式和日志级别。
type ZapLogger struct {
	// logger 是底层的 zap 日志实例
	logger *zap.Logger
}

// newZapLogger 创建基于 zap 的日志实例。
// 该函数是内部工厂方法，根据配置初始化 zap Logger。
//
// 参数：
//   - config: 日志配置
//   - skipStack: 调用栈跳过层数，用于正确显示 caller 信息
//
// 返回：
//   - Logger: 日志记录器接口实现
//   - error: 初始化错误
func newZapLogger(config Config, skipStack int) (Logger, error) {
	var core zapcore.Core

	if config.OutputPath == "stdout" {
		core = zapcore.NewCore(
			getEncoder(),
			zapcore.AddSync(os.Stdout),
			getZapLevel(config.Level),
		)
	} else {
		// 使用 lumberjack 进行日志轮转
		writer := &lumberjack.Logger{
			Filename:   config.OutputPath,
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			Compress:   config.Compress,
		}

		core = zapcore.NewCore(
			getEncoder(),
			zapcore.AddSync(writer),
			getZapLevel(config.Level),
		)
	}

	// 添加调用者信息和堆栈跟踪
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(skipStack), zap.AddStacktrace(zapcore.ErrorLevel))

	return &ZapLogger{logger: zapLogger}, nil
}

// Debug 记录调试级别的日志信息。
// 调试级别的日志仅在日志级别设置为 DebugLevel 时输出。
func (z *ZapLogger) Debug(msg string, fields ...Field) {
	z.logger.Debug(msg, z.convertFields(fields)...)
}

// Info 记录信息级别的日志信息。
// 信息级别适合记录常规的应用事件和状态变化。
func (z *ZapLogger) Info(msg string, fields ...Field) {
	z.logger.Info(msg, z.convertFields(fields)...)
}

// Infof 使用格式化字符串记录信息级别的日志。
// 参数与 fmt.Sprintf 兼容。
func (z *ZapLogger) Infof(msg string, args ...any) {
	z.logger.Info(fmt.Sprintf(msg, args...))
}

// DeepInfof 使用格式化字符串和深度展开记录信息级别的日志。
// 对复杂数据结构进行深度展开，便于调试。
// 使用 spew 库进行深度打印，性能相比普通日志会有所下降。
func (z *ZapLogger) DeepInfof(msg string, args ...any) {
	var expandedArgs []any
	for _, arg := range args {
		expandedArgs = append(expandedArgs, spew.Sdump(arg))
	}
	z.logger.Info(fmt.Sprintf(msg, expandedArgs...))
}

// Warn 记录警告级别的日志信息。
// 警告级别用于记录潜在的问题和非致命错误。
func (z *ZapLogger) Warn(msg string, fields ...Field) {
	z.logger.Warn(msg, z.convertFields(fields)...)
}

// Warnf 使用格式化字符串记录警告级别的日志。
// 参数与 fmt.Sprintf 兼容。
func (z *ZapLogger) Warnf(msg string, args ...any) {
	z.logger.Warn(fmt.Sprintf(msg, args...))
}

// Error 记录错误级别的日志信息。
// 错误级别用于记录运行时错误。
func (z *ZapLogger) Error(msg string, fields ...Field) {
	z.logger.Error(msg, z.convertFields(fields)...)
}

// Errorf 使用格式化字符串记录错误级别的日志。
// 参数与 fmt.Sprintf 兼容。
func (z *ZapLogger) Errorf(msg string, args ...any) {
	z.logger.Error(fmt.Sprintf(msg, args...))
}

// Fatal 记录致命错误日志并立即退出程序。
// 仅在不可恢复的错误发生时使用。调用此方法后程序将立即终止。
func (z *ZapLogger) Fatal(msg string, fields ...Field) {
	z.logger.Fatal(msg, z.convertFields(fields)...)
}

// With 创建一个新的子 Logger 实例，该实例在所有日志中都会附加指定的字段。
// 常用于为某个特定的服务或组件创建专用的 Logger。
func (z *ZapLogger) With(fields ...Field) Logger {
	return &ZapLogger{
		logger: z.logger.With(z.convertFields(fields)...),
	}
}

// WithContext 根据上下文创建一个子 Logger 实例。
// 自动从上下文中提取日志字段并附加到返回的 Logger 中。
func (z *ZapLogger) WithContext(ctx context.Context) Logger {
	fields := z.extractContextFields(ctx)
	return z.With(fields...)
}

// Sync 同步日志缓冲区，确保所有待输出的日志都被写入到目标位置。
// 应在程序退出前调用以确保日志不丢失。
func (z *ZapLogger) Sync() error {
	err := z.logger.Sync()
	// 忽略 stdout/stderr 的 sync 错误，这在某些环境中是正常的
	if err != nil {
		errStr := err.Error()
		// 检查是否是 stdout/stderr 的 sync 错误
		if errStr == "sync /dev/stdout: bad file descriptor" ||
			errStr == "sync /dev/stderr: bad file descriptor" ||
			errStr == "sync /dev/stdout: inappropriate ioctl for device" ||
			errStr == "sync /dev/stderr: inappropriate ioctl for device" {
			return nil // 忽略这些错误
		}
	}
	return err
}

// convertFields 将自定义 Field 切片转换为 zap 库的 Field 切片。
func (z *ZapLogger) convertFields(fields []Field) []zap.Field {
	zapFields := make([]zap.Field, len(fields))
	for i, field := range fields {
		zapFields[i] = zap.Any(field.Key, field.Value)
	}
	return zapFields
}

// extractContextFields 从上下文中提取日志字段，如 requestId 等。
func (z *ZapLogger) extractContextFields(ctx context.Context) []Field {
	var fields []Field

	// 提取常见的上下文字段
	if requestID := ctx.Value(RequestIdContextKey); requestID != nil {
		if rid, ok := requestID.(string); ok {
			fields = append(fields, String(RequestIdContextKey, rid))
		}
	}

	return fields
}

// buildZapConfig 构建 zap 配置。
func buildZapConfig(config Config) zap.Config {
	var zapConfig zap.Config = zap.NewProductionConfig()

	zapConfig.Level = zap.NewAtomicLevelAt(getZapLevel(config.Level))

	return zapConfig
}

// getEncoder 获取 zap 编码器。
func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	return zapcore.NewJSONEncoder(encoderConfig)
}

// getZapLevel 将自定义的 Level 转换为 zap 库的 zapcore.Level。
func getZapLevel(level Level) zapcore.Level {
	switch level {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	case FatalLevel:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}
