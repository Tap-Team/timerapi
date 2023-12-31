package timerhandler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Tap-Team/timerapi/internal/errorutils/timererror"
	"github.com/Tap-Team/timerapi/internal/model/timermodel"
	"github.com/Tap-Team/timerapi/pkg/amidtime"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func createTimer(ctx context.Context, userId int64, timer *timermodel.CreateTimer) (*httptest.ResponseRecorder, error) {
	b, _ := json.Marshal(timer)
	req := httptest.NewRequest(http.MethodPost, basePath("/create?vk_user_id="+fmt.Sprint(userId)), bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return rec, handler.CreateTimer(ctx)(c)
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

func deleteTimer(ctx context.Context, userId int64, timerId uuid.UUID) (*httptest.ResponseRecorder, error) {
	req := httptest.NewRequest(http.MethodDelete, basePath("/:id"+"?vk_user_id="+fmt.Sprint(userId)), new(bytes.Buffer))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprint(timerId))
	return rec, handler.DeleteTimer(ctx)(c)
}

func getTimer(ctx context.Context, timerId uuid.UUID) (*httptest.ResponseRecorder, error) {
	req := httptest.NewRequest(http.MethodGet, basePath("/:id"), &bytes.Buffer{})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprint(timerId))
	return rec, handler.Timer(ctx)(c)
}

func TestCRUD(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	userId := rand.Int63()
	randomtimer := randomTimer(func(t *timermodel.Timer) { t.Creator = userId })
	settingsList := []*timermodel.TimerSettings{
		randomTimerSettings(),
	}
	// test create timer
	createTimerTest(t, ctx, randomtimer, userId)
	// test update timer
	for _, settings := range settingsList {
		updateTimerTest(t, ctx, randomtimer.ID, userId, settings)
	}
	// test delete timer
	deleteTimerTest(t, ctx, randomtimer.ID, userId)
}

func createTimerTest(t *testing.T, ctx context.Context, ct *timermodel.Timer, userId int64) {
	var err error
	// create timer in first time
	rec, err := createTimer(ctx, userId, ct.CreateTimer())
	require.NoError(t, err, "create timer failed")
	require.Equal(t, http.StatusCreated, rec.Result().StatusCode, "wrong status code of create timer")

	// try create same time twice, except timer exists error
	_, err = createTimer(ctx, userId, ct.CreateTimer())
	if !errors.Is(err, timererror.ExceptionUserAlreadySubscriber()) && !errors.Is(err, timererror.ExceptionTimerExists()) {
		log.Fatalf("wrong create timer response error")
	}

	// get error from storage, it should equal input timer
	timer, err := timerStorage.Timer(ctx, ct.ID)
	require.NoError(t, err, "get timer from storage failed")
	field, ok := timer.Is(ct)
	require.True(t, ok, fmt.Sprintf("timer from storage not equal input timer, field %s", field))

	// get timer from subscribers
	subs, err := subscriberStorage.TimerSubscribers(ctx, timer.ID)
	require.NoError(t, err, "error while get subscribers")
	require.True(t, len(subs) == 1 && subs[userId] == struct{}{}, "wrong data from subscribers storage")
}

func updateTimerTest(t *testing.T, ctx context.Context, timerId uuid.UUID, userId int64, settings *timermodel.TimerSettings) {
	var err error
	// get timer from storage and compare data with input settings
	timer, err := timerStorage.Timer(ctx, timerId)
	require.NoError(t, err, "get timer from storage failed")
	if rand.Int63()%2 == 0 {
		settings.EndTime = amidtime.DateTime(time.Unix(timer.EndTime.Unix()-timer.Duration/2, 0))
	}

	// default update timer
	rec, err := updateTimer(ctx, userId, timerId, settings)
	require.NoError(t, err, "update timer failed")
	require.True(t, rec.Result().StatusCode == http.StatusNoContent, fmt.Sprintf("wrong request status code, %d", rec.Result().StatusCode))

	// update no-exists timer
	_, err = updateTimer(ctx, userId, uuid.New(), randomTimerSettings())
	require.ErrorIs(t, err, timererror.ExceptionTimerNotFound(), "update no exists timer wrong error")

	// get timer from storage and compare data with input settings
	tm, err := timerStorage.Timer(ctx, timerId)
	require.NoError(t, err, "get timer from storage failed")
	compareTimerSettings(tm, settings)
	require.True(t, compareTimerSettings(tm, settings), "timer update failed, timer not updated")

	addedDuration := settings.EndTime.Unix() - timer.EndTime.Unix()
	require.Equal(t, timer.Duration+addedDuration, tm.Duration, "duration not updated")

}

func deleteTimerTest(t *testing.T, ctx context.Context, timerId uuid.UUID, userId int64) {
	var err error
	// delete timer
	rec, err := deleteTimer(ctx, userId, timerId)
	require.NoError(t, err, "delete timer failed")
	require.Equal(t, http.StatusNoContent, rec.Result().StatusCode, "wrong status code of delete timer")

	// make sure timer is deleted
	_, err = timerStorage.Timer(ctx, timerId)
	require.ErrorIs(t, err, timererror.ExceptionTimerNotFound(), "get timer from storage wrong error")

	// make sure timer subscribers deleted from storage
	_, err = subscriberStorage.TimerSubscribers(ctx, timerId)
	require.NoError(t, err, "get timer subscribers wrong error")

	// try delete no exists timer
	_, err = deleteTimer(ctx, userId, timerId)
	require.ErrorIs(t, timererror.ExceptionTimerNotFound(), err, "delete no exists timer wrong error")
}

func TestGetTimer(t *testing.T) {
	const listSize = 100
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	timerList := randomTimerList(listSize)
	for _, timer := range timerList {
		_, err := createTimer(ctx, timer.Creator, timer.CreateTimer())
		require.NoError(t, err, "failed create timer")
	}
	for _, timer := range timerList {
		req, err := getTimer(ctx, timer.ID)
		require.NoError(t, err, "failed to get timer")
		require.Equal(t, http.StatusOK, req.Result().StatusCode, "wrong status code from request")
		tm := new(timermodel.Timer)
		err = json.Unmarshal(req.Body.Bytes(), tm)
		require.NoError(t, err, "failed to encode req body")
		message, ok := timer.Is(tm)
		require.True(t, ok, message)
	}
	clearTimers(t, ctx, timerList...)
}
