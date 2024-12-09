package config

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

var Log *logrus.Logger

type FileHook struct {
	fullWriter  io.Writer
	errorWriter io.Writer
	formatter   logrus.Formatter
}

func NewFileHook(fullPath, errorPath string) (*FileHook, error) {
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return nil, err
	}

	fullFile, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	errorFile, err := os.OpenFile(errorPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fullFile.Close()
		return nil, err
	}

	return &FileHook{
		fullWriter:  fullFile,
		errorWriter: errorFile,
		formatter: &logrus.TextFormatter{
			DisableColors:          true,
			TimestampFormat:        "2006-01-02 15:04:05",
			FullTimestamp:          true,
			DisableQuote:           false, // 确保字符串被正确引用
			QuoteEmptyFields:       true,  // 空字段也加引号
			DisableSorting:         false, // 允许字段排序
			PadLevelText:           true,  // 保持等级文本对齐
			DisableLevelTruncation: false, // 显示完整的等级名称
			CallerPrettyfier: func(f *runtime.Frame) (function string, file string) {
				return "", fmt.Sprintf("%s:%d", filepath.Base(f.File), f.Line)
			},
		},
	}, nil
}

func (hook *FileHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook *FileHook) Fire(entry *logrus.Entry) error {
	line, err := hook.formatter.Format(entry)
	if err != nil {
		return err
	}

	_, err = hook.fullWriter.Write(line)
	if err != nil {
		return err
	}

	if entry.Level <= logrus.ErrorLevel {
		_, err = hook.errorWriter.Write(line)
		if err != nil {
			return err
		}
	}

	return nil
}

func init() {
	Log = logrus.New()

	Log.Out = os.Stdout
	Log.SetLevel(logrus.DebugLevel)
	Log.SetReportCaller(true)

	// 更新格式化配置
	Log.SetFormatter(&logrus.TextFormatter{
		ForceColors:            true,
		DisableQuote:           false, // 确保字符串被正确引用
		QuoteEmptyFields:       true,  // 空字段也加引号
		DisableSorting:         false, // 允许字段排序
		PadLevelText:           true,  // 保持等级文本对齐
		DisableLevelTruncation: false, // 显示完整的等级名称
		CallerPrettyfier: func(f *runtime.Frame) (function string, file string) {
			return "", fmt.Sprintf("%s:%d", filepath.Base(f.File), f.Line)
		},
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})

	fileHook, err := NewFileHook("logs/full.log", "logs/error.log")
	if err != nil {
		Log.Fatal("Failed to setup file hook:", err)
	}
	Log.AddHook(fileHook)
}

func GetLogger() *logrus.Logger {
	return Log
}
