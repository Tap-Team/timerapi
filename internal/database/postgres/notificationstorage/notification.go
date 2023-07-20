package notificationstorage

import (
	"context"
	"fmt"

	"github.com/Tap-Team/timerapi/internal/model/notification"
	"github.com/Tap-Team/timerapi/internal/sqlmodel/colorsql"
	"github.com/Tap-Team/timerapi/internal/sqlmodel/notificationsql"
	"github.com/Tap-Team/timerapi/internal/sqlmodel/notificationtypesql"
	"github.com/Tap-Team/timerapi/internal/sqlmodel/timersql"
	"github.com/Tap-Team/timerapi/internal/sqlmodel/typesql"
	"github.com/Tap-Team/timerapi/pkg/exception"
	"github.com/Tap-Team/timerapi/pkg/sqlutils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var insertNotificationQuery = fmt.Sprintf(`
	INSERT INTO %s (
		%s, %s, %s
	) VALUES (
		$1,
		$2,
		(SELECT %s FROM %s WHERE %s = $3)
	)
	`,
	notificationsql.Table,

	notificationsql.UserId,
	notificationsql.TimerId,
	notificationsql.NotificationTypeId,

	// select id from notification type where type = $3
	notificationtypesql.ID,
	notificationtypesql.Table,
	notificationtypesql.Type,
)

func (s *Storage) InsertNotification(ctx context.Context, userId int64, notification notification.Notification) error {
	_, err := s.p.Pool.Exec(ctx, insertNotificationQuery, userId, notification.TimerId(), notification.Type())
	if err != nil {
		return Error(err, exception.NewCause("insert notification", "InsertNotification", _PROVIDER))
	}
	return nil
}

func userNotificationTemplate(query string) string {
	return fmt.Sprintf(`
	SELECT %s 
	FROM %s 
	INNER JOIN %s ON %s = %s
	INNER JOIN %s ON %s = %s
	INNER JOIN %s ON %s = %s
	INNER JOIN %s ON %s = %s

	%s

	GROUP BY %s

	ORDER BY %s
`,
		sqlutils.Full(
			notificationtypesql.Type,
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

		notificationsql.Table,

		// inner join on timers
		timersql.Table,
		sqlutils.Full(notificationsql.TimerId),
		sqlutils.Full(timersql.ID),
		// inner join colors
		colorsql.Table,
		sqlutils.Full(timersql.ColorId),
		sqlutils.Full(colorsql.ID),
		// inner join timer_types
		typesql.Table,
		sqlutils.Full(timersql.TypeId),
		sqlutils.Full(typesql.ID),
		// inner join notification types
		notificationtypesql.Table,
		sqlutils.Full(notificationsql.NotificationTypeId),
		sqlutils.Full(notificationtypesql.ID),

		query,

		// group by
		sqlutils.Full(
			timersql.ID,
			colorsql.ID,
			typesql.ID,
			notificationtypesql.ID,
		),

		sqlutils.Full(timersql.EndTime),
	)
}

var userNotificationsQuery = userNotificationTemplate(
	fmt.Sprintf(`WHERE %s = $1`, sqlutils.Full(notificationsql.UserId)),
)

func scanNotification(row pgx.Row, notification *notification.NotificationDTO) error {
	return row.Scan(
		&notification.Ntype,
		&notification.NTimer.ID,
		&notification.NTimer.UTC,
		&notification.NTimer.Creator,
		&notification.NTimer.EndTime,
		&notification.NTimer.Type,
		&notification.NTimer.Name,
		&notification.NTimer.Description,
		&notification.NTimer.Color,
		&notification.NTimer.WithMusic,
		&notification.NTimer.Duration,
	)
}

func (s *Storage) UserNotifications(ctx context.Context, userId int64) ([]*notification.NotificationDTO, error) {
	rows, err := s.p.Pool.Query(ctx, userNotificationsQuery, userId)
	if err != nil {
		return nil, Error(err, exception.NewCause("user notification query", "UserNotifications", _PROVIDER))
	}
	notifications, err := sqlutils.ScanList(rows, scanNotification)
	if err != nil {
		return nil, Error(err, exception.NewCause("scan notification list from rows", "UserNotifications", _PROVIDER))
	}
	return notifications, nil
}

var notificationQuery = userNotificationTemplate(
	fmt.Sprintf(
		`WHERE %s = $1 AND %s = $2`,
		sqlutils.Full(notificationsql.UserId),
		sqlutils.Full(notificationsql.TimerId),
	),
)

func (s *Storage) Notification(ctx context.Context, userId int64, timerId uuid.UUID) (*notification.NotificationDTO, error) {
	ntion := new(notification.NotificationDTO)
	err := scanNotification(s.p.Pool.QueryRow(ctx, notificationQuery, userId, timerId), ntion)
	if err != nil {
		return nil, Error(err, exception.NewCause("scan notification", "Notification", _PROVIDER))
	}
	return ntion, nil
}

var deleteUserNotificationQuery = fmt.Sprintf(`
	DELETE FROM %s
	WHERE %s = $1
`,
	notificationsql.Table,
	notificationsql.UserId,
)

func (s *Storage) DeleteUserNotifications(ctx context.Context, userId int64) error {
	_, err := s.p.Pool.Exec(ctx, deleteUserNotificationQuery, userId)
	if err != nil {
		return Error(err, exception.NewCause("delete query", "DeleteUserNotifications", _PROVIDER))
	}
	return nil
}
