package log

// NullLogger defines a null logger, i.e., a logger that does nothing
type NullLogger struct{}

func (l NullLogger) Debug(msg string, fields ...FieldFunc) {}
func (l NullLogger) Info(msg string, fields ...FieldFunc)  {}
func (l NullLogger) Warn(msg string, fields ...FieldFunc)  {}
func (l NullLogger) Error(msg string, fields ...FieldFunc) {}
