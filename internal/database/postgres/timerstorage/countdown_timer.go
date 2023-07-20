package timerstorage

import (
	"context"
	"fmt"

	"github.com/Tap-Team/timerapi/internal/errorutils/timererror"
	"github.com/Tap-Team/timerapi/internal/model/timermodel"
	"github.com/Tap-Team/timerapi/internal/model/timermodel/timerfields"
	"github.com/Tap-Team/timerapi/internal/sqlmodel/colorsql"
	"github.com/Tap-Team/timerapi/internal/sqlmodel/countdowntimersql"
	"github.com/Tap-Team/timerapi/internal/sqlmodel/timersql"
	"github.com/Tap-Team/timerapi/internal/sqlmodel/typesql"
	"github.com/Tap-Team/timerapi/pkg/amidtime"
	"github.com/Tap-Team/timerapi/pkg/exception"
	"github.com/Tap-Team/timerapi/pkg/sqlutils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var insertCountDownTimerQuery = fmt.Sprintf(
	`INSERT INTO %s (%s) VALUES($1)`,
	countdowntimersql.Table,
	countdowntimersql.TimerId,
)

func (s *Storage) InsertCountdownTimer(ctx context.Context, creator int64, timer *timermodel.CreateTimer) error {
	tx, err := s.p.Pool.Begin(ctx)
	if err != nil {
		return Error(err, exception.NewCause("begin tx", "InsertCountDownTimer", _PROVIDER))
	}
	defer tx.Rollback(ctx)
	timer.Type = timerfields.COUNTDOWN
	err = insertTimerTx(ctx, tx, creator, timer)
	if err != nil {
		return Error(err, exception.NewCause("insert timer into storage", "InsertCountDownTimer", _PROVIDER))
	}
	_, err = tx.Exec(
		ctx,
		insertCountDownTimerQuery,
		timer.ID,
	)
	if err != nil {
		return Error(err, exception.NewCause("insert countdown timer into storage", "InsertCountDownTimer", _PROVIDER))
	}
	err = tx.Commit(ctx)
	if err != nil {
		return Error(err, exception.NewCause("commit tx", "InsertCountDownTimer", _PROVIDER))
	}
	return nil
}

var updateTimerPauseTimeQuery = fmt.Sprintf(
	`UPDATE %s SET %s = $1, %s = $2 WHERE %s = $3`,
	countdowntimersql.Table,
	countdowntimersql.PauseTime,
	countdowntimersql.IsPaused,
	countdowntimersql.TimerId,
)

func (s *Storage) UpdatePauseTime(ctx context.Context, timerId uuid.UUID, pauseTime amidtime.DateTime, isPaused bool) error {
	cmd, err := s.p.Pool.Exec(ctx, updateTimerPauseTimeQuery, pauseTime, isPaused, timerId)
	if err != nil {
		return Error(err, exception.NewCause("update timer pause time failed", "UpdatePauseTime", _PROVIDER))
	}
	if cmd.RowsAffected() == 0 {
		return Error(timererror.ExceptionCountDownTimerNotFound, exception.NewCause("update timer rows = 0", "UpdatePauseTime", _PROVIDER))
	}
	return nil
}

var timerPauseQuery = fmt.Sprintf(
	`
	SELECT %s
	FROM %s 
	INNER JOIN %s ON %s = $1 AND %s = false
	`,
	sqlutils.Full(
		countdowntimersql.PauseTime,
		countdowntimersql.IsPaused,
	),
	countdowntimersql.Table,

	// inner join on timers on id = $1 and is_deleted = false
	timersql.Table,
	sqlutils.Full(timersql.ID),
	sqlutils.Full(timersql.IsDeleted),
)

func scanTimerPause(row pgx.Row, tp *timermodel.TimerPause) error {
	return row.Scan(
		&tp.PauseTime,
		&tp.IsPaused,
	)
}

func (s *Storage) TimerPause(ctx context.Context, timerId uuid.UUID) (*timermodel.TimerPause, error) {
	tp := &timermodel.TimerPause{ID: timerId}
	row := s.p.Pool.QueryRow(ctx, timerPauseQuery, timerId)
	err := scanTimerPause(row, tp)
	if err != nil {
		return nil, Error(err, exception.NewCause("scan timer pause", "TimerPause", _PROVIDER))
	}
	return tp, nil
}

var countdownTimerQuery = fmt.Sprintf(
	`SELECT 
		%s 
	FROM %s 
	INNER JOIN %s ON %s = %s AND NOT %s
	INNER JOIN %s ON %s = %s
	INNER JOIN %s ON %s = %s

	WHERE %s = $1

	GROUP BY %s
	`,

	// selectable variables
	sqlutils.Full(
		timersql.ID,
		timersql.UTC,
		timersql.Creator,
		timersql.EndTime,
		typesql.Type,
		timersql.Name,
		timersql.Description,
		colorsql.Color,
		timersql.WithMusic,
		timersql.Duration,
		countdowntimersql.PauseTime,
		countdowntimersql.IsPaused,
	),

	// from timers
	timersql.Table,

	// inner join colors
	colorsql.Table,
	sqlutils.Full(timersql.ColorId),
	sqlutils.Full(colorsql.ID),
	// AND NOT is_deleted
	sqlutils.Full(timersql.IsDeleted),

	// inner join types
	typesql.Table,
	sqlutils.Full(timersql.TypeId),
	sqlutils.Full(typesql.ID),

	// inner join countdowntimer
	countdowntimersql.Table,
	sqlutils.Full(timersql.ID),
	sqlutils.Full(countdowntimersql.TimerId),

	sqlutils.Full(timersql.ID),

	sqlutils.Full(
		timersql.ID,
		colorsql.ID,
		typesql.ID,
		countdowntimersql.TimerId,
	),
)

func scanCountdownTimer(row pgx.Row, timer *timermodel.CountdownTimer) error {
	return row.Scan(
		&timer.ID,
		&timer.UTC,
		&timer.Creator,
		&timer.EndTime,
		&timer.Type,
		&timer.Name,
		&timer.Description,
		&timer.Color,
		&timer.WithMusic,
		&timer.Duration,
		&timer.PauseTime,
		&timer.IsPaused,
	)
}

func (s *Storage) CountdownTimer(ctx context.Context, timerId uuid.UUID) (*timermodel.CountdownTimer, error) {
	row := s.p.Pool.QueryRow(ctx, countdownTimerQuery, timerId)
	timer := new(timermodel.CountdownTimer)
	err := scanCountdownTimer(row, timer)
	if err != nil {
		return nil, Error(err, exception.NewCause("error while scan timer", "CountdownTimer", _PROVIDER))
	}
	return timer, nil
}
