package logger

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Level = zapcore.Level
type Field = zap.Field
type Option = zap.Option

type Logger struct {
	l     *zap.Logger
	level Level
}

// NewByFileName create a new logger by filename
func NewByFileName(name string, level Level, opts ...Option) *Logger {
	file, err := os.OpenFile("./test.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(fmt.Sprintf("cannot open log file: %v", err))
	}
	return New(file, level, opts...)
}

// New create a new logger (not support log rotating).
func New(writer io.Writer, level Level, opts ...Option) *Logger {
	if writer == nil {
		panic("the writer is nil")
	}

	core := zapcore.NewCore(
		getEncoder(),
		zapcore.AddSync(writer),
		level,
	)
	opts = append(opts, zap.AddCaller())
	return &Logger{
		l:     zap.New(core, opts...),
		level: level,
	}
}

// SizeRotateLogConfig 日志大小分割
type SizeRotateLogConfig struct {
	Level      Level  `json:"level"`       // Level 最低日志等级
	FileName   string `json:"file_name"`   // FileName 日志文件位置
	MaxSize    int    `json:"max_size"`    // MaxSize 进行切割之前，日志文件的最大大小(MB为单位)，默认为100MB
	MaxAge     int    `json:"max_age"`     // MaxAge 是根据文件名中编码的时间戳保留旧日志文件的最大天数。
	MaxBackups int    `json:"max_backups"` // MaxBackups 是要保留的旧日志文件的最大数量。默认是保留所有旧的日志文件（尽管 MaxAge 可能仍会导致它们被删除。）
	Compress   bool   `json:"compress"`    // Compress 是否压缩保存
}

// NewWithSizeRotate create a new logger support log rotating.
func NewWithSizeRotate(lCfg SizeRotateLogConfig, opts ...Option) *Logger {
	opts = append(opts, zap.AddCaller())
	w := GetSizeLogWriter(lCfg.FileName, lCfg.MaxSize, lCfg.MaxBackups, lCfg.MaxAge, lCfg.Compress)
	core := zapcore.NewCore(
		getEncoder(),
		zapcore.AddSync(w),
		lCfg.Level,
	)

	//logger := zap.New(core, opts...)
	//zap.ReplaceGlobals(logger) // 替换zap包中全局的logger实例，后续在其他包中只需使用zap.l()调用即可

	return &Logger{
		l:     zap.New(core, opts...),
		level: lCfg.Level,
	}
}

// GetSizeLogWriter 负责日志写入的位置
func GetSizeLogWriter(filename string, maxsize, maxBackup, maxAge int, compress bool) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   filename,  // 文件位置
		MaxSize:    maxsize,   // 进行切割之前,日志文件的最大大小(MB为单位)
		MaxAge:     maxAge,    // 保留旧文件的最大天数
		MaxBackups: maxBackup, // 保留旧文件的最大个数
		Compress:   compress,  // 是否压缩/归档旧文件
	}
	return zapcore.AddSync(lumberJackLogger)
}

type LevelEnablerFunc func(lvl Level) bool

type TeeOption struct {
	W   io.Writer
	Lef LevelEnablerFunc
}

// NewTee 创建写多个log文件的logger。根据不同的日志级别，写入不同的文件
func NewTee(tops []TeeOption, opts ...Option) *Logger {
	// 需要传入 zap.AddCaller() 才会显示打日志点的文件名和行数
	opts = append(opts, zap.AddCaller())

	var cores []zapcore.Core
	for _, top := range tops {
		if top.W == nil {
			panic("the writer is nil")
		}

		lvl := func(top TeeOption) zapcore.LevelEnabler {
			return zap.LevelEnablerFunc(func(level zapcore.Level) bool {
				return top.Lef(level)
			})
		}(top)

		core := zapcore.NewCore(getEncoder(), zapcore.AddSync(top.W), lvl)
		cores = append(cores, core)
	}
	return &Logger{
		l: zap.New(zapcore.NewTee(cores...), opts...),
	}
}

// 获取配置项
func getEncoder() zapcore.Encoder {
	encodeConfig := zap.NewProductionEncoderConfig()
	encodeConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05"))
	}
	encodeConfig.TimeKey = "timex"
	// 将Level序列化为全大写字符串。例如，将info level序列化为INFO。
	encodeConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	// 以 package/file:行 的格式 序列化调用程序，从完整路径中删除除最后一个目录外的所有目录。
	encodeConfig.EncodeCaller = zapcore.ShortCallerEncoder
	return zapcore.NewJSONEncoder(encodeConfig)
}

var std = New(os.Stderr, zap.InfoLevel)

// 使用method value语法将std实例的各个方法以包级函数的形式暴露给用户，简化用户对logger实例的获取
var (
	Debug   = std.Debug
	Debugf  = std.Debugf
	Info    = std.Info
	Infof   = std.Infof
	Warn    = std.Warn
	Warnf   = std.Warnf
	Error   = std.Error
	Errorf  = std.Errorf
	DPanic  = std.DPanic
	DPanicf = std.DPanicf
	Panic   = std.Panic
	Panicf  = std.Panicf
	Fatal   = std.Fatal
	Fatalf  = std.Fatalf
	ZapL    = std.ZapL
	Sync    = std.Sync
)

// ResetDefault not safe for concurrent use
func ResetDefault(l *Logger) {
	std = l
	Debug = std.Debug
	Debugf = std.Debugf
	Info = std.Info
	Infof = std.Infof
	Warn = std.Warn
	Warnf = std.Warnf
	Error = std.Error
	Errorf = std.Errorf
	DPanic = std.DPanic
	DPanicf = std.DPanicf
	Panic = std.Panic
	Panicf = std.Panicf
	Fatal = std.Fatal
	Fatalf = std.Fatalf
	ZapL = std.ZapL
}

func NewDefaultLogger() *Logger {
	return std
}
