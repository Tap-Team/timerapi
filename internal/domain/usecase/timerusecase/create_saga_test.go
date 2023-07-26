package timerusecase_test

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/Tap-Team/timerapi/internal/domain/usecase/timerusecase"
	"github.com/Tap-Team/timerapi/internal/errorutils/timererror"
	timermodel "github.com/Tap-Team/timerapi/internal/model/timermodel"
	"github.com/Tap-Team/timerapi/internal/model/timermodel/timerfields"
	"github.com/Tap-Team/timerapi/internal/timerservice"
	"github.com/golang/mock/gomock"
	uuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func sleepRandom(d time.Duration) {
	sleepTime := rand.Int63n(int64(d))
	time.Sleep(time.Duration(sleepTime))
}

func checkInService(t *testing.T, ctx context.Context, timerId uuid.UUID) {
	err := timerService.Remove(ctx, timerId)
	s, ok := status.FromError(err)
	require.True(t, ok, "wrong error type")
	require.Equal(t, codes.NotFound, s.Code())
}

func checkInStorage(t *testing.T, ctx context.Context, timerId uuid.UUID) {
	_, err := timerStorage.Timer(ctx, timerId)
	require.ErrorIs(t, err, timererror.ExceptionTimerNotFound(), "timer not deleted")
}

func checkInCache(t *testing.T, ctx context.Context, timerId uuid.UUID) {
	_, err := subscriberStorage.TimerSubscribers(ctx, timerId)
	require.ErrorIs(t, err, timererror.ExceptionTimerSubscribersNotFound(), "timer subscribers not deleted")
}

func TestCreateFailedTimerStorageInsert(t *testing.T) {
	var usecase *timerusecase.UseCase
	var err error
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctrl := gomock.NewController(t)

	// test case when insert timer in database failed
	userId := rand.Int63()
	timer := randomTimer(func(t *timermodel.Timer) { t.Creator = userId })
	var expectedErr error = timererror.ExceptionTimerExists()
	insertFailedTimerStorage := timerusecase.NewMockTimerStorage(ctrl)

	stime := time.Millisecond * 100
	switch timer.Type {
	case timerfields.COUNTDOWN:
		insertFailedTimerStorage.EXPECT().
			InsertCountdownTimer(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
			func(context.Context, int64, *timermodel.CreateTimer) error {
				sleepRandom(stime)
				return expectedErr
			}).Times(1)
	case timerfields.DATE:
		insertFailedTimerStorage.EXPECT().
			InsertDateTimer(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
			func(context.Context, int64, *timermodel.CreateTimer) error {
				sleepRandom(stime)
				return expectedErr
			}).Times(1)
	}

	// test
	usecase = timerusecase.New(insertFailedTimerStorage, subscriberStorage, timerService, esender, nsender)

	err = usecase.Create(ctx, userId, timer.CreateTimer())
	require.ErrorIs(t, err, expectedErr, "wrong error")

	checkInCache(t, ctx, timer.ID)
	checkInService(t, ctx, timer.ID)
}

func TestCreateFailedSubscribeInCache(t *testing.T) {
	var usecase *timerusecase.UseCase
	var err error
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctrl := gomock.NewController(t)

	// test case when insert timer in database failed
	userId := rand.Int63()
	timer := randomTimer(func(t *timermodel.Timer) { t.Creator = userId })
	subscribeFailedCacheStorage := timerusecase.NewMockSubscriberCacheStorage(ctrl)

	stime := time.Millisecond * 100

	expectedErr := timererror.ExceptionUserAlreadySubscriber()
	subscribeFailedCacheStorage.EXPECT().Subscribe(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(context.Context, uuid.UUID, ...int64) error {
			sleepRandom(stime)
			return expectedErr
		},
	).Times(1)

	usecase = timerusecase.New(timerStorage, subscribeFailedCacheStorage, timerService, esender, nsender)

	err = usecase.Create(ctx, userId, timer.CreateTimer())
	require.ErrorIs(t, err, expectedErr, "wrong error from create")

	checkInStorage(t, ctx, timer.ID)
	checkInService(t, ctx, timer.ID)
}

func TestAddToServiceFailed(t *testing.T) {
	var usecase *timerusecase.UseCase
	var err error
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctrl := gomock.NewController(t)

	// test case when insert timer in database failed
	userId := rand.Int63()
	timer := randomTimer(func(t *timermodel.Timer) { t.Creator = userId })

	failedAddTimerService := timerservice.NewMockTimerServiceClient(ctrl)
	expectedErr := status.Error(codes.Canceled, "connection refused")
	stime := time.Millisecond * 100
	failedAddTimerService.EXPECT().Add(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(context.Context, uuid.UUID, int64) error {
			sleepRandom(stime)
			return expectedErr
		},
	).Times(1)

	usecase = timerusecase.New(timerStorage, subscriberStorage, failedAddTimerService, esender, nsender)

	err = usecase.Create(ctx, userId, timer.CreateTimer())
	require.ErrorIs(t, err, expectedErr, "wrong error from create")

	checkInStorage(t, ctx, timer.ID)
	checkInCache(t, ctx, timer.ID)
}
