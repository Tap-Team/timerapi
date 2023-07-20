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
)

type timerSettingsOption func(t *timermodel.TimerSettings)

func randomTimerSettings(opts ...timerSettingsOption) *timermodel.TimerSettings {
	settings := timermodel.NewTimerSettings(
		timerfields.Name(amidstr.MakeString(timerfields.NameMaxSize)),
		timerfields.Description(amidstr.MakeString(timerfields.DescriptionMaxSize)),
		timerfields.BLUE,
		rand.Intn(3)%2 == 0,
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
	endTime := amidtime.DateTime(time.Now().Add(time.Second * time.Duration(rand.Uint32())))
	err = testTimerStorage.UpdateTime(ctx, timer.ID, endTime)
	if err != nil {
		t.Fatalf("update timer test failed, %s", err)
	}
	timer, err = testTimerStorage.Timer(ctx, timer.ID)
	if err != nil {
		t.Fatalf("select timer test failed, %s", err)
	}
	u1 := timer.EndTime.T().Unix()
	u2 := endTime.T().Unix()
	if u1 != u2 {
		t.Fatalf("timer update failed, settings not updated, expected %d, actual %d", u2, u1)
	}
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
	settings := randomTimerSettings()
	err = testTimerStorage.UpdateTimer(ctx, timer.ID, settings)
	if err != nil {
		t.Fatalf("update timer test failed, %s", err)
	}
	timer, err = testTimerStorage.Timer(ctx, timer.ID)
	if err != nil {
		t.Fatalf("select timer test failed, %s", err)
	}
	if !compare(timer, settings) {
		t.Fatal("timer update failed, settings not updated")
	}
}
