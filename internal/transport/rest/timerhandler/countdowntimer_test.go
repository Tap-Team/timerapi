package timerhandler_test

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/Tap-Team/timerapi/internal/errorutils/timererror"
	"github.com/Tap-Team/timerapi/internal/model/timermodel"
	"github.com/Tap-Team/timerapi/internal/model/timermodel/timerfields"
	"github.com/Tap-Team/timerapi/pkg/amidtime"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

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

func randomPause(until time.Duration) {
	time.Sleep(time.Duration(rand.Int63n(int64(until))))
}

func TestCountdownTimer(t *testing.T) {
	var err error
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	userId := rand.Int63()
	duration := int64(5)
	timer := randomTimer(func(t *timermodel.Timer) {
		t.Creator = userId
		t.EndTime = amidtime.DateTime(time.Now().Add(time.Second * time.Duration(duration)))
		t.Duration = duration
		t.Type = timerfields.COUNTDOWN
	})
	ptime := amidtime.Now()

	_, err = createTimer(ctx, userId, timer.CreateTimer())
	require.NoError(t, err, "create timer failed")

	pauseMaxTime := time.Second * 2
	timer = stop(t, ctx, timer, ptime)
	randomPause(pauseMaxTime)
	timer = start(t, ctx, timer)
	randomPause(pauseMaxTime)
	timer = reset(t, ctx, timer)
	randomPause(pauseMaxTime)
	timer = stop(t, ctx, timer, amidtime.Now())
	randomPause(pauseMaxTime)
	timer = reset(t, ctx, timer)

	clearTimers(t, ctx, timer)
}

func stop(t *testing.T, ctx context.Context, timer *timermodel.Timer, pauseTime amidtime.DateTime) *timermodel.Timer {
	// stop timer right case
	rec, err := stopTimer(ctx, timer.ID, timer.Creator, pauseTime.Unix())
	require.NoError(t, err, "pause timer failed")
	require.Equal(t, http.StatusNoContent, rec.Result().StatusCode, "stop timer wrong status code of pause time")

	// try stop timer from random user
	_, err = stopTimer(ctx, timer.ID, rand.Int63(), pauseTime.Unix())
	require.ErrorIs(t, err, timererror.ExceptionUserForbidden(), "stop timer check access failed")

	// try stop no exists timer
	_, err = stopTimer(ctx, uuid.New(), timer.Creator, pauseTime.Unix())
	require.ErrorIs(t, err, timererror.ExceptionTimerNotFound(), "stop timer storage found no exists timer")

	// try stop timer which already stopped
	_, err = stopTimer(ctx, timer.ID, timer.Creator, pauseTime.Unix())
	require.ErrorIs(t, err, timererror.ExceptionTimerIsPaused(), "stop timer which already stopped")

	tm, err := timerStorage.Timer(ctx, timer.ID)
	require.NoError(t, err, "get timer from storage failed")
	require.Equal(t, pauseTime.Unix(), tm.PauseTime.Unix(), "wrong pause time")
	require.True(t, tm.IsPaused, "timer not paused")
	return tm
}

func start(t *testing.T, ctx context.Context, timer *timermodel.Timer) *timermodel.Timer {
	timeInPause := 3
	time.Sleep(time.Second * time.Duration(timeInPause))
	rec, err := startTimer(ctx, timer.ID, timer.Creator)
	require.NoError(t, err, "pause timer failed")
	require.Equal(t, http.StatusOK, rec.Result().StatusCode, "start timer wrong status code of pause time")

	tm := timerFromBody(t, rec)
	difference := tm.EndTime.Unix() - timer.EndTime.Unix()
	require.True(t, difference >= int64(timeInPause), "stop timer not update end time")

	// start timer with random user id
	_, err = startTimer(ctx, timer.ID, rand.Int63())
	require.ErrorIs(t, err, timererror.ExceptionUserForbidden(), "stop timer check access failed")

	// start timer with non exists timer id
	_, err = startTimer(ctx, uuid.New(), timer.Creator)
	require.ErrorIs(t, err, timererror.ExceptionTimerNotFound(), "storage found no exists timer")

	// start already started timer
	_, err = startTimer(ctx, timer.ID, timer.Creator)
	require.ErrorIs(t, err, timererror.ExceptionTimerIsPlaying(), "start timer which already playing")

	// compare timer from storage
	timer, err = timerStorage.Timer(ctx, timer.ID)
	require.NoError(t, err, "failed get timer from storage")
	field, ok := timer.Is(tm)
	require.True(t, ok, "timer from requrest not equal timer from database, %s no equal", field)

	return timer
}

func reset(t *testing.T, ctx context.Context, timer *timermodel.Timer) *timermodel.Timer {
	now := time.Now()
	rec, err := resetTimer(ctx, timer.ID, timer.Creator)
	require.NoError(t, err, "reset timer failed")
	require.Equal(t, http.StatusOK, rec.Result().StatusCode, "wrong reset status code")

	// check timer from request
	tm := timerFromBody(t, rec)
	require.Equal(t, tm.IsPaused, timer.IsPaused, "reset change pause status")
	if timer.IsPaused {
		require.Equal(t, tm.EndTime.Unix()-tm.PauseTime.Unix(), tm.Duration)
	} else {
		require.Equal(t, tm.EndTime.Unix()-now.Unix(), tm.Duration, "reseted end time non equal")
	}

	// compare timer from storage
	timer, err = timerStorage.Timer(ctx, timer.ID)
	require.NoError(t, err, "failed get timer from storage")
	field, ok := timer.Is(tm)
	require.True(t, ok, "timer from requrest not equal timer from database, %s no equal", field)

	return timer
}
