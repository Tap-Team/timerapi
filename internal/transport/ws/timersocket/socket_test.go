package timersocket_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/Tap-Team/timerapi/internal/database/postgres/notificationstorage"
	"github.com/Tap-Team/timerapi/internal/database/postgres/timerstorage"
	"github.com/Tap-Team/timerapi/internal/database/redis/subscriberstorage"
	"github.com/Tap-Team/timerapi/internal/domain/datastream/timereventstream"
	"github.com/Tap-Team/timerapi/internal/domain/datastream/timernotificationstream"
	"github.com/Tap-Team/timerapi/internal/domain/usecase/countdowntimerusecase"
	"github.com/Tap-Team/timerapi/internal/domain/usecase/notificationusecase"
	"github.com/Tap-Team/timerapi/internal/domain/usecase/timerusecase"
	"github.com/Tap-Team/timerapi/internal/model/notification"
	"github.com/Tap-Team/timerapi/internal/model/timerevent"
	"github.com/Tap-Team/timerapi/internal/model/timermodel"
	"github.com/Tap-Team/timerapi/internal/timerservice"
	"github.com/Tap-Team/timerapi/internal/transport/rest/timerhandler"
	"github.com/Tap-Team/timerapi/internal/transport/ws/timersocket"
	"github.com/Tap-Team/timerapi/pkg/postgres"
	"github.com/Tap-Team/timerapi/pkg/rediscontainer"
	"github.com/Tap-Team/timerapi/proto/timerservicepb"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func basePath(path string) string {
	return "/timers" + path
}

func createTimer(ctx context.Context, userId int64, timer *timermodel.CreateTimer) (*httptest.ResponseRecorder, error) {
	b, _ := json.Marshal(timer)
	req := httptest.NewRequest(http.MethodPost, basePath("/create?vk_user_id="+fmt.Sprint(userId)), bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return rec, handler.CreateTimer(ctx)(c)
}
func deleteTimer(ctx context.Context, userId int64, timerId uuid.UUID) (*httptest.ResponseRecorder, error) {
	req := httptest.NewRequest(http.MethodDelete, basePath("/:id"+"?vk_user_id="+fmt.Sprint(userId)), new(bytes.Buffer))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprint(timerId))
	return rec, handler.DeleteTimer(ctx)(c)
}

func clearTimers(t *testing.T, ctx context.Context, timers ...*timermodel.Timer) {
	for _, timer := range timers {
		_, err := deleteTimer(ctx, timer.Creator, timer.ID)
		require.NoError(t, err, "failed delete timer")
	}
}

const tickerServiceUrl = "localhost:50001"

type NotificationStorage interface {
	timernotificationstream.NotificationStorage
	notificationusecase.NotificationStorage
}

var e *echo.Echo = echo.New()

var (
	notificationStorage NotificationStorage

	timerStorage      timerusecase.TimerStorage
	subscriberStorage timerusecase.SubscriberCacheStorage

	handler *timerhandler.Handler

	server *httptest.Server
)

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
	timerStorage = ts
	subst := subscriberstorage.New(r)
	subscriberStorage = subst
	notificationStorage = notificationstorage.New(p)

	es := timereventstream.New()
	ns := timernotificationstream.New(
		timerService,
		ts,
		subst,
		notificationStorage,
	)
	go ns.Start(ctx)

	timerUseCase := timerusecase.New(ts, subst, timerService, es, ns)
	countdownUseCase := countdowntimerusecase.New(timerService, ts, es)

	handler = timerhandler.New(countdownUseCase, timerUseCase)
	timersocket.Init(e.Group(""), es, ns)

	server = httptest.NewServer(e)
	defer server.Close()

	m.Run()
}

type WsConn struct {
	userId int64
	ws     *websocket.Conn
	es     chan timerevent.TimerEvent
	ns     chan notification.Notification
}

func NewConn(t *testing.T, s *httptest.Server, userId int64) *WsConn {
	u := "ws" + strings.TrimPrefix(s.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(u+"/ws/timer?vk_user_id="+fmt.Sprint(userId), nil)
	require.NoError(t, err, "failed to start websocket")
	return &WsConn{
		userId: userId,
		ws:     ws,
		es:     make(chan timerevent.TimerEvent),
		ns:     make(chan notification.Notification),
	}
}

func (ws *WsConn) UserId() int64 {
	return ws.userId
}

func (ws *WsConn) Listen(t *testing.T, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
		default:
			_, b, err := ws.ws.ReadMessage()
			if err != nil {
				t.Logf("\nfailed to receive message from websocket connection, %s", err)
			}
			ws.matchMessage(t, b)
		}
	}
}

func (ws *WsConn) matchMessage(t *testing.T, data []byte) {
	tp := struct {
		Type string `json:"type"`
	}{}

	err := json.Unmarshal(data, &tp)
	if err != nil {
		log.Printf("\nfailed to get message type from websocket connection, %s", err)
	}

	switch tp.Type {
	case string(notification.Delete):
		n := new(notification.NotificationDTO)
		err = json.Unmarshal(data, n)
		if err != nil {
			t.Logf("\nfailed to unmarshal delete notification")
		}
		ws.ns <- n
	case string(notification.Expired):
		n := new(notification.NotificationDTO)
		err = json.Unmarshal(data, n)
		if err != nil {
			t.Logf("\nfailed to unmarshal expired notification")
		}
		ws.ns <- n
	case string(timerevent.Reset):
		er := new(timerevent.ResetEvent)
		err = json.Unmarshal(data, er)
		if err != nil {
			t.Logf("\nfailed to unmarshal reset event")
		}
		ws.es <- er
	case string(timerevent.Stop):
		est := new(timerevent.StopEvent)
		err = json.Unmarshal(data, est)
		if err != nil {
			t.Logf("\nfailed to unmarshal stop event")
		}
		ws.es <- est
	case string(timerevent.Start):
		est := new(timerevent.StartEvent)
		err = json.Unmarshal(data, est)
		if err != nil {
			t.Logf("\nfailed to unmarshal start event")
		}
		ws.es <- est
	case string(timerevent.Update):
		eu := new(timerevent.UpdateEvent)
		err = json.Unmarshal(data, eu)
		if err != nil {
			t.Logf("\nfailed to unmarshal update event")
		}
		ws.es <- eu
	}
}

func (ws *WsConn) Subscribe(t *testing.T, ctx context.Context, timers ...uuid.UUID) {
	err := ws.ws.WriteJSON(timerevent.NewSubscribe(timers...))
	require.NoError(t, err, "failed to subscribe by %d user", ws.userId)
}

func (ws *WsConn) Unsubscribe(t *testing.T, ctx context.Context, timers ...uuid.UUID) {
	err := ws.ws.WriteJSON(timerevent.NewUnsubscribe(timers...))
	require.NoError(t, err, "failed to unsubscribe by %d user", ws.userId)
}

func (ws *WsConn) EventStream() <-chan timerevent.TimerEvent {
	return ws.es
}

func (ws *WsConn) NotificationStream() <-chan notification.Notification {
	return ws.ns
}
