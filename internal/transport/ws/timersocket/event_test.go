package timersocket_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/Tap-Team/timerapi/internal/model/timerevent"
	"github.com/Tap-Team/timerapi/internal/model/timermodel"
	"github.com/Tap-Team/timerapi/internal/model/timermodel/timerfields"
	"github.com/Tap-Team/timerapi/pkg/amidstr"
	"github.com/Tap-Team/timerapi/pkg/amidtime"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

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

type timerSettingsOption func(t *timermodel.TimerSettings)

func randomTimerSettings(opts ...timerSettingsOption) *timermodel.TimerSettings {
	settings := timermodel.NewTimerSettings(
		timerfields.Name(amidstr.MakeString(timerfields.NameMaxSize)),
		timerfields.Description(amidstr.MakeString(timerfields.DescriptionMaxSize)),
		timerfields.BLUE,
		rand.Intn(3)%2 == 0,
		amidtime.DateTime(time.Now().Add(time.Duration(rand.Int31())*time.Second)),
	)
	for _, opt := range opts {
		opt(settings)
	}
	return settings
}

func receiveEvent[T timerevent.TimerEvent](
	t *testing.T,
	ctx context.Context,
	conns []*WsConn,
	timers []*timermodel.Timer,
	compare func([]T),
) {
	wg := new(sync.WaitGroup)
	wg.Add(len(conns))
	// for every connection start listener which collect all timer ids for compare with timer list
	for _, conn := range conns {
		conn := conn
		go func() {
			defer wg.Done()
			events := make([]T, 0, len(timers))
			// create ids for compare
			ids := make([]uuid.UUID, 0)
		Loop:
			for {
				select {
				case <-ctx.Done():
					break Loop
				case <-time.After(time.Second):
					break Loop
				case event, ok := <-conn.EventStream():
					if !ok {
						break Loop
					}
					et := event.Type()
					id := event.TimerId().String()
					// require type
					event, ok = event.(T)
					require.True(t, ok, "wrong event type, expect %T, %s, %s", *new(T), et, id)
					ids = append(ids, event.TimerId())
					events = append(events, event.(T))
				}
			}
			// sort ids by string id
			sort.Slice(ids, func(i, j int) bool {
				return ids[i].String() > ids[j].String()
			})
			message, ok := compareIds(ids, timers)
			// compare with subscribe list
			require.True(t, ok, message)
			if compare != nil {
				compare(events)
			}
		}()
	}
	wg.Wait()
}

func compareIds(ids []uuid.UUID, timers []*timermodel.Timer) (string, bool) {
	if len(ids) != len(timers) {
		return fmt.Sprintf("wrong len, len(ids) = %d, len(timers) = %d", len(ids), len(timers)), false
	}
	for i := range timers {
		if timers[i].ID != ids[i] {
			return fmt.Sprintf("uuid not equal, index %d, timers uuid %s, ids uuid %s", i, timers[i].ID, ids[i]), false
		}
	}
	return "", true
}

func stopTimer(ctx context.Context, timerId uuid.UUID, userId int64, pauseTime int64) (*httptest.ResponseRecorder, error) {
	v := make(url.Values)
	v.Set("vk_user_id", fmt.Sprint(userId))
	v.Set("pauseTime", fmt.Sprint(pauseTime))
	req := httptest.NewRequest(http.MethodPatch, basePath("/:id/stop?"+v.Encode()), new(bytes.Buffer))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(timerId.String())
	return rec, handler.StopTimer(ctx)(c)
}

func startTimer(ctx context.Context, timerId uuid.UUID, userId int64) (*httptest.ResponseRecorder, error) {
	v := make(url.Values)
	v.Set("vk_user_id", fmt.Sprint(userId))
	req := httptest.NewRequest(http.MethodPatch, basePath("/:id/start?"+v.Encode()), new(bytes.Buffer))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(timerId.String())
	return rec, handler.StartTimer(ctx)(c)
}

func resetTimer(ctx context.Context, timerId uuid.UUID, userId int64) (*httptest.ResponseRecorder, error) {
	v := make(url.Values)
	v.Set("vk_user_id", fmt.Sprint(userId))
	req := httptest.NewRequest(http.MethodPatch, basePath("/:id/reset?"+v.Encode()), new(bytes.Buffer))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(timerId.String())
	return rec, handler.ResetTimer(ctx)(c)
}

func TestEvents(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	amount := 30
	userIds := []int64{
		rand.Int63(),
		rand.Int63(),
		rand.Int63(),
	}
	conns := make([]*WsConn, 0)
	for _, userId := range userIds {
		conn := NewConn(t, ctx, server, userId)
		conns = append(conns, conn)
		go conn.Listen(t, ctx)
	}
	timers := randomTimerList(amount, func(t *timermodel.Timer) { t.Type = timerfields.COUNTDOWN })

	timersIds := make([]uuid.UUID, 0, len(timers))
	for _, timer := range timers {
		timersIds = append(timersIds, timer.ID)
		_, err := createTimer(ctx, timer.Creator, timer.CreateTimer())
		require.NoError(t, err, "create timer failed")
	}
	sort.Slice(timers, func(i, j int) bool {
		return timers[i].ID.String() > timers[j].ID.String()
	})
	sort.Slice(timersIds, func(i, j int) bool {
		return timersIds[i].String() > timersIds[j].String()
	})

	resetEvent(t, ctx, conns, timers, timersIds)
	stopStartEvent(t, ctx, conns, timers, timersIds)
	updateEvent(t, ctx, conns, timers, timersIds)
	personalEvent(t, ctx, conns, timers, timersIds)
	clearTimers(t, ctx, timers...)

	for _, conn := range conns {
		err := conn.Close()
		require.NoError(t, err, "failed to close conn")
	}
}
func resetEvent(t *testing.T, ctx context.Context, conns []*WsConn, timers []*timermodel.Timer, timersIds []uuid.UUID) {
	x := 5
	// subscribe
	for _, conn := range conns {
		conn.Subscribe(t, ctx, timersIds[:len(timersIds)/x]...)
	}
	// reset timer
	for _, timer := range timers {
		_, err := resetTimer(ctx, timer.ID, timer.Creator)
		require.NoError(t, err, "reset timer failed")
	}
	// func for compare timer end time and event end time
	compare := func(events []*timerevent.ResetEvent) {
		var wg sync.WaitGroup
		wg.Add(len(events))
		for _, event := range events {
			event := event
			go func() {
				defer wg.Done()
				timer, err := timerStorage.Timer(ctx, event.TimerId())
				require.NoError(t, err, "failed to get timer from storage")
				require.Equal(t, timer.EndTime.Unix(), event.EndTime.Unix(), "wrong reset event end time")
			}()
		}
		wg.Wait()
	}

	receiveEvent(t, ctx, conns, timers[:len(timers)/x], compare)

	// unsubscribe
	for _, conn := range conns {
		conn.Unsubscribe(t, ctx, timersIds...)
	}

	// reset timer
	for _, timer := range timers {
		_, err := resetTimer(ctx, timer.ID, timer.Creator)
		require.NoError(t, err, "reset timer failed")
	}

	// expect zero events
	receiveEvent(t, ctx, conns, make([]*timermodel.Timer, 0), compare)
}

func stopStartEvent(t *testing.T, ctx context.Context, conns []*WsConn, timers []*timermodel.Timer, timersIds []uuid.UUID) {
	x := 5
	for _, conn := range conns {
		conn.Subscribe(t, ctx, timersIds[:len(timersIds)/x]...)
	}

	// pause timers
	pauseTime := amidtime.Now().Unix()
	for _, timer := range timers {
		_, err := stopTimer(ctx, timer.ID, timer.Creator, pauseTime)
		require.NoError(t, err, "stop timer failed")
	}

	// compare event pause time and original pause time
	compareStop := func(events []*timerevent.StopEvent) {
		for _, event := range events {
			require.Equal(t, pauseTime, event.PauseTime.Unix(), "wrong stop event pause time")
		}
	}
	receiveEvent(t, ctx, conns, timers[:len(timers)/x], compareStop)

	// start timers
	for _, timer := range timers {
		_, err := startTimer(ctx, timer.ID, timer.Creator)
		require.NoError(t, err, "start timer failed")
	}
	// func for compare event end time and timer end time
	compareStart := func(events []*timerevent.StartEvent) {
		var wg sync.WaitGroup
		wg.Add(len(events))
		for _, event := range events {
			event := event
			go func() {
				defer wg.Done()
				timer, err := timerStorage.Timer(ctx, event.TimerId())
				require.NoError(t, err, "failed to get timer from storage")
				require.Equal(t, timer.EndTime.Unix(), event.EndTime.Unix(), "wrong start event end time")
			}()
		}
		wg.Wait()
	}
	receiveEvent(t, ctx, conns, timers[:len(timers)/x], compareStart)

	// unsubscribe
	for _, conn := range conns {
		conn.Unsubscribe(t, ctx, timersIds...)
	}

	// stop timers without subscribers
	for _, timer := range timers {
		_, err := stopTimer(ctx, timer.ID, timer.Creator, pauseTime)
		require.NoError(t, err, "stop timer failed")
	}
	// expect zero events
	receiveEvent(t, ctx, conns, make([]*timermodel.Timer, 0), compareStop)

	// start timers without subscribers
	for _, timer := range timers {
		_, err := startTimer(ctx, timer.ID, timer.Creator)
		require.NoError(t, err, "start timer failed")
	}
	// expect zero events
	receiveEvent(t, ctx, conns, make([]*timermodel.Timer, 0), compareStart)

}

func updateTimer(ctx context.Context, userId int64, timerId uuid.UUID, settings *timermodel.TimerSettings) (*httptest.ResponseRecorder, error) {
	b, _ := json.Marshal(settings)
	req := httptest.NewRequest(http.MethodPut, basePath("/:id"+"?vk_user_id="+fmt.Sprint(userId)), bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprint(timerId))
	return rec, handler.UpdateTimer(ctx)(c)
}

func updateEvent(t *testing.T, ctx context.Context, conns []*WsConn, timers []*timermodel.Timer, timersIds []uuid.UUID) {
	timerSettings := randomTimerSettings()
	x := 5

	for _, conn := range conns {
		conn.Subscribe(t, ctx, timersIds[:len(timersIds)/x]...)
	}

	for _, timer := range timers {
		_, err := updateTimer(ctx, timer.Creator, timer.ID, timerSettings)
		require.NoError(t, err, "update timer failed")
	}
	compare := func(events []*timerevent.UpdateEvent) {
		for _, event := range events {
			require.Equal(t, event.Name, timerSettings.Name, "name not equal")
			require.Equal(t, event.Description, timerSettings.Description, "description not equal")
			require.Equal(t, event.Color, timerSettings.Color, "color not equal")
			require.Equal(t, event.WithMusic, timerSettings.WithMusic, "with music not equal")
			require.Equal(t, event.EndTime.Unix(), timerSettings.EndTime.Unix(), "end time not equal")
		}
	}
	receiveEvent(t, ctx, conns, timers[:len(timers)/x], compare)

	// test with unsubscribe
	for _, conn := range conns {
		conn.Unsubscribe(t, ctx, timersIds...)
	}

	for _, timer := range timers {
		_, err := updateTimer(ctx, timer.Creator, timer.ID, timerSettings)
		require.NoError(t, err, "update timer failed")
	}
	receiveEvent(t, ctx, conns, make([]*timermodel.Timer, 0), compare)
}

var eventTypes = []timerevent.EventType{timerevent.Stop, timerevent.Reset, timerevent.Update}

func randomEvent() timerevent.EventType {
	i := rand.Intn(len(eventTypes))
	return eventTypes[i]
}

func personalEvent(t *testing.T, ctx context.Context, conns []*WsConn, timers []*timermodel.Timer, timersIds []uuid.UUID) {
	x := len(conns)
	for i := range conns {
		si := len(timersIds) / x * i
		ei := len(timersIds) / x * (i + 1)
		conns[i].Subscribe(t, ctx, timersIds[si:ei]...)
	}
	wg := new(sync.WaitGroup)
	wg.Add(len(conns))
	for i := range conns {
		go func(i int) {
			event := randomEvent()
			si := len(timersIds) / x * i
			ei := len(timersIds) / x * (i + 1)
			defer wg.Done()
			switch event {
			case timerevent.Stop:
				for _, timer := range timers[si:ei] {
					_, err := stopTimer(ctx, timer.ID, timer.Creator, amidtime.Now().Unix())
					require.NoError(t, err, "stop timer failed, personal")
				}
				receiveEvent[*timerevent.StopEvent](t, ctx, []*WsConn{conns[i]}, timers[si:ei], nil)
				for _, timer := range timers[si:ei] {
					_, err := startTimer(ctx, timer.ID, timer.Creator)
					require.NoError(t, err, "start timer failed, personal")
				}
				receiveEvent[*timerevent.StartEvent](t, ctx, []*WsConn{conns[i]}, timers[si:ei], nil)
			case timerevent.Reset:
				for _, timer := range timers[si:ei] {
					_, err := resetTimer(ctx, timer.ID, timer.Creator)
					require.NoError(t, err, "reset timer failed, personal")
				}
				receiveEvent[*timerevent.ResetEvent](t, ctx, []*WsConn{conns[i]}, timers[si:ei], nil)
			case timerevent.Update:
				for _, timer := range timers[si:ei] {
					_, err := updateTimer(ctx, timer.Creator, timer.ID, randomTimerSettings())
					require.NoError(t, err, "update timer failed, personal")
				}
				receiveEvent[*timerevent.UpdateEvent](t, ctx, []*WsConn{conns[i]}, timers[si:ei], nil)
			}
		}(i)
	}
	wg.Wait()
	for i := range conns {
		si := len(timersIds) / x * i
		ei := len(timersIds) / x * (i + 1)
		conns[i].Unsubscribe(t, ctx, timersIds[si:ei]...)
	}
}
