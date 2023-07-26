package timerstorage

import (
	"context"
	"errors"
	"fmt"

	"github.com/Tap-Team/timerapi/internal/errorutils/timererror"
	"github.com/Tap-Team/timerapi/internal/model/timermodel"
	"github.com/Tap-Team/timerapi/internal/sqlmodel/colorsql"
	"github.com/Tap-Team/timerapi/internal/sqlmodel/timersql"
	"github.com/Tap-Team/timerapi/pkg/amidtime"
	"github.com/Tap-Team/timerapi/pkg/exception"
	"github.com/Tap-Team/timerapi/pkg/sqlutils"
	"github.com/google/uuid"
)

var updateEndTimeQuery = fmt.Sprintf(
	`UPDATE %s SET %s = $1 WHERE %s = $2`,
	timersql.Table,
	timersql.EndTime,
	timersql.ID,
)

func (s *Storage) UpdateTime(ctx context.Context, timerId uuid.UUID, endTime amidtime.DateTime) error {
	cmd, err := s.p.Pool.Exec(ctx, updateEndTimeQuery, endTime, timerId)
	if err != nil {
		return Error(err, exception.NewCause("update timer endTime", "UpdateTime", _PROVIDER))
	}
	if cmd.RowsAffected() == 0 {
		return timererror.ExceptionTimerNotFound()
	}
	if cmd.RowsAffected() > 1 {
		return Error(errors.New("many than 1 rows was updated"), exception.NewCause("update timer end time", "UpdateTime", _PROVIDER))
	}
	return nil
}

var updateTimerQuery = fmt.Sprintf(
	`
	UPDATE %s 
	SET %s = $1, %s = $2, %s = %s, %s = $4
	FROM %s 
	WHERE %s = $5 AND %s = $3
	`,
	timersql.Table,

	// update variables
	timersql.Name,
	timersql.Description,
	timersql.ColorId,
	sqlutils.Full(colorsql.ID),
	timersql.WithMusic,

	// update from
	colorsql.Table,

	// where timer id = $5 and color = $3
	sqlutils.Full(timersql.ID),
	sqlutils.Full(colorsql.Color),
)

func (s *Storage) UpdateTimer(ctx context.Context, timerId uuid.UUID, timerSettings *timermodel.TimerSettings) error {
	cmd, err := s.p.Pool.Exec(
		ctx,
		updateTimerQuery,
		timerSettings.Name, timerSettings.Description, timerSettings.Color, timerSettings.WithMusic,
		timerId,
	)
	if err != nil {
		return Error(err, exception.NewCause("update timer endTime", "UpdateEndTime", _PROVIDER))
	}
	if cmd.RowsAffected() == 0 {
		return Error(timererror.ExceptionTimerNotFound(), exception.NewCause("update timer endTime", "UpdateEndTime", _PROVIDER))
	}
	if cmd.RowsAffected() > 1 {
		return Error(errors.New("many than 1 rows was updated"), exception.NewCause("update timer end time", "UpdateEndTime", _PROVIDER))
	}
	return nil
}
