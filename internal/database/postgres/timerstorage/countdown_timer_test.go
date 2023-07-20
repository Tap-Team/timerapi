package timerstorage_test

import (
	"context"
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/Tap-Team/timerapi/internal/errorutils/timererror"
	"github.com/Tap-Team/timerapi/internal/model/timermodel"
	"github.com/Tap-Team/timerapi/internal/model/timermodel/timerfields"
	"github.com/Tap-Team/timerapi/pkg/amidtime"
	"github.com/stretchr/testify/require"
)

func TestCountDownTimerCrud(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	timer := randomTimer(func(t *timermodel.Timer) { t.Type = timerfields.COUNTDOWN })
	defer testTimerStorage.DeleteTimer(ctx, timer.ID)
	t1 := *timer
	err := testTimerStorage.InsertCountdownTimer(ctx, timer.Creator, timer.CreateTimer())
	if err != nil {
		t.Fatalf("insert timer test failed, %s", err)
	}
	// get timer to make sure timer insert is success
	timer, err = testTimerStorage.Timer(ctx, timer.ID)
	if err != nil {
		t.Fatalf("select timer test failed, %s", err)
	}
	// get timer pause to make sure timer insert is success
	_, err = testTimerStorage.TimerPause(ctx, timer.ID)
	if err != nil {
		t.Fatalf("select timer test failed, %s", err)
	}
	// get timer pause to make sure timer insert is success
	_, err = testTimerStorage.CountdownTimer(ctx, timer.ID)
	if err != nil {
		t.Fatalf("select timer test failed, %s", err)
	}

	// check timer input equal timer from database
	if f, eq := timer.Is(&t1); !eq {
		t.Fatalf("timer from database not equal input, field %s not equal", f)
	}

	// check timer is deleted
	err = testTimerStorage.DeleteTimer(ctx, timer.ID)
	if err != nil {
		t.Fatalf("delete timer test failed, %s", err)
	}
	_, err = testTimerStorage.Timer(ctx, timer.ID)
	if !errors.Is(err, timererror.ExceptionTimerNotFound) {
		t.Fatalf("test timer failed, timer not deleted, %s", err)
	}
	_, err = testTimerStorage.TimerPause(ctx, timer.ID)
	if !errors.Is(err, timererror.ExceptionTimerNotFound) {
		t.Fatalf("test timer failed, timer not deleted, %s", err)
	}
}

func TestUpdateCountDownTimer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	timer := randomTimer(func(t *timermodel.Timer) { t.Type = timerfields.COUNTDOWN })
	err := testTimerStorage.InsertCountdownTimer(ctx, timer.Creator, timer.CreateTimer())
	if err != nil {
		t.Fatalf("insert timer test failed, %s", err)
	}
	defer testTimerStorage.DeleteTimer(ctx, timer.ID)
	// pause timer data for update
	pauseTime := amidtime.DateTime(time.Now().Add(time.Second * time.Duration(rand.Uint32())))
	isPaused := rand.Int()%2 == 0
	err = testTimerStorage.UpdatePauseTime(ctx, timer.ID, pauseTime, isPaused)
	if err != nil {
		t.Fatalf("update pause time test failed, %s", err)
	}
	// get countdown timer to compare
	ctTimer, err := testTimerStorage.CountdownTimer(ctx, timer.ID)
	if err != nil {
		t.Fatalf("select countdown timer test failed, %s", err)
	}
	require.Equal(t, ctTimer.PauseTime.T().Unix(), pauseTime.T().Unix(), "update failed, pause time not equal")
	if ctTimer.IsPaused != isPaused {
		t.Fatal("update failed, is paused not equal")
	}
}
