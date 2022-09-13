package ezap

import "fmt"

var c *Logger

// autonew logger
func init() {
	c = New()
}

func New() *Logger {
	logger := &Logger{
		Config: newConfig(),
	}
	logger.syncConfig()
	return logger
}

func Print(args ...interface{}) {
	fmt.Print(args...)
}
func Println(args ...interface{}) {
	fmt.Println(args...)
}
func Printf(template string, args ...interface{}) {
	fmt.Printf(template, args...)
}

func Debug(args ...interface{}) {
	c.Logger.Debug(args...)
}
func Debugf(template string, args ...interface{}) {
	c.Logger.Debugf(template, args...)
}
func Debugw(msg string, keysAndValues ...interface{}) {
	c.Logger.Debugw(msg, keysAndValues...)
}

func Info(args ...interface{}) {
	c.Logger.Info(args...)
}
func Infof(template string, args ...interface{}) {
	c.Logger.Infof(template, args...)
}
func Infow(msg string, keysAndValues ...interface{}) {
	c.Logger.Infow(msg, keysAndValues...)
}

func Warn(args ...interface{}) {
	c.Logger.Warn(args...)
}
func Warnf(template string, args ...interface{}) {
	c.Logger.Warnf(template, args...)
}
func Warnw(msg string, keysAndValues ...interface{}) {
	c.Logger.Warnw(msg, keysAndValues...)
}

func Error(args ...interface{}) {
	c.Logger.Error(args...)
}
func Errorf(template string, args ...interface{}) {
	c.Logger.Errorf(template, args...)
}
func Errorw(msg string, keysAndValues ...interface{}) {
	c.Logger.Errorw(msg, keysAndValues...)
}

func Fatal(args ...interface{}) {
	c.Logger.Fatal(args...)
}
func Fatalf(template string, args ...interface{}) {
	c.Logger.Fatalf(template, args...)
}
func Fatalw(msg string, keysAndValues ...interface{}) {
	c.Logger.Fatalw(msg, keysAndValues...)
}

func Panic(args ...interface{}) {
	c.Logger.Panic(args...)
}
func Panicf(template string, args ...interface{}) {
	c.Logger.Panicf(template, args...)
}
func Panicw(msg string, keysAndValues ...interface{}) {
	c.Logger.Panicw(msg, keysAndValues...)
}
