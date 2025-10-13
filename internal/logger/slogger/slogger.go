package slogger

import "go.uber.org/zap"

type Slogger struct {
	slogger *zap.SugaredLogger
}

func (s *Slogger) LogErr(target string, err error) {
	s.slogger.Errorln(
		"target", target,
		"error", err,
	)
}

func (s *Slogger) LogInfo(target string, err error) {
	s.slogger.Infoln(
		"target", target,
		"error", err,
	)
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
