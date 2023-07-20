package timereventstream_test

import (
	"context"
	"sync"
	"testing"

	"github.com/Tap-Team/timerapi/internal/domain/datastream/timereventstream"
	"github.com/Tap-Team/timerapi/internal/model/timerevent"
	"github.com/Tap-Team/timerapi/pkg/amidtime"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var timerIds = [4]uuid.UUID{
	uuid.New(),
	uuid.New(),
	uuid.New(),
	uuid.New(),
}

var events = []timerevent.TimerEvent{
	timerevent.NewStart(timerIds[0], amidtime.Now()),
	timerevent.NewStop(timerIds[0], amidtime.Now()),

	timerevent.NewStart(timerIds[1], amidtime.Now()),
	timerevent.NewStop(timerIds[1], amidtime.Now()),

	timerevent.NewStart(timerIds[2], amidtime.Now()),
	timerevent.NewStop(timerIds[2], amidtime.Now()),

	timerevent.NewStart(timerIds[3], amidtime.Now()),
	timerevent.NewStop(timerIds[3], amidtime.Now()),
}

func TestEventStream(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	handler := timereventstream.New()

	wg := new(sync.WaitGroup)
	streamNum := 100
	wg.Add(streamNum)
	for i := 0; i < streamNum; i++ {

		// subscribe in first after handle stream in goroutine
		s := handler.NewStream()
		s.Subscribe(timerIds[:]...)
		go func() {
			stream(t, ctx, s, wg)
		}()
	}
	// send events
	for _, event := range events {
		handler.Send(event)
	}
	// wait until all subscribers receive all possible events
	wg.Wait()
	// cancel context and finish test
	cancel()
}

func stream(t *testing.T, ctx context.Context, s timereventstream.Stream, wg *sync.WaitGroup) {
	i := 0
Loop:
	for {
		select {
		case <-ctx.Done():
			// if cancel context and amount of received event now equal total amount of events, then fatal error
			if i != len(events) {
				t.Fatalf("low recieved events, expected %d, actual %d", len(events), i)
			}
			break Loop
		case event, ok := <-s.Stream():
			if !ok {
				break Loop
			}
			// compare event
			require.Equal(t, events[i], event, "event handle error")
			i++

			// if events are over call wg.Done()
			if i == len(events) {
				wg.Done()
			}
		}
	}
	s.Close()

}
