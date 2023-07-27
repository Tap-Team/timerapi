package notificationhandler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/Tap-Team/timerapi/internal/database/postgres/notificationstorage"
	"github.com/Tap-Team/timerapi/internal/database/postgres/timerstorage"
	"github.com/Tap-Team/timerapi/internal/database/redis/subscriberstorage"
	"github.com/Tap-Team/timerapi/internal/domain/datastream/timereventstream"
	"github.com/Tap-Team/timerapi/internal/domain/datastream/timernotificationstream"
	"github.com/Tap-Team/timerapi/internal/domain/usecase/countdowntimerusecase"
	"github.com/Tap-Team/timerapi/internal/domain/usecase/notificationusecase"
	"github.com/Tap-Team/timerapi/internal/domain/usecase/timerusecase"
	"github.com/Tap-Team/timerapi/internal/model/notification"
	"github.com/Tap-Team/timerapi/internal/model/timermodel"
	"github.com/Tap-Team/timerapi/internal/model/timermodel/timerfields"
	"github.com/Tap-Team/timerapi/internal/timerservice"
	"github.com/Tap-Team/timerapi/internal/transport/rest/notificationhandler"
	"github.com/Tap-Team/timerapi/internal/transport/rest/timerhandler"
	"github.com/Tap-Team/timerapi/pkg/amidtime"
	"github.com/Tap-Team/timerapi/pkg/postgres"
	"github.com/Tap-Team/timerapi/pkg/rediscontainer"
	"github.com/Tap-Team/timerapi/proto/timerservicepb"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func timerPath(path string) string {
	return "/timers" + path
}

func notificationPath(path string) string {
	return "/notifications" + path
}

func createTimer(ctx context.Context, userId int64, timer *timermodel.CreateTimer) (*httptest.ResponseRecorder, error) {
	b, _ := json.Marshal(timer)
	req := httptest.NewRequest(http.MethodPost, timerPath("/create?vk_user_id="+fmt.Sprint(userId)), bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return rec, timerHandler.CreateTimer(ctx)(c)
}
func deleteTimer(ctx context.Context, userId int64, timerId uuid.UUID) (*httptest.ResponseRecorder, error) {
	req := httptest.NewRequest(http.MethodDelete, timerPath("/:id"+"?vk_user_id="+fmt.Sprint(userId)), new(bytes.Buffer))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprint(timerId))
	return rec, timerHandler.DeleteTimer(ctx)(c)
}

func subscribe(ctx context.Context, userId int64, timerId uuid.UUID) (*httptest.ResponseRecorder, error) {
	req := httptest.NewRequest(http.MethodPost, timerPath("/:id/subscribe?vk_user_id="+fmt.Sprint(userId)), new(bytes.Buffer))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(timerId.String())
	return rec, timerHandler.Subscribe(ctx)(c)
}

func userNotifications(ctx context.Context, userId int64) (*httptest.ResponseRecorder, error) {
	req := httptest.NewRequest(http.MethodGet, notificationPath("?vk_user_id="+fmt.Sprint(userId)), new(bytes.Buffer))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return rec, notificationHandler.NotificationsByUser(ctx)(c)
}

func deleteNotifications(ctx context.Context, userId int64) (*httptest.ResponseRecorder, error) {
	req := httptest.NewRequest(http.MethodDelete, notificationPath("?vk_user_id="+fmt.Sprint(userId)), new(bytes.Buffer))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return rec, notificationHandler.Delete(ctx)(c)
}

var (
	e                   *echo.Echo = echo.New()
	timerHandler        *timerhandler.Handler
	notificationHandler *notificationhandler.Handler
	notificationStorage notificationusecase.NotificationStorage
	timerUseCase        timerhandler.TimerUseCase
)

var tickerServiceUrl = "localhost:50001"

func TestMain(m *testing.M) {
	os.Setenv("TZ", "UTC")
	ctx := context.Background()
	p, termp, err := postgres.NewContainer(ctx, postgres.DEFAULT_MIGRATION_PATH)
	if err != nil {
		log.Fatalf("create postgres container failed, %s", err)
	}
	defer termp(ctx)
	r, termr, err := rediscontainer.New(ctx)
	if err != nil {
		log.Fatalf("create redis container failed, %s", err)
	}
	defer termr(ctx)
	conn, err := grpc.DialContext(ctx, tickerServiceUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("dial context failed, %s", err)
	}
	timerService := timerservice.GrpcClient(timerservicepb.NewTimerServiceClient(conn))

	ts := timerstorage.New(p)
	subst := subscriberstorage.New(r)
	nst := notificationstorage.New(p)

	notificationStorage = nst

	es := timereventstream.New()
	ns := timernotificationstream.New(
		timerService,
		ts,
		subst,
		nst,
	)
	go ns.Start(ctx)

	timerUseCase = timerusecase.New(ts, subst, timerService, es, ns)
	countdownUseCase := countdowntimerusecase.New(timerService, ts, es)
	notificationUseCase := notificationusecase.New(notificationStorage)

	timerHandler = timerhandler.New(countdownUseCase, timerUseCase)
	notificationHandler = notificationhandler.New(notificationUseCase)

	m.Run()
}

type randomTimerOption func(t *timermodel.Timer)

func randomTimer(opts ...randomTimerOption) *timermodel.Timer {
	duration := rand.Int31()
	timer := timermodel.NewTimer(
		uuid.New(),
		240,
		rand.Int63(),
		amidtime.DateTime(time.Now().Add(time.Second*time.Duration(duration))),
		amidtime.DateTime{},
		timerfields.DATE,
		"",
		"",
		timerfields.BLUE,
		rand.Intn(3)%2 == 0,
		int64(duration),
		false,
	)
	for _, opt := range opts {
		opt(timer)
	}
	return timer
}

func randomTimerList(size int, opts ...randomTimerOption) []*timermodel.Timer {
	tl := make([]*timermodel.Timer, 0, size)
	for i := 0; i < size; i++ {
		tl = append(tl, randomTimer(opts...))
	}
	return tl
}

func compareId(subs []*timermodel.Timer, notifications []*notification.NotificationDTO) (string, bool) {
	if len(subs) != len(notifications) {
		return fmt.Sprintf("wrong len, len(timer) = %d, len(notifications) = %d", len(subs), len(notifications)), false
	}
	for i := range notifications {
		if notifications[i].TimerId() != subs[i].ID {
			return fmt.Sprintf("uuid not equal, index %d, notification timer uuid %s, timer uuid %s", i, notifications[i].TimerId(), subs[i].ID), false
		}
	}
	return "", true
}

func TestNotification(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	userIds := []int64{
		rand.Int63(),
		rand.Int63(),
		rand.Int63(),
	}

	deleteNotification(t, ctx, userIds)
	expiredNotification(t, ctx, userIds)
}

type UserWithTimers struct {
	UserId int64
	Timers []*timermodel.Timer
}

func deleteNotification(t *testing.T, ctx context.Context, userIds []int64) {
	tam := 100
	userTimers := make([]*UserWithTimers, 0)

	var wg sync.WaitGroup
	wg.Add(len(userIds))
	// for every user
	for _, userId := range userIds {
		timers := randomTimerList(tam)
		userTimers = append(userTimers, &UserWithTimers{
			UserId: userId,
			Timers: timers,
		})
		go func(userId int64) {
			defer wg.Done()
			// for create notification we need create timer and delete
			// for receive subscribe we need subscribe on timer
			for _, timer := range timers {
				require.False(t, timer.Creator == userId, "timer creator equal userid ")
				_, err := createTimer(ctx, timer.Creator, timer.CreateTimer())
				require.NoError(t, err, "failed to create timer")
				_, err = subscribe(ctx, userId, timer.ID)
				require.NoError(t, err, "subscribe failed")
				_, err = deleteTimer(ctx, timer.Creator, timer.ID)
				require.NoError(t, err, "delete timer failed")
			}
		}(userId)
	}
	wg.Wait()

	time.Sleep(time.Second)

	compareNotification(t, ctx, userTimers, notification.Delete)
}

func expiredNotification(t *testing.T, ctx context.Context, userIds []int64) {
	tam := 100
	userTimers := make([]*UserWithTimers, 0)

	var wg sync.WaitGroup
	wg.Add(len(userIds))

	duration := int64(3)
	endTime := amidtime.DateTime(time.Now().Add(time.Second * time.Duration(duration)))
	endTimeOption := func(t *timermodel.Timer) { t.EndTime = endTime; t.Duration = duration }
	// for every user
	for _, userId := range userIds {
		userId := userId
		timers := randomTimerList(tam, endTimeOption)
		userTimers = append(userTimers, &UserWithTimers{
			UserId: userId,
			Timers: timers,
		})
		go func(userId int64) {
			defer wg.Done()
			// for create notification we need create timer and delete
			// for receive subscribe we need subsccribe on timer
			for _, timer := range timers {
				_, err := createTimer(ctx, timer.Creator, timer.CreateTimer())
				require.NoError(t, err, "failed to create timer")
				_, err = subscribe(ctx, userId, timer.ID)
				require.NoError(t, err, "subscribe failed")
			}
		}(userId)
	}
	wg.Wait()

	time.Sleep(time.Second * (time.Duration(duration) + 1))

	compareNotification(t, ctx, userTimers, notification.Expired)
}

func compareNotification(t *testing.T, ctx context.Context, userTimers []*UserWithTimers, nType notification.NotificationType) {
	// check notifications is created
	for _, ut := range userTimers {
		notifications := make([]*notification.NotificationDTO, 0, len(ut.Timers))
		var rec *httptest.ResponseRecorder
		var err error
		// get user notifications, compare timerid from notifications and user timers
		rec, err = userNotifications(ctx, ut.UserId)
		require.NoError(t, err, "failed to get user notifications")
		require.Equal(t, http.StatusOK, rec.Result().StatusCode, "wrong status code")

		err = json.Unmarshal(rec.Body.Bytes(), &notifications)
		require.NoError(t, err, "failed to unmarshal body")

		sort.Slice(notifications, func(i, j int) bool { return notifications[i].TimerId().String() > notifications[j].TimerId().String() })
		sort.Slice(ut.Timers, func(i, j int) bool { return ut.Timers[i].ID.String() > ut.Timers[j].ID.String() })
		// compare ids

		message, ok := compareId(ut.Timers, notifications)
		require.True(t, ok, message)
		// compare notifications type
		for _, n := range notifications {
			require.Equal(t, nType, n.Type(), "wrong notification type")
		}
		// delete notifications
		rec, err = deleteNotifications(ctx, ut.UserId)
		require.NoError(t, err, "failed delete user notifications")
		require.Equal(t, http.StatusNoContent, rec.Result().StatusCode, "wrong status code")

		// make sure we delete notifications
		rec, err = userNotifications(ctx, ut.UserId)
		require.NoError(t, err, "failed to get user notifications")
		require.Equal(t, http.StatusOK, rec.Result().StatusCode, "wrong status code")

		err = json.Unmarshal(rec.Body.Bytes(), &notifications)
		require.NoError(t, err, "failed to unmarshal body")

		require.Equal(t, 0, len(notifications), "non zero response of deleted notifications")
	}
}
