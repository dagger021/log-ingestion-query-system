package domain

type LogLevel string

const (
	Debug    LogLevel = "debug"
	Info     LogLevel = "info"
	Warn     LogLevel = "warn"
	Error    LogLevel = "error"
	Critical LogLevel = "critical"
)

var logLevelMap = map[string]LogLevel{
	string(Debug):    Debug,
	string(Info):     Info,
	string(Warn):     Warn,
	string(Error):    Error,
	string(Critical): Critical,
}

func ToLogLevel(level string) (LogLevel, bool) {
	logLevel, ok := logLevelMap[level]
	return logLevel, ok
}
