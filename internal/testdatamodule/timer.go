package testdatamodule

import (
	"math/rand"
	"time"

	"github.com/Tap-Team/timerapi/internal/model/timermodel"
	"github.com/Tap-Team/timerapi/internal/model/timermodel/timerfields"
	"github.com/Tap-Team/timerapi/pkg/amidstr"
	"github.com/Tap-Team/timerapi/pkg/amidtime"
	"github.com/google/uuid"
)

type TimerOption func(t *timermodel.Timer)

func RandomTimer(opts ...TimerOption) *timermodel.Timer {
	duration := rand.Int31()
	timer := timermodel.NewTimer(
		uuid.New(),
		240,
		rand.Int63(),
		amidtime.DateTime(time.Now().Add(time.Second*time.Duration(duration))),
		amidtime.DateTime{},
		timerfields.DATE,
		timerfields.Name(amidstr.MakeString(timerfields.NameMaxSize)),
		timerfields.Description(amidstr.MakeString(timerfields.DescriptionMaxSize)),
		timerfields.BLUE,
		rand.Intn(3)%2 == 0,
		int64(duration),
		false,
	)
	if rand.Int63()%2 == 0 {
		timer.Type = timerfields.COUNTDOWN
	}
	for _, opt := range opts {
		opt(timer)
	}
	return timer
}

func RandomTimerList(size int, opts ...TimerOption) []*timermodel.Timer {
	tl := make([]*timermodel.Timer, 0, size)
	for i := 0; i < size; i++ {
		tl = append(tl, RandomTimer(opts...))
	}
	return tl
}
