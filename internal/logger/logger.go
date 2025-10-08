package logger

type Logger interface {
	LogErr(args ...any)
	LogInfo(args ...any)
	Sync()
}
