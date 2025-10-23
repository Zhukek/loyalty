package pgerr

import (
	"errors"

	"github.com/Zhukek/loyalty/internal/errs"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func ClassifyUserErr(err error) error {
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			return errs.ErrUsernameTaken
		case pgerrcode.CheckViolation:
			return errs.ErrLowBalance
		}
	}
	return err
}
