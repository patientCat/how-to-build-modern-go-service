package logfacade

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

// Logger 定义了日志记录接口，隐藏底层实现细节，支持多种日志库的无缝切换。
// 该接口提供了结构化日志记录的完整功能，包括基础日志方法、上下文感知日志、
// 以及创建子 Logger 的能力。
//
// 上下文感知日志的推荐用法：
//
//	logger.WithContext(ctx).Info("处理请求", logfacade.String("action", "create"))
type Logger interface {
	// Debug 记录调试级别的日志信息。调试级别的日志仅在调试模式下输出。
	Debug(msg string, fields ...Field)

	// Info 记录信息级别的日志信息。信息级别适合记录常规的应用事件。
	Info(msg string, fields ...Field)

	// Warn 记录警告级别的日志信息。警告级别用于记录潜在的问题。
	Warn(msg string, fields ...Field)

	// Error 记录错误级别的日志信息。错误级别用于记录运行时错误。
	Error(msg string, fields ...Field)

	// Fatal 记录致命错误日志并立即退出程序。仅在不可恢复的错误发生时使用。
	Fatal(msg string, fields ...Field)

	// With 创建一个新的子 Logger 实例，该实例在所有日志中都会附加指定的字段。
	// 常用于为某个特定的服务或组件创建专用的 Logger。
	With(fields ...Field) Logger

	// WithContext 根据上下文创建一个子 Logger 实例，
	// 自动从上下文中提取日志字段（如 request_id、trace_id）并附加到返回的 Logger 中。
	// 推荐用法：logger.WithContext(ctx).Info("message")
	WithContext(ctx context.Context) Logger

	// Sync 同步日志缓冲区，确保所有待输出的日志都被写入到目标位置。
	// 应在程序退出前调用以确保日志不丢失。
	Sync() error

	// Infof 使用格式化字符串记录信息级别的日志。
	// 参数与 fmt.Sprintf 兼容。
	Infof(format string, args ...any)

	// Warnf 使用格式化字符串记录警告级别的日志。
	// 参数与 fmt.Sprintf 兼容。
	Warnf(format string, args ...any)

	// DeepInfof 使用格式化字符串和深度展开记录信息级别的日志。
	// 对复杂数据结构进行深度展开，便于调试。
	DeepInfof(msg string, args ...any)

	// Errorf 使用格式化字符串记录错误级别的日志。
	// 参数与 fmt.Sprintf 兼容。
	Errorf(format string, args ...any)
}

// Field 表示一个日志字段，由字段名和值组成。
// 日志字段用于为日志消息提供结构化的上下文信息。
type Field struct {
	// Key 是字段的名称
	Key string
	// Value 是字段的值，可以是任意类型
	Value interface{}
}

// Level 表示日志级别，用于控制日志的详细程度。
type Level string

const (
	// DebugLevel 调试级别，用于开发调试，包含最详细的信息
	DebugLevel Level = "debug"
	// InfoLevel 信息级别，用于记录常规的应用事件
	InfoLevel = "info"
	// WarnLevel 警告级别，用于记录潜在的问题
	WarnLevel = "warn"
	// ErrorLevel 错误级别，用于记录运行时错误
	ErrorLevel = "error"
	// FatalLevel 致命错误级别，用于不可恢复的错误，会导致程序退出
	FatalLevel = "fatal"
)

// Config 包含日志记录器的配置选项。
type Config struct {
	// Level 设置日志的最低输出级别，低于该级别的日志不会被输出
	Level Level `json:"level" mapstructure:"level"`
	// OutputPath 日志输出路径，可以是 "stdout" 或具体的文件路径
	OutputPath string `json:"output_path" mapstructure:"output_path"`
	// MaxSize 日志文件的最大大小，单位为 MB，超过此大小时日志文件会轮转
	MaxSize int `json:"max_size" mapstructure:"max_size"`
	// MaxBackups 保留的旧日志文件数量
	MaxBackups int `json:"max_backups" mapstructure:"max_backups"`
	// MaxAge 保留旧日志文件的最大天数
	MaxAge int `json:"max_age" mapstructure:"max_age"`
	// Compress 是否压缩被轮转的旧日志文件
	Compress bool `json:"compress" mapstructure:"compress"`
}

// String 创建一个字符串类型的日志字段。
//
// 参数：
//   - key: 字段名称
//   - val: 字符串值
//
// 返回：
//   - Field: 包含键值对的日志字段
//
// 示例：
//
//	logger.Info("用户登录", logfacade.String("username", "john"))
func String(key, val string) Field {
	return Field{Key: key, Value: val}
}

// Int 创建一个整数类型的日志字段。
//
// 参数：
//   - key: 字段名称
//   - val: 整数值
//
// 返回：
//   - Field: 包含键值对的日志字段
//
// 示例：
//
//	logger.Info("用户信息", logfacade.Int("user_id", 123))
func Int(key string, val int) Field {
	return Field{Key: key, Value: val}
}

// Int64 创建一个 64 位整数类型的日志字段。
//
// 参数：
//   - key: 字段名称
//   - val: 64 位整数值
//
// 返回：
//   - Field: 包含键值对的日志字段
//
// 示例：
//
//	logger.Info("统计信息", logfacade.Int64("total_count", 9223372036854775807))
func Int64(key string, val int64) Field {
	return Field{Key: key, Value: val}
}

// Float64 创建一个浮点数类型的日志字段。
//
// 参数：
//   - key: 字段名称
//   - val: 浮点数值
//
// 返回：
//   - Field: 包含键值对的日志字段
//
// 示例：
//
//	logger.Info("性能指标", logfacade.Float64("response_time", 0.123))
func Float64(key string, val float64) Field {
	return Field{Key: key, Value: val}
}

// Bool 创建一个布尔类型的日志字段。
//
// 参数：
//   - key: 字段名称
//   - val: 布尔值
//
// 返回：
//   - Field: 包含键值对的日志字段
//
// 示例：
//
//	logger.Info("系统状态", logfacade.Bool("is_running", true))
func Bool(key string, val bool) Field {
	return Field{Key: key, Value: val}
}

// Duration 创建一个时间间隔类型的日志字段。
//
// 参数：
//   - key: 字段名称
//   - val: 时间间隔值
//
// 返回：
//   - Field: 包含键值对的日志字段
//
// 示例：
//
//	logger.Info("请求耗时", logfacade.Duration("elapsed", 150*time.Millisecond))
func Duration(key string, val time.Duration) Field {
	return Field{Key: key, Value: val}
}

// Err 创建一个错误类型的日志字段。
// 该函数自动使用 "error" 作为字段名，将错误信息作为值。
//
// 参数：
//   - err: 错误对象
//
// 返回：
//   - Field: 包含键值对的日志字段，其中 Key 为 "error"
//
// 示例：
//
//	logger.Error("操作失败", logfacade.Err(err))
func Err(err error) Field {
	// 如果有堆栈信息
	if _, ok := err.(interface{ StackTrace() errors.StackTrace }); ok {
		return Field{Key: "error", Value: fmt.Sprintf("%+v", err)}
	}

	// 普通错误
	return Field{Key: "error", Value: err}
}

// Any 创建一个任意类型的日志字段。
// 该函数用于记录不属于上述任何预定义类型的值。
//
// 参数：
//   - key: 字段名称
//   - val: 任意类型的值
//
// 返回：
//   - Field: 包含键值对的日志字段
//
// 示例：
//
//	logger.Info("复杂数据", logfacade.Any("user", user))
func Any(key string, val interface{}) Field {
	return Field{Key: key, Value: val}
}
