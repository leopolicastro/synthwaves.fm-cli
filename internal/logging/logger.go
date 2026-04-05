package logging

import (
	"os"

	"github.com/charmbracelet/log"
)

var Logger *log.Logger

func init() {
	Logger = log.NewWithOptions(os.Stderr, log.Options{
		ReportTimestamp: false,
		Prefix:          "synthwaves",
	})
	Logger.SetStyles(synthwaveStyles())
}

func synthwaveStyles() *log.Styles {
	s := log.DefaultStyles()
	return s
}

func Debug(msg string, keyvals ...interface{}) { Logger.Debug(msg, keyvals...) }
func Info(msg string, keyvals ...interface{})  { Logger.Info(msg, keyvals...) }
func Warn(msg string, keyvals ...interface{})  { Logger.Warn(msg, keyvals...) }
func Error(msg string, keyvals ...interface{}) { Logger.Error(msg, keyvals...) }
func Fatal(msg string, keyvals ...interface{}) { Logger.Fatal(msg, keyvals...) }

func SetVerbose(v bool) {
	if v {
		Logger.SetLevel(log.DebugLevel)
	} else {
		Logger.SetLevel(log.InfoLevel)
	}
}
