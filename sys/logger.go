package sys

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
)

var Logger *logrus.Logger
var LogFile *os.File

func (hook *FileHook) Levels() []logrus.Level {
	return hook.LogLevels
}

func InitLogger(path string) {
	Logger = logrus.New()

	infoFile := &lumberjack.Logger{
		Filename:   path + "/info.log", // 日志文件的位置
		MaxSize:    100,                // 每个日志文件保存的最大尺寸, 单位：MB
		MaxBackups: 30,                 // 日志文件最多保存多少个备份(最近多少个文件)
		MaxAge:     30,                 // 文件最多保存多少天
		Compress:   true,               // 是否压缩/归档旧文件
	}
	errorFile := &lumberjack.Logger{
		Filename:   path + "/error.log", // 日志文件的位置
		MaxSize:    100,                 // 每个日志文件保存的最大尺寸, 单位：MB
		MaxBackups: 30,                  // 日志文件最多保存多少个备份(最近多少个文件)
		MaxAge:     30,                  // 文件最多保存多少天
		Compress:   true,                // 是否压缩/归档旧文件
	}
	Logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.999",
	})
	Logger.SetLevel(logrus.InfoLevel)
	Logger.SetReportCaller(true) // Enable reporting caller info

	Logger.SetOutput(os.Stdout)

	Logger.AddHook(&FileHook{
		Writer: infoFile,
		LogLevels: []logrus.Level{
			logrus.InfoLevel,
			logrus.ErrorLevel,
			logrus.WarnLevel,
		},
	})

	Logger.AddHook(&FileHook{
		Writer: errorFile,
		LogLevels: []logrus.Level{
			logrus.ErrorLevel,
			logrus.FatalLevel,
			logrus.PanicLevel,
		},
	})
	Logger.Info("Logger initialized")
}

type FileHook struct {
	Writer    io.Writer
	LogLevels []logrus.Level
}

func (hook *FileHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}
	_, err = hook.Writer.Write([]byte(line))
	//if err != nil {
	//	fmt.Printf("Error writing log: %v\n", err)
	//} else {
	//	fmt.Printf("Successfully wrote log entry: %s\n", line)
	//}
	return err
}
