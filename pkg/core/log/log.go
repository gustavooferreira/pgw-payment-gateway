// Package log provides an interface and a few helper functions/constants.
package log

type FieldsMap map[string]interface{}

type FieldFunc func(FieldsMap)

// Logger is the logger interface that should be used throughout the whole application.
type Logger interface {
	Debug(msg string, fields ...FieldFunc)
	Info(msg string, fields ...FieldFunc)
	Warn(msg string, fields ...FieldFunc)
	Error(msg string, fields ...FieldFunc)
}

// LogLevel defines the log level constants.
type Level uint

const (
	DEBUG Level = 10
	INFO  Level = 20
	WARN  Level = 30
	ERROR Level = 40
)

func Field(key string, value interface{}) FieldFunc {
	return func(newFields FieldsMap) {
		newFields[key] = value
	}
}

func Fields(fields map[string]interface{}) FieldFunc {
	return func(newFields FieldsMap) {
		for k, v := range fields {
			newFields[k] = v
		}
	}
}
