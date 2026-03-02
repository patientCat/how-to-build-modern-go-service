// Package main 演示 logfacade 包的使用方法
//
// logfacade 是一个日志门面（Facade）设计模式的实现，
// 它隐藏了底层日志库（如 zap）的复杂性，提供统一、简洁的日志接口。
//
// 运行方式:
//
//	cd logfacade
//	go run ./cmd/main.go
package main

import (
	"context"
	"fmt"
	"time"

	"logfacade"
)

func main() {
	fmt.Println("=== LogFacade 日志门面设计模式演示 ===")
	fmt.Println()

	// 1. 基础用法：使用全局 Setup 初始化
	fmt.Println("【1. 基础用法：全局日志初始化】")
	basicUsage()

	// 2. 结构化日志字段
	fmt.Println("\n【2. 结构化日志字段】")
	structuredLogging()

	// 3. 上下文感知日志
	fmt.Println("\n【3. 上下文感知日志】")
	contextualLogging()

	// 4. 子 Logger（With 方法）
	fmt.Println("\n【4. 子 Logger - 组件专用日志】")
	childLoggerDemo()

	// 5. 日志级别演示
	fmt.Println("\n【5. 日志级别演示】")
	logLevelDemo()

	// 6. 写入文件
	fmt.Println("\n【6. 日志写入文件】")
	fileLogging()

	// 7. 依赖注入模式
	fmt.Println("\n【7. 依赖注入模式】")
	dependencyInjectionDemo()

	fmt.Println("\n=== 演示完成 ===")
}

// basicUsage 演示基础的全局日志初始化和使用
func basicUsage() {
	// 使用 Setup 初始化全局日志
	config := logfacade.Config{
		Level:      logfacade.InfoLevel,
		OutputPath: "stdout",
	}

	if err := logfacade.Setup(config); err != nil {
		panic(err)
	}

	// 使用全局默认 logger（仅在初始化阶段推荐）
	logfacade.GetDefault().Info("应用启动成功",
		logfacade.String("version", "1.0.0"),
		logfacade.String("env", "development"),
	)
}

// structuredLogging 演示结构化日志字段的使用
func structuredLogging() {
	logger, _ := logfacade.New(logfacade.Config{
		Level:      logfacade.InfoLevel,
		OutputPath: "stdout",
	})

	// 字符串字段
	logger.Info("用户登录",
		logfacade.String("username", "john_doe"),
		logfacade.String("ip", "192.168.1.100"),
	)

	// 数字字段
	logger.Info("订单创建",
		logfacade.Int("order_id", 12345),
		logfacade.Int64("total_cents", 9999),
		logfacade.Float64("discount", 0.15),
	)

	// 布尔字段
	logger.Info("功能开关",
		logfacade.Bool("feature_enabled", true),
	)

	// 时间间隔字段
	logger.Info("请求处理完成",
		logfacade.Duration("elapsed", 150*time.Millisecond),
	)

	// 任意类型字段
	type User struct {
		ID   int
		Name string
	}
	logger.Info("用户信息",
		logfacade.Any("user", User{ID: 1, Name: "Alice"}),
	)

	// 错误字段
	logger.Error("操作失败",
		logfacade.Err(fmt.Errorf("database connection timeout")),
	)
}

// contextualLogging 演示上下文感知日志
func contextualLogging() {
	logger, _ := logfacade.New(logfacade.Config{
		Level:      logfacade.InfoLevel,
		OutputPath: "stdout",
	})

	// 模拟 HTTP 请求上下文
	ctx := context.Background()
	ctx = context.WithValue(ctx, "request_id", "req-abc-123")
	ctx = context.WithValue(ctx, "trace_id", "trace-xyz-789")

	// 使用 WithContext 方法创建带上下文的子 Logger
	// 这种方式比每次调用都传递 ctx 更简洁，且支持链式调用
	ctxLogger := logger.WithContext(ctx)

	ctxLogger.Info("开始处理请求")
	ctxLogger.Info("查询数据库",
		logfacade.String("table", "users"),
		logfacade.Int("limit", 10),
	)
	ctxLogger.Info("请求处理完成",
		logfacade.Int("status_code", 200),
	)
}

// childLoggerDemo 演示使用 With 创建子 Logger
func childLoggerDemo() {
	baseLogger, _ := logfacade.New(logfacade.Config{
		Level:      logfacade.InfoLevel,
		OutputPath: "stdout",
	})

	// 为不同组件创建专用 logger
	userServiceLogger := baseLogger.With(
		logfacade.String("component", "UserService"),
	)

	orderServiceLogger := baseLogger.With(
		logfacade.String("component", "OrderService"),
	)

	// 每个组件的日志都会自动带上组件标识
	userServiceLogger.Info("用户注册", logfacade.String("email", "user@example.com"))
	orderServiceLogger.Info("订单创建", logfacade.Int("order_id", 1001))

	// 子 logger 还可以继续派生
	userAuthLogger := userServiceLogger.With(
		logfacade.String("module", "authentication"),
	)
	userAuthLogger.Info("用户认证成功", logfacade.String("method", "OAuth2"))
}

// logLevelDemo 演示不同日志级别
func logLevelDemo() {
	// 创建 Debug 级别的 logger（会输出所有级别）
	logger, _ := logfacade.New(logfacade.Config{
		Level:      logfacade.DebugLevel,
		OutputPath: "stdout",
	})

	logger.Debug("调试信息：变量状态",
		logfacade.String("state", "initialized"),
	)
	logger.Info("信息：正常运行事件",
		logfacade.String("event", "startup"),
	)
	logger.Warn("警告：潜在问题",
		logfacade.String("issue", "high_memory_usage"),
	)
	logger.Error("错误：运行时错误",
		logfacade.Err(fmt.Errorf("connection refused")),
	)
	// logger.Fatal("致命错误：程序将退出") // 注意：Fatal 会导致程序退出

	// 格式化日志方法
	logger.Infof("格式化日志：用户 %s 登录成功，ID=%d", "Alice", 123)
	logger.Warnf("格式化警告：内存使用率 %.2f%%", 85.5)
}

// fileLogging 演示日志写入文件
func fileLogging() {
	config := logfacade.Config{
		Level:      logfacade.InfoLevel,
		OutputPath: "./logs/app.log", // 写入文件
		MaxSize:    10,               // 单文件最大 10MB
		MaxBackups: 3,                // 保留 3 个备份
		MaxAge:     7,                // 保留 7 天
		Compress:   true,             // 压缩旧文件
	}

	logger, err := logfacade.New(config)
	if err != nil {
		fmt.Printf("创建文件日志失败: %v\n", err)
		return
	}
	defer logger.Sync()

	logger.Info("这条日志会写入文件",
		logfacade.String("file", "./logs/app.log"),
	)

	fmt.Println("日志已写入 ./logs/app.log")
}

// dependencyInjectionDemo 演示依赖注入模式
func dependencyInjectionDemo() {
	// 创建 logger 实例
	logger, _ := logfacade.New(logfacade.Config{
		Level:      logfacade.InfoLevel,
		OutputPath: "stdout",
	})

	// 通过依赖注入传递给服务
	userService := NewUserService(logger)
	userService.CreateUser("alice", "alice@example.com")
}

// UserService 示例服务，演示依赖注入
type UserService struct {
	logger logfacade.Logger
}

// NewUserService 创建 UserService，接收 Logger 依赖
func NewUserService(logger logfacade.Logger) *UserService {
	return &UserService{
		logger: logger.With(logfacade.String("service", "UserService")),
	}
}

// CreateUser 创建用户
func (s *UserService) CreateUser(username, email string) {
	s.logger.Info("创建用户",
		logfacade.String("username", username),
		logfacade.String("email", email),
	)
}
