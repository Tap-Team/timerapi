package timerstorage_test

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	randomUserList = func() []int64 {
		list := make([]int64, 0, 100)
		for i := int64(0); i < 100; i++ {
			list = append(list, rand.Int63())
		}
		return list
	}()
)

func TestTimerWithSubscribers(t *testing.T) {
	var err error
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	timersSize := 100
	timerList := randomTimerList(timersSize)

	for _, timer := range timerList {
		if rand.Intn(5)%2 == 0 {
			err = testTimerStorage.InsertDateTimer(ctx, randomUserList[0], timer.CreateTimer())
			require.NoError(t, err, "insert date timer failed")
		} else {
			err = testTimerStorage.InsertDateTimer(ctx, randomUserList[0], timer.CreateTimer())
			require.NoError(t, err, "insert countdown timer failed")
		}
		err = testTimerStorage.SubscribeAll(ctx, timer.ID, randomUserList[1:]...)
		require.NoError(t, err, "subscribe all error")
	}

	timersSubsList, err := testTimerStorage.TimerWithSubscribers(ctx, 0, math.MaxInt)
	require.NoError(t, err, "get timers with subscribers failed")

	sort.Slice(randomUserList, func(i, j int) bool {
		return randomUserList[i] > randomUserList[j]
	})
	for index, timerWithSubs := range timersSubsList {
		sort.Slice(
			timerWithSubs.Subscribers,
			func(i, j int) bool { return timerWithSubs.Subscribers[i] > timerWithSubs.Subscribers[j] },
		)
		require.True(t, sliceEqual(randomUserList, timerWithSubs.Subscribers), fmt.Sprintf("subscribers not equal by %d index", index))
	}
	for _, timer := range timerList {
		err = testTimerStorage.DeleteTimer(ctx, timer.ID)
		require.NoError(t, err, "delete timer failed")
	}
}

func sliceEqual[T comparable](slice1, slice2 []T) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	for i := range slice1 {
		if slice1[i] != slice2[i] {
			return false
		}
	}
	return true

}
