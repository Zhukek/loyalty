package slogger

import "go.uber.org/zap"

type Slogger struct {
	slogger *zap.SugaredLogger
}

func (s *Slogger) LogErr(err error) {
	s.slogger.Errorln(err)
}

func (s *Slogger) LogInfo(str string) {
	s.slogger.Infoln(str)
}

func (s *Slogger) Sync() {
	s.slogger.Sync()
}

func NewSlogger() (*Slogger, error) {
	logger, err := zap.NewProduction()

	if err != nil {
		return nil, err
	}

	return &Slogger{logger.Sugar()}, nil
}
