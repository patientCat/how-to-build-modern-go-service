# LogFacade - 日志门面设计模式

LogFacade 是一个日志门面（Facade）设计模式的 Go 语言实现，它隐藏了底层日志库（如 zap）的复杂性，提供统一、简洁的日志接口。

## 设计理念

### 为什么需要日志门面？

1. **解耦业务代码与日志实现**：业务代码不依赖具体的日志库（如 zap、logrus、zerolog）
2. **统一接口**：所有日志调用使用相同的 API，便于维护和理解
3. **灵活切换**：可以在不修改业务代码的情况下更换底层日志实现
4. **标准化字段**：通过 Field 类型提供类型安全的结构化日志

### 架构设计

```
┌─────────────────────────────────────────────────────────┐
│                    业务代码 (Business)                    │
│   userService.logger.Info("用户登录", String("id", "1")) │
└─────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────┐
│                  Logger 接口 (Interface)                  │
│   Info/Debug/Warn/Error/Fatal + Context 方法             │
└─────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────┐
│               ZapLogger 实现 (Implementation)             │
│   底层使用 uber-go/zap 高性能日志库                        │
└─────────────────────────────────────────────────────────┘
```

## 快速开始

### 运行示例

```bash
cd logfacade
go run ./cmd/main.go
```

### 基础用法

```go
package main

import "logfacade"

func main() {
    // 1. 初始化全局日志
    config := logfacade.Config{
        Level:      logfacade.InfoLevel,
        OutputPath: "stdout",
    }
    logfacade.Setup(config)
    defer logfacade.GetDefault().Sync()

    // 2. 使用结构化字段记录日志
    logfacade.GetDefault().Info("应用启动",
        logfacade.String("version", "1.0.0"),
        logfacade.Int("port", 8080),
    )
}
```

### 依赖注入模式（推荐）

```go
// 服务定义
type UserService struct {
    logger logfacade.Logger
}

func NewUserService(logger logfacade.Logger) *UserService {
    return &UserService{
        logger: logger.With(logfacade.String("service", "UserService")),
    }
}

func (s *UserService) CreateUser(name string) {
    s.logger.Info("创建用户", logfacade.String("name", name))
}

// 使用
func main() {
    logger, _ := logfacade.New(logfacade.Config{
        Level:      logfacade.InfoLevel,
        OutputPath: "stdout",
    })
    
    userService := NewUserService(logger)
    userService.CreateUser("alice")
}
```

## 核心功能

### 1. 结构化字段

```go
logger.Info("用户登录",
    logfacade.String("username", "john"),
    logfacade.Int("user_id", 123),
    logfacade.Bool("is_admin", false),
    logfacade.Duration("session_timeout", 30*time.Minute),
    logfacade.Err(err),
    logfacade.Any("metadata", map[string]string{"key": "value"}),
)
```

### 2. 日志级别

```go
logger.Debug("调试信息")  // 仅在 Debug 级别时输出
logger.Info("普通信息")   // Info 及以上级别输出
logger.Warn("警告信息")   // Warn 及以上级别输出
logger.Error("错误信息")  // Error 及以上级别输出
logger.Fatal("致命错误")  // 输出后程序退出
```

### 3. 上下文感知

```go
ctx := context.WithValue(context.Background(), "request_id", "req-123")
ctx = context.WithValue(ctx, "trace_id", "trace-abc")

logger.InfoContext(ctx, "处理请求")  // 自动附加 request_id 和 trace_id
```

### 4. 子 Logger

```go
// 为特定组件创建专用 logger
authLogger := logger.With(
    logfacade.String("component", "auth"),
    logfacade.String("module", "oauth"),
)

authLogger.Info("用户认证成功")  // 自动带上 component 和 module 字段
```

### 5. 日志轮转

```go
config := logfacade.Config{
    Level:      logfacade.InfoLevel,
    OutputPath: "./logs/app.log",
    MaxSize:    100,    // 单文件最大 100MB
    MaxBackups: 5,      // 保留 5 个备份
    MaxAge:     30,     // 保留 30 天
    Compress:   true,   // 压缩旧文件
}
```

## 文件结构

```
logfacade/
├── cmd/
│   └── main.go           # 演示程序入口
├── interface.go          # Logger 接口和 Field 定义
├── logger.go             # Logger 工厂函数
├── zap_logger.go         # Zap 实现
├── global.go             # 全局日志实例管理
├── constant.go           # 上下文 Key 常量
├── example_test.go       # 测试用例
├── go.mod                # Go 模块定义
└── README.md             # 本文档
```

## 设计原则

1. **接口隔离**：业务代码只依赖 `Logger` 接口，不依赖具体实现
2. **依赖注入**：推荐通过构造函数注入 Logger，而非使用全局变量
3. **全局仅用于启动**：`GetDefault()` 仅在应用启动阶段使用
4. **结构化优先**：优先使用结构化字段而非字符串拼接
5. **上下文传递**：在请求链路中传递上下文，实现日志追踪

## 测试

```bash
# 运行所有测试
go test -v ./...

# 运行性能测试
go test -bench=. -benchmem
```
