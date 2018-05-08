package logfmt

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

type Formatter struct{}

const (
	FieldKeyTime   = "time"
	FieldKeyLevel  = "level"
	FieldKeyMsg    = "msg"
	FieldKeyError  = "err"
	FieldKeyTrace  = "stacktrace"
	FieldKeyLogger = "logger"
)

const escapeChars = " ="

const entryFormat = "%s=%s "

const genericError = "Generic"

const logfmtTimeFormat = "2006-01-02T15:04:05.999Z"

const defaultLogger = "default"

func escapeValue(value interface{}) string {
	strValue := fmt.Sprintf("%v", value)
	if strings.ContainsAny(strValue, escapeChars) {
		return fmt.Sprintf("\"%s\"", strValue)
	}
	return strValue
}

// Format entry into logfmt format
func (f *Formatter) Format(entry *logrus.Entry) ([]byte, error) {
	buf := &bytes.Buffer{}
	if _, err := fmt.Fprintf(buf, entryFormat, FieldKeyTime, entry.Time.Format(logfmtTimeFormat)); nil != err {
		return nil, err
	}
	if _, err := fmt.Fprintf(buf, entryFormat, FieldKeyLevel, entry.Level); nil != err {
		return nil, err
	}
	if _, err := fmt.Fprintf(buf, entryFormat, FieldKeyMsg, escapeValue(entry.Message)); nil != err {
		return nil, err
	}

	var logger string

	if l, ok := entry.Data[FieldKeyLogger]; ok {
		logger = fmt.Sprintf("%s", l)
	} else {
		logger = defaultLogger
	}

	if _, err := fmt.Fprintf(buf, entryFormat, FieldKeyLogger, escapeValue(logger)); nil != err {
		return nil, err
	}

	if entry.Level <= logrus.ErrorLevel {
		var errClass string
		if err, ok := entry.Data[logrus.ErrorKey]; ok {
			errClass = fmt.Sprintf("%T", err)
		} else {
			errClass = genericError
		}

		if _, err := fmt.Fprintf(buf, entryFormat, FieldKeyError, errClass); nil != err {
			return nil, err
		}
	}

	for key, value := range entry.Data {
		if key == FieldKeyTime || key == FieldKeyLevel || key == FieldKeyMsg || key == logrus.ErrorKey || key == FieldKeyLogger {
			continue
		}

		escapedValue := escapeValue(value)

		if _, err := fmt.Fprintf(buf, entryFormat, key, escapedValue); nil != err {
			return nil, err
		}
	}
	if _, err := buf.WriteRune('\n'); nil != err {
		return nil, err
	}
	return buf.Bytes(), nil
}

type loggerHook string

func (hook loggerHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook loggerHook) Fire(entry *logrus.Entry) error {
	entry.Data[FieldKeyLogger] = hook
	return nil
}

// NewLogger creates a logfmt logger with a given name
func NewLogger(name string) *logrus.Logger {
	logger := logrus.New()
	logger.Formatter = &Formatter{}
	logger.AddHook(loggerHook(name))
	return logger
}