package timerstorage_test

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/Tap-Team/timerapi/internal/model/timermodel"
	"github.com/Tap-Team/timerapi/internal/model/timermodel/timerfields"
	"github.com/Tap-Team/timerapi/pkg/amidstr"
	"github.com/Tap-Team/timerapi/pkg/amidtime"
	"github.com/stretchr/testify/require"
)

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

func compare(timer *timermodel.Timer, settings *timermodel.TimerSettings) bool {
	if timer.Color != settings.Color {
		return false
	}
	if timer.Name != settings.Name {
		return false
	}
	if timer.Description != settings.Description {
		return false
	}
	if timer.WithMusic != settings.WithMusic {
		return false
	}
	if timer.EndTime.Unix() != settings.EndTime.Unix() {
		return false
	}
	return true
}

func TestUpdateTime(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	timer := randomTimer()
	err := testTimerStorage.InsertDateTimer(ctx, timer.Creator, timer.CreateTimer())
	if err != nil {
		t.Fatalf("insert timer test failed, %s", err)
	}
	timer2 := randomTimer()
	err = testTimerStorage.InsertDateTimer(ctx, timer2.Creator, timer2.CreateTimer())
	if err != nil {
		t.Fatalf("insert timer test failed, %s", err)
	}
	duration := int64(rand.Int31n(10000))
	endTime := amidtime.DateTime(timer.EndTime.T().Add(time.Second * time.Duration(duration)))
	err = testTimerStorage.UpdateTime(ctx, timer.ID, endTime)
	if err != nil {
		t.Fatalf("update timer test failed, %s", err)
	}
	tm, err := testTimerStorage.Timer(ctx, timer.ID)
	if err != nil {
		t.Fatalf("select timer test failed, %s", err)
	}
	u2 := tm.EndTime.T().Unix()
	require.Equal(t, endTime.Unix(), u2, "end time not updated")
}

func TestUpdateTimer(t *testing.T) {
	var err error
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	timer := randomTimer()
	err = testTimerStorage.InsertDateTimer(ctx, timer.Creator, timer.CreateTimer())
	if err != nil {
		t.Fatalf("insert timer test failed, %s", err)
	}
	timer2 := randomTimer()
	err = testTimerStorage.InsertDateTimer(ctx, timer2.Creator, timer2.CreateTimer())
	if err != nil {
		t.Fatalf("insert timer test failed, %s", err)
	}
	// positive/negative endTime option
	settings := randomTimerSettings(func(t *timermodel.TimerSettings) {
		if rand.Int63()%2 == 0 {
			t.EndTime = amidtime.DateTime(time.Unix(timer.EndTime.Unix()-timer.Duration/2, 0))
		}
	})
	addedDuration := settings.EndTime.Unix() - timer.EndTime.Unix()
	err = testTimerStorage.UpdateTimer(ctx, timer.ID, settings)
	if err != nil {
		t.Fatalf("update timer test failed, %s", err)
	}
	tm, err := testTimerStorage.Timer(ctx, timer.ID)
	if err != nil {
		t.Fatalf("select timer test failed, %s", err)
	}
	if !compare(tm, settings) {
		t.Fatal("timer update failed, settings not updated")
	}
	require.Equal(t, timer.Duration+addedDuration, tm.Duration, "duration not updated")
}
