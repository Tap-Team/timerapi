package timerstorage

import (
	"context"
	"fmt"

	"github.com/Tap-Team/timerapi/internal/errorutils/timererror"
	"github.com/Tap-Team/timerapi/internal/model/timermodel"
	"github.com/Tap-Team/timerapi/internal/model/timermodel/timerfields"
	"github.com/Tap-Team/timerapi/internal/sqlmodel/colorsql"
	"github.com/Tap-Team/timerapi/internal/sqlmodel/countdowntimersql"
	"github.com/Tap-Team/timerapi/internal/sqlmodel/subscribersql"
	"github.com/Tap-Team/timerapi/internal/sqlmodel/timersql"
	"github.com/Tap-Team/timerapi/internal/sqlmodel/typesql"
	"github.com/Tap-Team/timerapi/pkg/exception"
	"github.com/Tap-Team/timerapi/pkg/sqlutils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var insertTimerQuery = fmt.Sprintf(
	`
	INSERT INTO %s 
		(%s,%s,%s,%s,%s,%s,%s,%s,%s,%s)
	VALUES 
		(
			$1,
			$2,
			$3,
			$4,
			(SELECT %s FROM %s WHERE %s = $5),
			$6,
			$7,
			(SELECT %s FROM %s WHERE %s = $8),
			$9,
			$10
		)
	`,

	// insert into timers table
	timersql.Table,

	// insert variables
	timersql.ID,
	timersql.UTC,
	timersql.Creator,
	timersql.EndTime,
	timersql.TypeId,
	timersql.Name,
	timersql.Description,
	timersql.ColorId,
	timersql.WithMusic,
	timersql.Duration,

	// select type id from types
	typesql.ID,
	typesql.Table,
	typesql.Type,

	// select color id from types
	colorsql.ID,
	colorsql.Table,
	colorsql.Color,
)

func insertTimerTx(ctx context.Context, tx pgx.Tx, creator int64, timer *timermodel.CreateTimer) error {
	var err error
	_, err = tx.Exec(
		ctx,
		insertTimerQuery,
		timer.ID,
		timer.UTC,
		creator,
		timer.EndTime,
		timer.Type,
		timer.Name,
		timer.Description,
		timer.Color,
		timer.WithMusic,
		timer.DefaultDuration(),
	)
	if err != nil {
		return Error(err, exception.NewCause("insert timer into storage", "insertTimerTx", _PROVIDER))
	}
	_, err = tx.Exec(ctx, subscribeQuery, creator, timer.ID)
	if err != nil {
		return Error(err, exception.NewCause("subscribe creator", "insertTimerTx", _PROVIDER))
	}
	return nil
}

func (s *Storage) InsertDateTimer(ctx context.Context, creator int64, timer *timermodel.CreateTimer) error {
	tx, err := s.p.Pool.Begin(ctx)
	if err != nil {
		return Error(err, exception.NewCause("begin tx", "InsertTimer", _PROVIDER))
	}
	defer tx.Rollback(ctx)
	timer.Type = timerfields.DATE
	err = insertTimerTx(ctx, tx, creator, timer)
	if err != nil {
		return Error(err, exception.NewCause("insert date timer into storage", "InsertTimer", _PROVIDER))
	}
	err = tx.Commit(ctx)
	if err != nil {
		return Error(err, exception.NewCause("commit tx", "InsertTimer", _PROVIDER))
	}
	return nil
}

var deleteTimerQuery = fmt.Sprintf(
	`
	UPDATE %s SET %s = true WHERE %s = $1
	`,
	timersql.Table,
	timersql.IsDeleted,
	timersql.ID,
)

func (s *Storage) DeleteTimer(ctx context.Context, id uuid.UUID) error {
	cmd, err := s.p.Pool.Exec(ctx, deleteTimerQuery, id)
	if err != nil {
		return Error(err, exception.NewCause("delete timer query", "DeleteTimer", _PROVIDER))
	}
	if cmd.RowsAffected() == 0 {
		return timererror.ExceptionTimerNotFound()
	}
	return nil
}

// template select timer query,
// need add GROUP BY timerId, colorId, typeId and ORDER BY

var selectTimerQueryTemplate = fmt.Sprintf(
	`SELECT 
		%s,coalesce(%s, false), coalesce(%s, NULL)
	FROM %s 
	INNER JOIN %s ON %s = %s AND NOT %s
	INNER JOIN %s ON %s = %s
	LEFT JOIN %s ON %s = %s`,
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
	),
	sqlutils.Full(countdowntimersql.IsPaused),
	sqlutils.Full(countdowntimersql.PauseTime),

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

	// left join on countdowntimers for coalesce(is_paused, false) field
	countdowntimersql.Table,
	sqlutils.Full(timersql.ID),
	sqlutils.Full(countdowntimersql.TimerId),
)

func timerQueryTemplate(query string) string {
	return fmt.Sprintf(
		`
		%s

		%s

		GROUP BY %s
		`,

		selectTimerQueryTemplate,

		// added query
		query,

		// group by
		sqlutils.Full(
			countdowntimersql.TimerId,
			timersql.ID,
			colorsql.ID,
			typesql.ID,
		),
	)
}

func scanTimer(row pgx.Row, timer *timermodel.Timer) error {
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
		&timer.IsPaused,
		&timer.PauseTime,
	)
}

var timerQuery = timerQueryTemplate(fmt.Sprintf(`WHERE %s = $1`, sqlutils.Full(timersql.ID)))

func (s *Storage) Timer(ctx context.Context, timerId uuid.UUID) (*timermodel.Timer, error) {
	row := s.p.Pool.QueryRow(ctx, timerQuery, timerId)
	timer := new(timermodel.Timer)
	err := scanTimer(row, timer)
	if err != nil {
		return nil, Error(err, exception.NewCause("error while scan timer", "Timer", _PROVIDER))
	}
	return timer, nil
}

var userSubscriptionsQuery = fmt.Sprintf(`
	%s
	INNER JOIN %s ON %s = %s AND %s = $1 AND %s != $1

	GROUP BY %s

	ORDER BY %s

	LIMIT $2
	OFFSET $3
`,
	// default select timer query
	selectTimerQueryTemplate,

	subscribersql.Table,
	// inner join by timer id
	sqlutils.Full(timersql.ID),
	sqlutils.Full(subscribersql.TimerId),
	// inner join by userId = $1
	sqlutils.Full(subscribersql.UserId),
	// and user id not equal to creator
	sqlutils.Full(timersql.Creator),

	sqlutils.Full(
		countdowntimersql.TimerId,
		timersql.ID,
		colorsql.ID,
		typesql.ID,
	),
	sqlutils.Full(timersql.EndTime),
)

// return list of user subcriptions on timers
func (s *Storage) UserSubscriptions(ctx context.Context, userId int64, offset, limit int) ([]*timermodel.Timer, error) {
	rows, err := s.p.Pool.Query(ctx, userSubscriptionsQuery, userId, limit, offset)
	if err != nil {
		return nil, Error(err, exception.NewCause("user subscriptions query", "UserSubscriptions", _PROVIDER))
	}
	timers, err := sqlutils.ScanList(rows, scanTimer)
	if err != nil {
		return nil, Error(err, exception.NewCause("scan rows into timer list", "UserSubcriptions", _PROVIDER))
	}
	return timers, nil
}

var userTimersQuery = fmt.Sprintf(`
%s
INNER JOIN %s ON %s = %s AND %s = $1 OR %s = $1
GROUP BY %s
ORDER BY %s
LIMIT $2
OFFSET $3
`,
	// default select timer query
	selectTimerQueryTemplate,

	subscribersql.Table,
	// inner join by timer id
	sqlutils.Full(timersql.ID),
	sqlutils.Full(subscribersql.TimerId),

	// inner join by userId = $1
	sqlutils.Full(subscribersql.UserId),
	// where creator = $2
	sqlutils.Full(timersql.Creator),

	sqlutils.Full(
		countdowntimersql.TimerId,
		timersql.ID,
		colorsql.ID,
		typesql.ID,
	),
	sqlutils.Full(timersql.EndTime),
)

var createdTimers = timerQueryTemplate(
	fmt.Sprintf(`WHERE %s = $1`, sqlutils.Full(timersql.Creator)),
) + fmt.Sprintf("ORDER BY %s LIMIT $2 OFFSET $3", sqlutils.Full(timersql.EndTime))

func (s *Storage) UserCreatedTimers(ctx context.Context, userId int64, offset, limit int) ([]*timermodel.Timer, error) {
	rows, err := s.p.Pool.Query(ctx, createdTimers, userId, limit, offset)
	if err != nil {
		return nil, Error(err, exception.NewCause("user created timers query", "UserCreatedTimers", _PROVIDER))
	}
	timers, err := sqlutils.ScanList(rows, scanTimer)
	if err != nil {
		return nil, Error(err, exception.NewCause("scan rows into timer list", "UserCreatedTimers", _PROVIDER))
	}
	return timers, nil
}

// return list of all user timers include subcriptions
func (s *Storage) UserTimers(ctx context.Context, userId int64, offset, limit int) ([]*timermodel.Timer, error) {
	rows, err := s.p.Pool.Query(ctx, userTimersQuery, userId, limit, offset)
	if err != nil {
		return nil, Error(err, exception.NewCause("user timers query", "UserTimers", _PROVIDER))
	}
	timers, err := sqlutils.ScanList(rows, scanTimer)
	if err != nil {
		return nil, Error(err, exception.NewCause("scan rows into timer list", "UserTimers", _PROVIDER))
	}
	return timers, nil
}

var subscribeQuery = fmt.Sprintf(`
	INSERT INTO %s (%s,%s) VALUES ($1,$2)
`,
	subscribersql.Table,
	subscribersql.UserId,
	subscribersql.TimerId,
)

func (s *Storage) Subscribe(ctx context.Context, timerId uuid.UUID, userId int64) error {
	_, err := s.p.Pool.Exec(ctx, subscribeQuery, userId, timerId)
	if err != nil {
		return Error(err, exception.NewCause("insert into subcribers table", "Subcribe", _PROVIDER))
	}
	return nil
}

func (s *Storage) SubscribeAll(ctx context.Context, timerId uuid.UUID, userIds ...int64) error {
	tx, err := s.p.Pool.Begin(ctx)
	if err != nil {
		return Error(err, exception.NewCause("begin tx", "SubscribeAll", _PROVIDER))
	}
	defer tx.Rollback(ctx)

	rows := make([][]any, 0)
	for _, userId := range userIds {
		rows = append(rows, []any{timerId, userId})
	}

	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{subscribersql.Table},
		[]string{
			subscribersql.TimerId.String(),
			subscribersql.UserId.String(),
		},
		pgx.CopyFromRows(rows),
	)

	if err != nil {
		return Error(err, exception.NewCause("copy from rows", "SubscribeAll", _PROVIDER))
	}
	err = tx.Commit(ctx)

	if err != nil {
		return Error(err, exception.NewCause("commit tx", "SubscribeAll", _PROVIDER))
	}
	return nil
}

var unsubcribeQuery = fmt.Sprintf(`
	DELETE FROM %s WHERE %s = $1 AND %s = $2
`,
	subscribersql.Table,
	subscribersql.UserId,
	subscribersql.TimerId,
)

func (s *Storage) Unsubscribe(ctx context.Context, timerId uuid.UUID, userId int64) error {
	_, err := s.p.Pool.Exec(ctx, unsubcribeQuery, userId, timerId)
	if err != nil {
		return Error(err, exception.NewCause("insert into subcribers table", "Subcribe", _PROVIDER))
	}
	return nil
}
