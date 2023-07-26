package subscriberstorage_test

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"testing"

	"github.com/Tap-Team/timerapi/internal/database/redis/subscriberstorage"
	"github.com/Tap-Team/timerapi/internal/errorutils/timererror"
	"github.com/Tap-Team/timerapi/pkg/rediscontainer"
	"github.com/google/uuid"
)

var (
	testSubscriberStorage *subscriberstorage.Storage
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	rc, term, err := rediscontainer.New(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer term(ctx)
	testSubscriberStorage = subscriberstorage.New(rc)
	m.Run()
}

func TestSubscribeUnsubscribe(t *testing.T) {
	var err error
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	timerId := uuid.New()
	userId := rand.Int63()
	// test subscribe
	err = testSubscriberStorage.Subscribe(ctx, timerId, userId)
	if err != nil {
		t.Fatalf("subscribe test failed, %s", err)
	}

	// flud subscribe
	for i := 0; i < 100; i++ {
		userid := rand.Int63()
		timerid := uuid.New()
		err = testSubscriberStorage.Subscribe(ctx, timerid, userid)
		if err != nil {
			t.Fatalf("subscribe test failed, %s", err)
		}
	}
	// get timer subscribers
	subscribers, err := testSubscriberStorage.TimerSubscribers(ctx, timerId)
	if err != nil {
		t.Fatalf("get subscirbers test failed,%s", err)
	}
	// if user id not found in timer subscribers fatal with error
	if _, ok := subscribers[userId]; !ok {
		t.Fatal("subscribers test failed, user not subscribe")
	}
	// unsubscribe with single subscribers, timer should be deleted without subscribers
	err = testSubscriberStorage.Unsubscribe(ctx, timerId, userId)
	if err != nil {
		t.Fatalf("unsubscribe test failed, %s", err)
	}
	// make sure timer was deleted
	_, err = testSubscriberStorage.TimerSubscribers(ctx, timerId)
	if !errors.Is(err, timererror.ExceptionTimerSubscribersNotFound()) {
		t.Fatal("unsubscribe test failed, key record not deleted")
	}
}

func TestSubscribeAll(t *testing.T) {
	var err error
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// len of subscribe users
	l := 100
	ids := make([]int64, l)
	for i := range ids {
		ids[i] = int64(i)
	}

	// new timerId
	timerId := uuid.New()
	// subscribe
	err = testSubscriberStorage.Subscribe(ctx, timerId, ids...)
	if err != nil {
		t.Fatalf("timer subscribe all test failed, %s", err)
	}

	// get timer subscribers from redis
	subscribers, err := testSubscriberStorage.TimerSubscribers(ctx, timerId)
	if err != nil {
		t.Fatalf("get timer subscribers test failed, %s", err)
	}
	// if len of subscriber not equal of started length of subscribers fatal with err
	if len(subscribers) != l {
		t.Fatal("len of saved subscribers not equal of len input subscribers")
	}
	// test delete timer
	err = testSubscriberStorage.DeleteTimer(ctx, timerId)
	if err != nil {
		t.Fatalf("delete timer test failed, %s", err)
	}
	// if timer not deleted fatal with error
	_, err = testSubscriberStorage.TimerSubscribers(ctx, timerId)
	if !errors.Is(err, timererror.ExceptionTimerSubscribersNotFound()) {
		t.Fatal("delete test failed, key record not deleted")
	}
}
