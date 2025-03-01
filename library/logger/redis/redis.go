package redis

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/why444216978/gin-api/library/logger"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

type contextKey int

const (
	cmdStart contextKey = iota
)

type RedisConfig struct {
	InfoFile    string
	ErrorFile   string
	Level       string
	ServiceName string
	Host        string
	Port        int
}

// RedisLogger is go-redis logger Hook
type RedisLogger struct {
	*logger.Logger
	Config RedisConfig
}

type RedisOption func(rl *RedisLogger)

// NewRedisLogger
func NewRedisLogger(cfg *RedisConfig, opts ...RedisOption) (rl *RedisLogger, err error) {
	rl = &RedisLogger{
		Config: *cfg,
	}

	for _, o := range opts {
		o(rl)
	}

	l, err := logger.NewLogger(&logger.Config{
		InfoFile:  cfg.InfoFile,
		ErrorFile: cfg.ErrorFile,
		Level:     cfg.Level,
	}, logger.WithModule(logger.ModuleRedis), logger.WithServiceName(cfg.ServiceName), logger.WithCallerSkip(5))
	if err != nil {
		return
	}
	rl.Logger = l

	return
}

//BeforeProcess redis before execute action do something
func (rl *RedisLogger) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	ctx = rl.setCmdStart(ctx)
	return ctx, nil
}

//AfterProcess redis after execute action do something
func (rl *RedisLogger) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	if rl.Logger == nil {
		return nil
	}

	cost := rl.getCmdCost(ctx)
	if err := cmd.Err(); err != nil && err != redis.Nil {
		rl.Error(ctx, false, []redis.Cmder{cmd}, cost)
		return nil
	}

	rl.Info(ctx, false, []redis.Cmder{cmd}, cost)

	return nil
}

// BeforeProcessPipeline before command process handle
func (rl *RedisLogger) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	ctx = rl.setCmdStart(ctx)
	return ctx, nil
}

// AfterProcessPipeline after command process handle
func (rl *RedisLogger) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	if rl.Logger == nil {
		return nil
	}
	cost := rl.getCmdCost(ctx)

	hasErr := false
	for _, cmd := range cmds {
		if err := cmd.Err(); err != nil && err != redis.Nil {
			hasErr = true
		}
	}

	if hasErr {
		rl.Error(ctx, true, cmds, cost)
		return nil
	}
	rl.Info(ctx, true, cmds, cost)

	return nil
}

func (rl *RedisLogger) Info(ctx context.Context, isPipeline bool, cmds []redis.Cmder, cost int64) {
	newCtx, logFields := rl.fields(ctx, isPipeline, cmds, cost)
	rl.Logger.Info(newCtx, "info", logFields...)
}

func (rl *RedisLogger) Error(ctx context.Context, isPipeline bool, cmds []redis.Cmder, cost int64) {
	errs := make([]string, 0)
	for idx, cmd := range cmds {
		err := cmd.Err()
		if err == nil {
			return
		}
		errs = append(errs, strconv.Itoa(idx)+"-"+err.Error())
	}
	newCtx, logFields := rl.fields(ctx, isPipeline, cmds, cost)
	rl.Logger.Error(newCtx, strings.Join(errs, ","), logFields...)
}

func (rl *RedisLogger) fields(ctx context.Context, isPipeline bool, cmds []redis.Cmder, cost int64) (context.Context, []zap.Field) {
	l := len(cmds)
	names := make([]string, l)
	args := make([]interface{}, l)
	response := make([]string, l)
	for idx, cmd := range cmds {
		names[idx] = cmd.Name()
		args[idx] = cmd.Args()
		response[idx] = cmd.String()
	}

	method := "pipeline"
	if !isPipeline {
		method = cmds[0].Name()
	}

	logFields := logger.ValueHTTPFields(ctx)
	logFields.Header = http.Header{}
	logFields.Method = method
	logFields.Request = args
	logFields.Response = response
	logFields.Code = 0
	logFields.ClientIP = logFields.ServerIP
	logFields.ClientPort = logFields.ServerPort
	logFields.ServerIP = rl.Config.Host
	logFields.ServerPort = rl.Config.Port
	logFields.API = method
	logFields.Cost = cost

	newCtx := context.WithValue(ctx, "rpc", "rpc")
	newCtx = logger.WithHTTPFields(newCtx, logFields)
	return newCtx, []zap.Field{}
}

func (rl *RedisLogger) setCmdStart(ctx context.Context) context.Context {
	return context.WithValue(ctx, cmdStart, time.Now())
}

func (rl *RedisLogger) getCmdCost(ctx context.Context) int64 {
	return time.Since(ctx.Value(cmdStart).(time.Time)).Milliseconds()
}
