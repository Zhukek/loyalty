package slogger

import "go.uber.org/zap"

type Slogger struct {
	slogger *zap.SugaredLogger
}

func (s *Slogger) LogErr(args ...any) {
	s.slogger.Errorln(args)
}

func (s *Slogger) LogInfo(args ...any) {
	s.slogger.Infoln(args...)
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
