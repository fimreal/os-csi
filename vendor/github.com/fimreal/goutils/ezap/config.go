package ezap

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
	LogLevel    zap.AtomicLevel
	ProjectName string
	JSONFormat  bool
	Console     bool
	AddCaller   bool
	LogFile     *Logfile
	TimeFormat  string
}

type Logfile struct {
	FileName   string
	MaxSize    int
	MaxAge     int
	MaxBackups int
	Compress   bool
}

type Logger struct {
	Config *Config
	Logger *zap.SugaredLogger
}

func newConfig() *Config {
	return &Config{
		LogLevel:    zap.NewAtomicLevel(),
		ProjectName: "",
		JSONFormat:  false,
		Console:     true,
		AddCaller:   false,
		LogFile: &Logfile{
			FileName:   "",
			MaxSize:    100,
			MaxAge:     24,
			MaxBackups: 7,
			Compress:   true,
		},
		TimeFormat: "2006-1-2T15:04:05.000Z0700",
	}
}

// 应用修改后的配置
func (l *Logger) syncConfig() {
	conf := l.Config
	cores := []zapcore.Core{}
	var logger *zap.Logger
	var encoder zapcore.Encoder
	// cfg := zap.NewProductionConfig()
	// encoderConfig := cfg.EncoderConfig
	encoderConfig := zap.NewProductionConfig().EncoderConfig
	encoderConfig.LevelKey = "lv"
	encoderConfig.NameKey = "pj"
	if conf.TimeFormat == "" {
		encoderConfig.EncodeTime = nil
	} else {
		encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format(conf.TimeFormat))
		}
	}

	if conf.JSONFormat {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		// 终端日志级别显示颜色
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	if conf.Console {
		writer := zapcore.Lock(os.Stdout)
		core := zapcore.NewCore(encoder, writer, conf.LogLevel)
		cores = append(cores, core)
	}

	if conf.LogFile.FileName != "" {
		w := zapcore.AddSync(&lumberjack.Logger{
			Filename:   conf.LogFile.FileName,
			MaxSize:    conf.LogFile.MaxSize,
			MaxAge:     conf.LogFile.MaxAge,
			MaxBackups: conf.LogFile.MaxBackups,
			Compress:   conf.LogFile.Compress,
			LocalTime:  true,
		})
		writer := zapcore.AddSync(w)
		core := zapcore.NewCore(encoder, writer, conf.LogLevel)
		cores = append(cores, core)
	}

	if conf.AddCaller {
		logger = zap.New(zapcore.NewTee(cores...), zap.AddCaller())
	} else {
		logger = zap.New(zapcore.NewTee(cores...))
	}

	if conf.ProjectName != "" {
		logger = logger.Named(conf.ProjectName)
	}

	defer logger.Sync()
	l.Logger = logger.Sugar()
}

/*
	**设置输出的日志等级

Options: -1(info), 0(debug), 1(warn), 2(error), 3(dpanic), 4(panic), 5(fatal)
Default: -1(info)
*/
func SetLevel(lv string) {
	setlevel := c.Config.LogLevel.SetLevel
	switch lv {
	case "debug":
		// zap.DebugLevel zapcore.Level = -1
		setlevel(-1)
	case "info":
		// zap.InfoLevel zapcore.Level = 0
		setlevel(0)
	case "warn":
		// zap.WarnLevel zapcore.Level = 1
		setlevel(1)
	case "error":
		// zap.ErrorLevel zapcore.Level = 2
		setlevel(2)
	case "dpanic":
		// zap.DPanicLevel zapcore.Level = 3
		setlevel(3)
	case "panic":
		// zap.PanicLevel zapcore.Level = 4
		setlevel(4)
	case "fatal":
		// zap.FatalLevel zapcore.Level = 5
		setlevel(5)
	default:
		setlevel(0)
	}
	c.syncConfig()
}

// 设置工程名称
func SetProjectName(projectname string) {
	c.Config.ProjectName = projectname
	c.syncConfig()
}

// 启用 JSON 格式输出
func EnableJSONFormat() {
	c.Config.JSONFormat = true
	c.syncConfig()
}

// 关闭 JSON 格式输出
func DisableJSONFormat() {
	c.Config.JSONFormat = false
	c.syncConfig()
}

// 开启调用日志输出
func EnableCaller() {
	c.Config.AddCaller = true
	c.syncConfig()
}

// 关闭调用日志输出
func DisableCaller() {
	c.Config.AddCaller = false
	c.syncConfig()
}

// 开启控制台日志输出
func EnableConsole() {
	c.Config.Console = true
	c.syncConfig()
}

// 关闭控制台日志输出
func DisableConsole() {
	c.Config.Console = false
	c.syncConfig()
}

// 设置日志保存文件，例如 /var/log/myapp/myapp.log
// 如果配置为空，则不会保存到日志文件
func SetLogFile(name string) {
	c.Config.LogFile.FileName = name
	c.syncConfig()
}

// 配置日志滚动配置，默认 100M，24h，7天，启用压缩
func SetLogrotate(maxsize, maxage, maxbackups int, compress bool) {
	f := c.Config.LogFile
	f.MaxSize = maxsize
	f.MaxAge = maxage
	f.MaxBackups = maxbackups
	f.Compress = compress
	c.syncConfig()
}

// 配置时间格式，默认为 2006-1-2T15:04:05.000Z0700
func SetLogTime(timeFormat string) {
	c.Config.TimeFormat = timeFormat
	c.syncConfig()
}
