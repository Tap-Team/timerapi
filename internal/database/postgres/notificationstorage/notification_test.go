package notificationstorage_test

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/Tap-Team/timerapi/internal/model/notification"
	"github.com/Tap-Team/timerapi/internal/model/timermodel"
	"github.com/Tap-Team/timerapi/internal/model/timermodel/timerfields"
	"github.com/Tap-Team/timerapi/pkg/amidstr"
	"github.com/Tap-Team/timerapi/pkg/amidtime"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type randomTimerOption func(t *timermodel.Timer)

func randomTimer(opts ...randomTimerOption) *timermodel.Timer {
	timer := timermodel.NewTimer(
		uuid.New(),
		240,
		rand.Int63(),
		amidtime.DateTime(time.Now().Add(time.Hour*time.Duration(rand.Intn(100)))),
		amidtime.DateTime{},
		timerfields.DATE,
		timerfields.Name(amidstr.MakeString(timerfields.NameMaxSize)),
		timerfields.Description(amidstr.MakeString(timerfields.DescriptionMaxSize)),
		timerfields.BLUE,
		rand.Intn(3)%2 == 0,
		0,
		false,
	)
	timer.Duration = timer.EndTime.Unix() - time.Now().Unix()
	for _, opt := range opts {
		opt(timer)
	}
	return timer
}

func TestNotificationCrud(t *testing.T) {
	var err error
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	userId := rand.Int63()
	timer := randomTimer()
	err = testTimerStorage.InsertDateTimer(ctx, timer.Creator, timer.CreateTimer())
	require.NoError(t, err, "insert date timer err")
	err = testTimerStorage.DeleteTimer(ctx, timer.ID)
	require.NoError(t, err, "delete date timer err")

	err = testNotificationStorage.InsertNotification(ctx, userId, notification.NewDelete(*timer))
	require.NoError(t, err, "delete date timer err")

	notification, err := testNotificationStorage.Notification(ctx, userId, timer.ID)
	require.NoError(t, err, "get user notification err")

	userNotifications, err := testNotificationStorage.UserNotifications(ctx, userId)
	require.NoError(t, err, "get user notificationss err")

	require.Equal(t, 1, len(userNotifications), "user notification list wrong len")

	field, ok := notification.NTimer.Is(timer)
	if !ok {
		t.Fatalf("notification wrong timer data, field %s not equal", field)
	}
	field, ok = notification.NTimer.Is(&userNotifications[0].NTimer)
	if !ok {
		t.Fatalf("notification wrong timer data, field %s not equal", field)
	}

	err = testNotificationStorage.DeleteUserNotifications(ctx, userId)
	require.NoError(t, err, "delete user notifications")

	userNotifications, err = testNotificationStorage.UserNotifications(ctx, userId)
	require.NoError(t, err, "get user notifications after delete err")
	require.Equal(t, 0, len(userNotifications), "len of user notifications should equal 0")
}
