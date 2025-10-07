package logger

type Logger interface {
	LogErr(err error)
	LogInfo(str string)
	Sync()
}
