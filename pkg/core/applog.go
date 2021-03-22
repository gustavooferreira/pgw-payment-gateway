package core

import (
	"github.com/gustavooferreira/pgw-payment-gateway-service/pkg/core/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// AppLogger is the application logger.
type AppLogger struct {
	Level          log.Level
	zapLogger      *zap.Logger
	zapSugarLogger *zap.SugaredLogger
}

// NewAppLogger returns a new logger.
func NewAppLogger(ws zapcore.WriteSyncer, logLevel log.Level) *AppLogger {
	logger := AppLogger{}
	logger.setupLogger(ws, logLevel)
	return &logger
}

// Debug logs a debug message.
func (l AppLogger) Debug(msg string, fields ...log.FieldFunc) {
	l.logGeneric(log.DEBUG, msg, fields...)
}

// Info logs an info message.
func (l AppLogger) Info(msg string, fields ...log.FieldFunc) {
	l.logGeneric(log.INFO, msg, fields...)
}

// Warn logs a warning message.
func (l AppLogger) Warn(msg string, fields ...log.FieldFunc) {
	l.logGeneric(log.WARN, msg, fields...)
}

// Error logs an error message.
func (l AppLogger) Error(msg string, fields ...log.FieldFunc) {
	l.logGeneric(log.ERROR, msg, fields...)
}

// logGeneric logs a generic message.
func (l AppLogger) logGeneric(level log.Level, msg string, fields ...log.FieldFunc) {
	if l.zapLogger == nil {
		return
	}

	if l.Level <= level {
		if len(fields) != 0 {
			newFields := make(map[string]interface{})
			for _, f := range fields {
				f(newFields)
			}
			if len(newFields) != 0 {
				// Log with appropriate level
				if level == log.DEBUG {
					l.zapSugarLogger.Debugw(msg, "extra", newFields)
				} else if level == log.INFO {
					l.zapSugarLogger.Infow(msg, "extra", newFields)
				} else if level == log.WARN {
					l.zapSugarLogger.Warnw(msg, "extra", newFields)
				} else if level == log.ERROR {
					l.zapSugarLogger.Errorw(msg, "extra", newFields)
				}
				return
			}
		}

		// Log with appropriate level
		if level == log.DEBUG {
			l.zapSugarLogger.Debug(msg)
		} else if level == log.INFO {
			l.zapSugarLogger.Info(msg)
		} else if level == log.WARN {
			l.zapSugarLogger.Warn(msg)
		} else if level == log.ERROR {
			l.zapSugarLogger.Error(msg)
		}
	}
}

// setupLogger sets up Logger with all the relevant configuration params.
func (l *AppLogger) setupLogger(ws zapcore.WriteSyncer, logLevel log.Level) {
	atom := zap.NewAtomicLevel()

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	logger := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.Lock(ws),
			atom),
		zap.AddCaller(),
		zap.AddCallerSkip(2))

	l.zapLogger = logger
	l.zapSugarLogger = logger.Sugar()
	l.setLogLevel(atom, logLevel)
}

// setLogLevel sets the log level.
func (l *AppLogger) setLogLevel(atom zap.AtomicLevel, logLevel log.Level) {
	if logLevel == log.DEBUG {
		atom.SetLevel(zap.DebugLevel)
		l.Level = logLevel
	} else if logLevel == log.INFO {
		atom.SetLevel(zap.InfoLevel)
		l.Level = logLevel
	} else if logLevel == log.WARN {
		atom.SetLevel(zap.WarnLevel)
		l.Level = logLevel
	} else if logLevel == log.ERROR {
		atom.SetLevel(zap.ErrorLevel)
		l.Level = logLevel
	} else {
		atom.SetLevel(zap.InfoLevel)
		l.Level = log.INFO
		l.Warn("log level unrecognised. Setting log level to Info.", nil)
	}
}

// Sync syncs the logger, i.e., flushes any data in the buffer.
func (l AppLogger) Sync() {
	l.zapLogger.Sync()
}
