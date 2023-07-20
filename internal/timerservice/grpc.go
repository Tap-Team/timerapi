package timerservice

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/Tap-Team/timerapi/proto/timerservicepb"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
)

type TimerServiceClient interface {
	Add(ctx context.Context, timerId uuid.UUID, endTime int64) error
	AddMany(ctx context.Context, timers map[uuid.UUID]int64) error
	Start(ctx context.Context, timerId uuid.UUID, endTime int64) error
	Stop(ctx context.Context, timerId uuid.UUID) error
	Remove(ctx context.Context, timerId uuid.UUID) error
	TimerTick(ctx context.Context) (<-chan []uuid.UUID, error)
}

type timerServiceClientGrpc struct {
	client timerservicepb.TimerServiceClient
}

func GrpcClient(client timerservicepb.TimerServiceClient) *timerServiceClientGrpc {
	return &timerServiceClientGrpc{client: client}
}

func (c *timerServiceClientGrpc) Add(ctx context.Context, timerId uuid.UUID, endTime int64) error {
	_, err := c.client.Add(ctx, &timerservicepb.AddEvent{TimerId: timerId[:], EndTime: endTime})
	return err
}
func (c *timerServiceClientGrpc) AddMany(ctx context.Context, timers map[uuid.UUID]int64) error {
	events := make([]*timerservicepb.AddEvent, 0, len(timers))
	for id, endTime := range timers {
		id := id
		events = append(events, &timerservicepb.AddEvent{TimerId: id[:], EndTime: endTime})
	}
	_, err := c.client.AddMany(ctx, &timerservicepb.AddManyEvent{Events: events})
	return err
}
func (c *timerServiceClientGrpc) Start(ctx context.Context, timerId uuid.UUID, endTime int64) error {
	_, err := c.client.Start(ctx, &timerservicepb.StartEvent{TimerId: timerId[:], EndTime: endTime})
	return err
}
func (c *timerServiceClientGrpc) Stop(ctx context.Context, timerId uuid.UUID) error {
	_, err := c.client.Stop(ctx, &timerservicepb.StopEvent{TimerId: timerId[:]})
	return err
}
func (c *timerServiceClientGrpc) Remove(ctx context.Context, timerId uuid.UUID) error {
	_, err := c.client.Remove(ctx, &timerservicepb.RemoveEvent{TimerId: timerId[:]})
	return err
}
func (c *timerServiceClientGrpc) TimerTick(ctx context.Context) (<-chan []uuid.UUID, error) {
	return c.serviceStream(ctx)
}

func (c *timerServiceClientGrpc) serviceStream(ctx context.Context) (<-chan []uuid.UUID, error) {
	ctx, cancel := context.WithCancel(ctx)
	uuidChan := make(chan []uuid.UUID)

	// get tick stream
	tick, err := c.client.TimerTick(ctx, &emptypb.Empty{})
	if err != nil {
		cancel()
		return nil, fmt.Errorf("error while listen timer tick, %w", err)
	}
	go func() {
		// in loop receive values
		for {
			event, err := tick.Recv()
			// if err is io.EOF break Loop
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				log.Fatal("error while receive event from timerservice", err)
				cancel()
				break
			}
			// make list of uuid to send it to chan
			uuids := make([]uuid.UUID, 0, len(event.GetIds()))
			for _, b := range event.GetIds() {
				uuids = append(uuids, uuid.UUID(b))
			}
			uuidChan <- uuids
		}
		close(uuidChan)
	}()
	return uuidChan, nil
}
