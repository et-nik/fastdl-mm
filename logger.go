package main

import (
	metamod "github.com/et-nik/metamod-go"
	"strings"
)

type MetaLogWriter struct {
	MetaUtilFuncs *metamod.MUtilFuncs
}

func NewMetaLogWriter(metaUtilFuncs *metamod.MUtilFuncs) *MetaLogWriter {
	return &MetaLogWriter{
		MetaUtilFuncs: metaUtilFuncs,
	}
}

func (lw *MetaLogWriter) Write(p []byte) (n int, err error) {
	if lw.MetaUtilFuncs == nil {
		return 0, nil
	}

	msg := strings.TrimSuffix(string(p), "\n")

	lw.MetaUtilFuncs.LogMessage(msg)

	return len(p), nil
}

type Logger struct {
	MetaUtilFuncs *metamod.MUtilFuncs
}

func NewLogger(metaUtilFuncs *metamod.MUtilFuncs) *Logger {
	return &Logger{
		MetaUtilFuncs: metaUtilFuncs,
	}
}

func (l *Logger) Message(message string) {
	if l.MetaUtilFuncs == nil {
		return
	}

	l.MetaUtilFuncs.LogMessage(message)
}

func (l *Logger) Messagef(format string, args ...interface{}) {
	if l.MetaUtilFuncs == nil {
		return
	}

	l.MetaUtilFuncs.LogMessagef(format, args...)
}

func (l *Logger) Error(message string) {
	if l.MetaUtilFuncs == nil {
		return
	}

	l.MetaUtilFuncs.LogError(message)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	if l.MetaUtilFuncs == nil {
		return
	}

	l.MetaUtilFuncs.LogErrorf(format, args...)
}

func (l *Logger) Debug(message string) {
	if l.MetaUtilFuncs == nil {
		return
	}

	l.MetaUtilFuncs.LogDeveloper(message)
}

func (l *Logger) Debugf(message string, args ...interface{}) {
	if l.MetaUtilFuncs == nil {
		return
	}

	l.MetaUtilFuncs.LogDeveloperf(message, args...)
}
