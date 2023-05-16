package logger

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
)

type writerHook struct {
	Writer    []io.Writer
	LogLevels []logrus.Level
}

func (hook *writerHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}

	for _, w := range hook.Writer {
		w.Write([]byte(line))
	}

	return err
}

func (hook *writerHook) Levels() []logrus.Level {
	return hook.LogLevels
}

var e *logrus.Entry

type Logger struct {
	*logrus.Entry
}

func Get() Logger {
	return Logger{e}
}

func (l Logger) GetWithField(k string, v interface{}) Logger {
	return Logger{l.WithField(k, v)}
}

func Init(level, filePath string) error {
	if e != nil {
		return errors.New("Логгер уже инициализирован")
	}

	l := logrus.New()
	l.SetReportCaller(true)
	l.Formatter = &logrus.TextFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (function string, file string) {
			filename := path.Base(f.File)
			return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%d", filename, f.Line)
		},
		FullTimestamp: true,
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0644); err != nil {
		panic(err)
	}

	logFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}

	l.SetOutput(io.Discard)

	l.AddHook(&writerHook{
		Writer:    []io.Writer{logFile, os.Stdout},
		LogLevels: logrus.AllLevels,
	})

	if logrusLevel, err := logrus.ParseLevel(level); err != nil {
		panic(err)
	} else {
		l.SetLevel(logrusLevel)
	}

	e = logrus.NewEntry(l)

	return nil
}

func (l Logger) DebugOrError(err error, args ...interface{}) {
	if err != nil {
		l.Error(args...)
	} else {
		l.Debug(args...)
	}
}
