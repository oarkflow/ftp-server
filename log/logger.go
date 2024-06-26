package log

type Logger interface {

	// Debug logging: Every details
	Debug(event string, keyvals ...interface{})

	// Info logging: Core events
	Info(event string, keyvals ...interface{})

	// Warning logging: Anything out of the ordinary but non-life threatening
	Warn(event string, keyvals ...interface{})

	// Error logging: Major issue
	Error(event string, keyvals ...interface{})

	// Panic logging: We want to crash
	Panic(event string, keyvals ...interface{})

	// Context extending interface
	With(keyvals ...interface{}) Logger
}
