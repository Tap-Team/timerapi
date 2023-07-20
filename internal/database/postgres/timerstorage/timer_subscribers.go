package timerstorage

import (
	"context"
	"fmt"

	"github.com/Tap-Team/timerapi/internal/model/timermodel"
	"github.com/Tap-Team/timerapi/internal/sqlmodel/subscribersql"
	"github.com/Tap-Team/timerapi/internal/sqlmodel/timersql"
	"github.com/Tap-Team/timerapi/pkg/exception"
	"github.com/Tap-Team/timerapi/pkg/sqlutils"
	"github.com/jackc/pgx/v5"
)

var timerSubscribersQuery = fmt.Sprintf(
	`
	SELECT %s, array_agg(%s)
	FROM %s
	INNER JOIN %s ON %s = %s AND NOT %s

	GROUP BY %s
	ORDER BY %s

	LIMIT  $1
	OFFSET $2
	`,
	sqlutils.Full(
		timersql.ID,
		timersql.EndTime,
	),
	sqlutils.Full(subscribersql.UserId),

	timersql.Table,

	subscribersql.Table,
	sqlutils.Full(timersql.ID),
	sqlutils.Full(subscribersql.TimerId),

	sqlutils.Full(timersql.IsDeleted),

	sqlutils.Full(timersql.ID),

	sqlutils.Full(timersql.EndTime),
)

func scanTimerSubscribers(row pgx.Row, timerSubscribers *timermodel.TimerSubscribers) error {
	return row.Scan(
		&timerSubscribers.ID,
		&timerSubscribers.EndTime,
		&timerSubscribers.Subscribers,
	)
}

func (s *Storage) TimerWithSubscribers(ctx context.Context, offset, limit int) ([]*timermodel.TimerSubscribers, error) {
	rows, err := s.p.Pool.Query(ctx, timerSubscribersQuery, limit, offset)
	if err != nil {
		return nil, Error(err, exception.NewCause("timer subscriber query", "TimerWithSubscrribers", _PROVIDER))
	}
	timerSubList, err := sqlutils.ScanList(rows, scanTimerSubscribers)
	if err != nil {
		return nil, Error(err, exception.NewCause("scan timer subscribers list", "TimerWithSubscribers", _PROVIDER))
	}
	return timerSubList, nil
}
