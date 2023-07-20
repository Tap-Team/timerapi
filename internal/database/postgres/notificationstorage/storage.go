package notificationstorage

import (
	"errors"

	"github.com/Tap-Team/timerapi/internal/errorutils/notificationerror"
	"github.com/Tap-Team/timerapi/pkg/exception"
	"github.com/Tap-Team/timerapi/pkg/postgres"
	"github.com/jackc/pgx/v5"
)

const _PROVIDER = "internal/database/postgres/notificationstorage"

type Storage struct {
	p *postgres.Postgres
}

func New(p *postgres.Postgres) *Storage {
	return &Storage{p: p}
}

func Error(err error, cause exception.Cause) error {
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return exception.Wrap(notificationerror.ExceptionNotificationNotFound, cause)
	default:
		return exception.Wrap(err, cause)
	}
}
