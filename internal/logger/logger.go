package logger

type Logger interface {
	LogErr(target string, err error)
	LogInfo(target string, err error)
	Sync()
}
