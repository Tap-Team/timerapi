package timerstorage

import (
	"errors"

	"github.com/Tap-Team/timerapi/internal/errorutils/timererror"
	"github.com/Tap-Team/timerapi/internal/sqlmodel/countdowntimersql"
	"github.com/Tap-Team/timerapi/internal/sqlmodel/subscribersql"
	"github.com/Tap-Team/timerapi/internal/sqlmodel/timersql"
	"github.com/Tap-Team/timerapi/pkg/exception"
	"github.com/Tap-Team/timerapi/pkg/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const _PROVIDER = "internal/database/postgres/timerstorage"

type Storage struct {
	p *postgres.Postgres
}

func New(p *postgres.Postgres) *Storage {
	return &Storage{p: p}
}

func Error(err error, cause exception.Cause) error {
	pgerr := new(pgconn.PgError)
	if errors.As(err, &pgerr) {
		switch pgerr.ConstraintName {
		case countdowntimersql.FK_Timers:
			return exception.Wrap(timererror.ExceptionTimerNotFound(), cause)
		case subscribersql.FK_Timers:
			return exception.Wrap(timererror.ExceptionTimerNotFound(), cause)
		case subscribersql.PrimaryKey:
			return exception.Wrap(timererror.ExceptionUserAlreadySubscriber(), cause)
		case timersql.PrimaryKey:
			return exception.Wrap(timererror.ExceptionTimerExists(), cause)
		}
	}

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return exception.Wrap(timererror.ExceptionTimerNotFound(), cause)
	default:
		return exception.Wrap(err, cause)
	}
}
