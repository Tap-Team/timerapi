package timerstorage_test

import (
	"errors"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/Tap-Team/timerapi/internal/errorutils/timererror"
	"github.com/Tap-Team/timerapi/internal/model/timermodel"
	"github.com/Tap-Team/timerapi/internal/model/timermodel/timerfields"
	"github.com/Tap-Team/timerapi/pkg/amidstr"
	"github.com/Tap-Team/timerapi/pkg/amidtime"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
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
		timerfields.Name(amidstr.MakeString(timerfields.NameMaxSize)),
		timerfields.Description(amidstr.MakeString(timerfields.DescriptionMaxSize)),
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

func TestTimerCrud(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	timer := randomTimer()
	t1 := *timer
	err := testTimerStorage.InsertDateTimer(ctx, timer.Creator, timer.CreateTimer())
	if err != nil {
		t.Fatalf("insert timer test failed, %s", err)
	}
	timer, err = testTimerStorage.Timer(ctx, timer.ID)
	if err != nil {
		t.Fatalf("select timer test failed, %s", err)
	}
	if f, eq := timer.Is(&t1); !eq {
		t.Fatalf("timer from database not equal input, field %s not equal", f)
	}
	err = testTimerStorage.DeleteTimer(ctx, timer.ID)
	if err != nil {
		t.Fatalf("delete timer test failed, %s", err)
	}
	_, err = testTimerStorage.Timer(ctx, timer.ID)
	if !errors.Is(err, timererror.ExceptionTimerNotFound) {
		t.Fatalf("test timer failed, timer not deleted, %s", err)
	}
}

func randomTimerList(size int, opts ...randomTimerOption) []*timermodel.Timer {
	tl := make([]*timermodel.Timer, 0, size)
	for i := 0; i < size; i++ {
		tl = append(tl, randomTimer(opts...))
	}
	return tl
}

func TestSubcribe(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	userId := rand.Int63()
	// amount of inserted timers
	tam := 100

	timerList := randomTimerList(tam)
	for _, timer := range timerList {
		err := testTimerStorage.InsertDateTimer(ctx, timer.Creator, timer.CreateTimer())
		if err != nil {
			t.Fatalf("insert timer test failed, %s", err)
		}
	}

	// amount of user subcriptions
	sub := 0
	for _, timer := range timerList {
		if rand.Int63()%2 == 0 {
			// random subscriptions with userId
			err := testTimerStorage.Subscribe(ctx, timer.ID, userId)
			if err != nil {
				t.Fatalf("subcribe timer test failed, %s", err)
			}
			sub++
		} else {
			// create flud data
			uId := rand.Int63()
			err := testTimerStorage.Subscribe(ctx, timer.ID, uId)
			if err != nil {
				t.Fatalf("subcribe timer test failed, %s", err)
			}
		}
	}

	// test user subscribe
	timers, err := testTimerStorage.UserSubscriptions(ctx, userId, 0, math.MaxInt)
	if err != nil {
		t.Fatalf("user subcriptions test failed, %s", err)
	}
	if len(timers) != sub {
		t.Fatalf("subcribe timers len not equal amount of subcriptions, expected %d, actual %d", sub, len(timers))
	}

	// random unsubribe
	for _, timer := range timers {
		err := testTimerStorage.Unsubscribe(ctx, timer.ID, userId)
		if err != nil {
			t.Fatalf("subcribe timer test failed, %s", err)
		}
		sub--
	}

	// test user with unsubcribe
	timers, err = testTimerStorage.UserSubscriptions(ctx, userId, 0, math.MaxInt)
	if err != nil {
		t.Fatalf("user subcriptions test failed, %s", err)
	}
	if len(timers) != sub {
		t.Fatalf("subcribe timers len not equal amount of subcriptions, expected %d, actual %d", sub, len(timers))
	}

	// delete timers
	for _, timer := range timerList {
		err := testTimerStorage.DeleteTimer(ctx, timer.ID)
		if err != nil {
			t.Fatalf("delete timer test failed, %s", err)
		}
	}
}

func TestUserTimers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	userId := rand.Int63()
	tam := 100
	// ids for delete timers
	ids := make([]uuid.UUID, 0, tam*2)

	timerList := randomTimerList(tam)
	// in first create timer with random creator and subcribe with userid
	for _, timer := range timerList {
		ids = append(ids, timer.ID)
		err := testTimerStorage.InsertDateTimer(ctx, timer.Creator, timer.CreateTimer())
		if err != nil {
			t.Fatalf("insert timer test failed, %s", err)
		}
	}

	// amount of user subcriptions
	sub := 0
	for _, timer := range timerList {
		if rand.Int63()%2 == 0 {
			// random subscriptions with userId
			err := testTimerStorage.Subscribe(ctx, timer.ID, userId)
			if err != nil {
				t.Fatalf("subcribe timer test failed, %s", err)
			}
			sub++
		} else {
			// create flud data
			uId := rand.Int63()
			err := testTimerStorage.Subscribe(ctx, timer.ID, uId)
			if err != nil {
				t.Fatalf("subcribe timer test failed, %s", err)
			}
		}
	}

	// insert timer with fixed userId
	timerList = randomTimerList(tam, func(t *timermodel.Timer) { t.Creator = userId })
	for _, timer := range timerList {
		ids = append(ids, timer.ID)
		err := testTimerStorage.InsertDateTimer(ctx, timer.Creator, timer.CreateTimer())
		if err != nil {
			t.Fatalf("insert timer test failed, %s", err)
		}
	}

	timerList, err := testTimerStorage.UserTimers(ctx, userId, 0, (tam+sub)+1)
	if err != nil {
		t.Fatalf("get user timers test failed, %s", err)
	}
	// if len of timerList not equal amount of subcriptions plus user created times
	if len(timerList) != tam+sub {
		t.Fatalf("wrong amount of usertimers, expected %d, actual %d", tam+sub, len(timerList))
	}

	// delete timers
	for _, id := range ids {
		err := testTimerStorage.DeleteTimer(ctx, id)
		if err != nil {
			t.Fatalf("delete timer test failed, %s", err)
		}
	}
}

func TestUserCreatedTimers(t *testing.T) {
	var err error
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	userId := rand.Int63()
	createdSize := 10
	ids := make([]uuid.UUID, 0)
	timerList := randomTimerList(createdSize, func(t *timermodel.Timer) { t.Creator = userId })
	for _, timer := range timerList {
		err := testTimerStorage.InsertDateTimer(ctx, timer.Creator, timer.CreateTimer())
		require.NoError(t, err, "insert timer failed")
		ids = append(ids, timer.ID)
	}
	subSize := 100
	timerList = randomTimerList(subSize)
	for _, timer := range timerList {
		err := testTimerStorage.InsertDateTimer(ctx, timer.Creator, timer.CreateTimer())
		require.NoError(t, err, "insert timer for subscribe failed")
		ids = append(ids, timer.ID)
		err = testTimerStorage.Subscribe(ctx, timer.ID, userId)
		require.NoError(t, err, "subscriber timer failed")
	}

	timerList, err = testTimerStorage.UserCreatedTimers(ctx, userId, 0, math.MaxInt)
	require.NoError(t, err, "get user created timers failed")

	require.Equal(t, createdSize, len(timerList), "wrong amount of created timers")

	timerList, err = testTimerStorage.UserSubscriptions(ctx, userId, 0, math.MaxInt)
	require.NoError(t, err, "get user created timers failed")
	require.Equal(t, subSize, len(timerList), " wrong amount of user subscriptions")

	for _, id := range ids {
		err = testTimerStorage.DeleteTimer(ctx, id)
		require.NoError(t, err, "delete timer failed")
	}
}

func TestUserTimersIsPaused(t *testing.T) {
	var err error
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	userId := rand.Int63()
	dateTimerId := uuid.New()
	countdownTimerId := uuid.New()

	timer := randomTimer(func(t *timermodel.Timer) { t.ID = dateTimerId; t.Creator = userId })
	err = testTimerStorage.InsertDateTimer(ctx, timer.Creator, timer.CreateTimer())
	require.NoError(t, err, "insert date timer failed")
	timer = randomTimer(func(t *timermodel.Timer) { t.ID = countdownTimerId; t.Creator = userId })
	err = testTimerStorage.InsertCountdownTimer(ctx, timer.Creator, timer.CreateTimer())
	require.NoError(t, err, "insert  countdown timer failed")

	err = testTimerStorage.UpdatePauseTime(ctx, countdownTimerId, amidtime.Now(), true)
	require.NoError(t, err, "set countdown timer pause time failed")

	dateTimer, err := testTimerStorage.Timer(ctx, dateTimerId)
	require.NoError(t, err, "get date timer failed")
	require.False(t, dateTimer.IsPaused, "date timer is paused equal true, expected false")

	countdownTimer, err := testTimerStorage.Timer(ctx, countdownTimerId)
	require.NoError(t, err, "get countdown timer failed")
	require.True(t, countdownTimer.IsPaused, "countdown timer is paused equal false, expected true")

	err = testTimerStorage.DeleteTimer(ctx, dateTimerId)
	require.NoError(t, err, "delete date timer failed")

	err = testTimerStorage.DeleteTimer(ctx, countdownTimerId)
	require.NoError(t, err, "delete countdowntimer failed")
}
