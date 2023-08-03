package timersocket_test

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/Tap-Team/timerapi/internal/errorutils/timererror"
	"github.com/Tap-Team/timerapi/internal/model/notification"
	"github.com/Tap-Team/timerapi/internal/model/timermodel"
	"github.com/Tap-Team/timerapi/internal/model/timermodel/timerfields"
	"github.com/Tap-Team/timerapi/pkg/amidtime"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func subscribe(ctx context.Context, userId int64, timerId uuid.UUID) (*httptest.ResponseRecorder, error) {
	req := httptest.NewRequest(http.MethodPost, basePath("/:id/subscribe?vk_user_id="+fmt.Sprint(userId)), new(bytes.Buffer))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprint(timerId))
	return rec, handler.Subscribe(ctx)(c)
}

func TestNotificationStream(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	users := []int64{
		rand.Int63(),
		rand.Int63(),
		rand.Int63(),
	}
	conns := make([]*WsConn, 0)
	for _, userId := range users {
		conn := NewConn(t, ctx, server, userId)
		conns = append(conns, conn)
		go conn.Listen(t, ctx)
	}
	notificationDelete(t, ctx, conns)
	notificationExpired(t, ctx, conns)

	for _, userId := range users {
		nts, err := notificationStorage.UserNotifications(ctx, userId)
		require.NoError(t, err, "failed get user notifications")
		require.Equal(t, 0, len(nts), "handler insert notification while listener has been online")
	}
	for _, conn := range conns {
		err := conn.Close()
		require.NoError(t, err, "failed to close conn")
	}
}

// test notification right type, right timer data, check streamHandler remove subscribers
func notificationDelete(t *testing.T, ctx context.Context, conns []*WsConn) {
	timersPerConnection := 10
	connTimers := make([]struct {
		conn   *WsConn
		timers []*timermodel.Timer
	}, 0)
	for _, conn := range conns {
		// create random timer with connection creator
		timer := randomTimer(func(t *timermodel.Timer) { t.Creator = conn.UserId() })
		_, err := createTimer(ctx, conn.UserId(), timer.CreateTimer())
		require.NoError(t, err, "create timer failed")
		_, err = deleteTimer(ctx, conn.UserId(), timer.ID)
		require.NoError(t, err, "delete timer failed")
		timerList := randomTimerList(timersPerConnection-1, func(t *timermodel.Timer) {
			t.Type = timerfields.DATE
			if rand.Int63()%2 == 0 {
				t.Type = timerfields.COUNTDOWN
			}
		})
		// in loop create, subscribe and delete random timer
		// append this timer in connection timers
		// list of connection timers
		// if we delete own timer, notification handler don`t send delete notification
		tms := make([]*timermodel.Timer, 0, timersPerConnection)
		for _, tm := range timerList {
			_, err := createTimer(ctx, tm.Creator, tm.CreateTimer())
			require.NoError(t, err, "create timer failed")
			_, err = subscribe(ctx, conn.UserId(), tm.ID)
			require.NoError(t, err, "subscribe on timer failed")
			_, err = deleteTimer(ctx, tm.Creator, tm.ID)
			require.NoError(t, err, "delete timer failed")
			tms = append(tms, tm)
		}
		// append timers with conn
		connTimers = append(connTimers, struct {
			conn   *WsConn
			timers []*timermodel.Timer
		}{conn: conn, timers: tms})
	}

	// check notification in loop
	wg := new(sync.WaitGroup)
	wg.Add(len(conns))
	for _, conn := range connTimers {
		conn := conn
		go func() {
			defer wg.Done()
			ids := make([]uuid.UUID, 0)
		Loop:
			for {
				select {
				case <-ctx.Done():
					break Loop
				case <-time.After(time.Second):
					break Loop
				case n, ok := <-conn.conn.NotificationStream():
					if !ok {
						break Loop
					}
					require.Equal(t, notification.Delete, n.Type(), "wrong notification type")
					ids = append(ids, n.TimerId())
				}
			}
			sort.Slice(ids, func(i, j int) bool {
				return ids[i].String() > ids[j].String()
			})
			sort.Slice(conn.timers, func(i, j int) bool {
				return conn.timers[i].ID.String() > conn.timers[j].ID.String()
			})
			message, ok := compareIds(ids, conn.timers)
			require.True(t, ok, message)

			// check stream handler delete timer subscribers from cache
			for _, timer := range conn.timers {
				_, err := subscriberStorage.TimerSubscribers(ctx, timer.ID)
				require.ErrorIs(t, err, timererror.ExceptionTimerSubscribersNotFound(), "subscribers not deleted")
			}
		}()
	}
	wg.Wait()
}

func notificationExpired(t *testing.T, ctx context.Context, conns []*WsConn) {
	timersPerConnection := 10
	connTimers := make([]struct {
		conn   *WsConn
		timers []*timermodel.Timer
	}, 0)
	addedDuration := time.Second * 3
	endTime := amidtime.DateTime(time.Now().Add(addedDuration))
	endTimeOption := func(t *timermodel.Timer) { t.EndTime = endTime }

	for _, conn := range conns {
		timer := randomTimer(endTimeOption)
		_, err := createTimer(ctx, conn.UserId(), timer.CreateTimer())
		require.NoError(t, err, "failed to create timer with small end time")

		tms := make([]*timermodel.Timer, 0, timersPerConnection)
		tms = append(tms, timer)

		timerList := randomTimerList(timersPerConnection-1,
			endTimeOption,
			func(t *timermodel.Timer) {
				t.Type = timerfields.DATE
				if rand.Int63()%2 == 0 {
					t.Type = timerfields.COUNTDOWN
				}
			},
		)

		// create random timer and subscribe on expired
		for _, tm := range timerList {
			_, err := createTimer(ctx, tm.Creator, tm.CreateTimer())
			require.NoError(t, err, "failed to create timer with small end time")
			_, err = subscribe(ctx, conn.UserId(), tm.ID)
			require.NoError(t, err, "failed to subscribe on timer")
			tms = append(tms, tm)
		}

		// append new timers connection
		connTimers = append(connTimers, struct {
			conn   *WsConn
			timers []*timermodel.Timer
		}{
			conn:   conn,
			timers: tms,
		})
	}

	time.Sleep(addedDuration)

	wg := new(sync.WaitGroup)
	wg.Add(len(connTimers))
	for _, conn := range connTimers {
		conn := conn
		go func() {
			defer wg.Done()
			ids := make([]uuid.UUID, 0, timersPerConnection)
		Loop:
			for {
				select {
				case <-ctx.Done():
					break Loop
				case <-time.After(time.Second):
					break Loop
				case n, ok := <-conn.conn.NotificationStream():
					if !ok {
						break Loop
					}
					require.Equal(t, notification.Expired, n.Type())

					ids = append(ids, n.TimerId())
				}
			}
			sort.Slice(ids, func(i, j int) bool {
				return ids[i].String() > ids[j].String()
			})
			sort.Slice(conn.timers, func(i, j int) bool {
				return conn.timers[i].ID.String() > conn.timers[j].ID.String()
			})
			// compare timer ids
			message, ok := compareIds(ids, conn.timers)
			require.True(t, ok, message)

			wg := new(sync.WaitGroup)
			wg.Add(len(ids))
			for _, tm := range conn.timers {
				tm := tm
				go func() {
					defer wg.Done()
					// timer from storage
					timer, err := timerStorage.Timer(ctx, tm.ID)

					switch tm.Type {
					// if date timer expired, it should be deleted with subscribers
					case timerfields.DATE:
						require.ErrorIs(t, err, timererror.ExceptionTimerNotFound(), "expired timer not deleted")
						_, err := subscriberStorage.TimerSubscribers(ctx, tm.ID)
						require.ErrorIs(t, err, timererror.ExceptionTimerSubscribersNotFound(), "expired timer subscribers not deleted")
					// if countdown timer expired, we only change pause time
					case timerfields.COUNTDOWN:
						require.Equal(t, timer.PauseTime.Unix(), timer.EndTime.Unix()-timer.Duration, "countdown timer wrong time")
						subscribers, err := subscriberStorage.TimerSubscribers(ctx, timer.ID)
						require.NoError(t, err, "countdown timer subscribers has been deleted")
						_, ok := subscribers[conn.conn.UserId()]
						require.True(t, ok, "wrong timer subscriber")
					}
				}()
			}
			wg.Wait()
		}()
	}
	wg.Wait()
}
