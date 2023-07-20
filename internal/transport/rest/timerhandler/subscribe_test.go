package timerhandler_test

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
	"testing"

	"github.com/Tap-Team/timerapi/internal/model/timermodel"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func randomTimerList(size int, opts ...randomTimerOption) []*timermodel.Timer {
	tl := make([]*timermodel.Timer, 0, size)
	for i := 0; i < size; i++ {
		tl = append(tl, randomTimer(opts...))
	}
	return tl
}

func userSubscriptions(ctx context.Context, userId int64, offset, limit int) (*httptest.ResponseRecorder, error) {
	v := make(url.Values)
	v.Set("vk_user_id", fmt.Sprint(userId))
	v.Set("limit", fmt.Sprint(limit))
	v.Set("offset", fmt.Sprint(offset))
	req := httptest.NewRequest(http.MethodDelete, basePath("/user-subscriptions?"+v.Encode()), new(bytes.Buffer))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return rec, handler.UserSubscriptions(ctx)(c)
}

func userCreated(ctx context.Context, userId int64, offset, limit int) (*httptest.ResponseRecorder, error) {
	v := make(url.Values)
	v.Set("vk_user_id", fmt.Sprint(userId))
	v.Set("limit", fmt.Sprint(limit))
	v.Set("offset", fmt.Sprint(offset))
	req := httptest.NewRequest(http.MethodDelete, basePath("/user-created?"+v.Encode()), new(bytes.Buffer))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return rec, handler.UserCreated(ctx)(c)
}

func subscribe(ctx context.Context, userId int64, timerId uuid.UUID) (*httptest.ResponseRecorder, error) {
	req := httptest.NewRequest(http.MethodPost, basePath("/:id/subscribe?vk_user_id="+fmt.Sprint(userId)), new(bytes.Buffer))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprint(timerId))
	return rec, handler.Subscribe(ctx)(c)
}

func unsubscribe(ctx context.Context, userId int64, timerId uuid.UUID) (*httptest.ResponseRecorder, error) {
	req := httptest.NewRequest(http.MethodPost, basePath("/:id/subscribe?vk_user_id="+fmt.Sprint(userId)), new(bytes.Buffer))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprint(timerId))
	return rec, handler.Unsubscribe(ctx)(c)
}

func TestUserTimers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const createdAmount = 53
	userId := rand.Int63()
	userCreatedTimers := randomTimerList(createdAmount, func(t *timermodel.Timer) { t.Creator = userId })
	userSubscriptions := randomTimerList(createdAmount * 10)
	createTimersTest(t, ctx, userCreatedTimers)
	createTimersTest(t, ctx, userSubscriptions)

	subscribeTest(t, ctx, userId, userCreatedTimers, userSubscriptions)
	createdTest(t, ctx, userId, userCreatedTimers)

	clearTimers(t, ctx, userCreatedTimers...)
	clearTimers(t, ctx, userSubscriptions...)
}

func TestTimerSubscribers(t *testing.T) {
	var err error
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	subAmount := 100
	userIds := make([]int64, 0, subAmount)
	timer := randomTimer()
	userIds = append(userIds, timer.Creator)
	_, err = createTimer(ctx, timer.Creator, timer.CreateTimer())
	require.NoError(t, err, "failed create timer")

	for i := 0; i < subAmount-1; i++ {
		userId := rand.Int63()
		userIds = append(userIds, userId)
		_, err = subscribe(ctx, userId, timer.ID)
		require.NoError(t, err, "subscribe timer failed")
	}

	subs, err := subscriberStorage.TimerSubscribers(ctx, timer.ID)
	require.NoError(t, err, "failed get timer subscribers")
	subscribers := subs.Array()
	sort.Slice(subscribers, func(i, j int) bool {
		return subscribers[i] < subscribers[j]
	})
	sort.Slice(userIds, func(i, j int) bool {
		return userIds[i] < userIds[j]
	})

	require.Equal(t, len(subscribers), len(userIds), "wrong timer subscribers amount")
	for i := range subscribers {
		require.Equal(t, subscribers[i], userIds[i], "timer subscribers not equal by index %d", i)
	}

	clearTimers(t, ctx, timer)
}

func clearTimers(t *testing.T, ctx context.Context, timers ...*timermodel.Timer) {
	for _, timer := range timers {
		_, err := deleteTimer(ctx, timer.Creator, timer.ID)
		require.NoError(t, err, "failed delete timer")
	}
}

func createTimersTest(t *testing.T, ctx context.Context, timers []*timermodel.Timer) {
	for _, timer := range timers {
		_, err := createTimer(ctx, timer.Creator, timer.CreateTimer())
		require.NoError(t, err, "insert timer failed")
	}
}

func subscribeTest(t *testing.T, ctx context.Context, userId int64, created, subs []*timermodel.Timer) {
	subsId := make([]uuid.UUID, 0)
	// random subscribe by list
	subAmount := 0
	for _, timer := range subs {
		if rand.Int31()%2 == 0 {
			rec, err := subscribe(ctx, userId, timer.ID)
			require.NoError(t, err, "subscribe failed")
			tmb := timerFromBody(t, rec)
			field, ok := tmb.Is(timer)
			require.True(t, ok, fmt.Sprintf("wrong timer from subscribe, field %s no equal", field))
			subsId = append(subsId, timer.ID)
			subAmount++
		}
	}

	// test user subscriptions amount
	var offset, limit = 0, 10
	for offset < subAmount {
		rec, err := userSubscriptions(ctx, userId, offset, limit)
		require.NoError(t, err, "get user subscriptions failed")
		require.Equal(t, 200, rec.Result().StatusCode, "wrong user subscriptions status code")
		timers := timerListFromBody(t, rec)
		l := len(timers)
		require.True(t, l == limit || l == subAmount%limit, "wrong timers amount from subscriptions, %d", l)
		offset += len(timers)
	}
	require.Equal(t, subAmount, offset, "wrong offset amount")

	// random unsubscribe
	for _, id := range subsId {
		if rand.Int31()%2 == 0 {
			rec, err := unsubscribe(ctx, userId, id)
			require.NoError(t, err, "subscribe failed")
			require.Equal(t, http.StatusNoContent, rec.Result().StatusCode)
			subAmount--
		}
	}

	// test user subscriptions after unsubscribe
	offset, limit = 0, 10
	for offset < subAmount {
		rec, err := userSubscriptions(ctx, userId, offset, limit)
		require.NoError(t, err, "get user subscriptions after unsubscribe failed")
		require.Equal(t, http.StatusOK, rec.Result().StatusCode, "wrong user subscriptions after unsubscribe status code")
		timers := timerListFromBody(t, rec)
		l := len(timers)
		require.True(t, l == limit || l == subAmount%limit, "wrong timers amount from subscriptions after unsubscribe, %d", l)
		offset += len(timers)
	}
	require.Equal(t, subAmount, offset, "wrong offset amount after unsubscribe")
}

func createdTest(t *testing.T, ctx context.Context, userId int64, created []*timermodel.Timer) {

	// get timers from storage
	var offset, limit = 0, 10
	timers := make([]*timermodel.Timer, 0)
	for offset < len(created) {
		rec, err := userCreated(ctx, userId, offset, limit)
		require.NoError(t, err, "get user created failed")
		require.Equal(t, http.StatusOK, rec.Result().StatusCode, "wrong created timers status code")
		tl := timerListFromBody(t, rec)
		l := len(tl)
		timers = append(timers, tl...)
		require.True(t, l == limit || l == len(created)%limit, "wrong timers amount from user created %d", l)
		offset += l
	}
	require.Equal(t, len(created), offset)

	// sort slices
	sort.Slice(timers, func(i, j int) bool {
		return timers[i].EndTime.Unix() < timers[j].EndTime.Unix()
	})
	sort.Slice(created, func(i, j int) bool {
		return created[i].EndTime.Unix() < created[j].EndTime.Unix()
	})

	// compare input timers and timers from database
	require.Equal(t, len(created), len(timers), "wrong timers length")
	for i := range created {
		field, ok := created[i].Is(timers[i])
		require.True(t, ok, "timer by index %d field %s no equal", i, field)
	}

}

func timerFromBody(t *testing.T, rec *httptest.ResponseRecorder) *timermodel.Timer {
	b := rec.Body.Bytes()
	timer := new(timermodel.Timer)
	err := json.Unmarshal(b, timer)
	require.NoError(t, err, "failed encode timer from subscribe")
	return timer
}

func timerListFromBody(t *testing.T, rec *httptest.ResponseRecorder) []*timermodel.Timer {
	b := rec.Body.Bytes()
	l := make([]*timermodel.Timer, 0)
	err := json.Unmarshal(b, &l)
	require.NoError(t, err, "failed encode timer list from recorder")
	return l
}
