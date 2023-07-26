package timersocket

import (
	"context"
	"strconv"

	"github.com/Tap-Team/timerapi/internal/model/notification"
	"github.com/Tap-Team/timerapi/internal/model/timerevent"
	"github.com/Tap-Team/timerapi/pkg/exception"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var (
	upgrader = websocket.Upgrader{}
)

const _PROVIDER = "internal/transport/ws/timersocket"

type Streamer interface {
	NewStream() interface {
		Subscribe(...uuid.UUID)
		Unsubscribe(...uuid.UUID)
		Stream() <-chan timerevent.TimerEvent
		Close()
	}
}

type NotificationStreamer interface {
	NewUserStream(int64) interface {
		Stream() <-chan notification.Notification
		Close()
	}
}

type TimerSocket struct {
	streamer             Streamer
	notificationStreamer NotificationStreamer
}

func New(
	streamer Streamer,
	expTimerStream NotificationStreamer,
) *TimerSocket {
	return &TimerSocket{
		streamer:             streamer,
		notificationStreamer: expTimerStream,
	}
}

func Init(
	e *echo.Group,
	streamer Streamer,
	notificationStreamer NotificationStreamer,
) {
	socket := &TimerSocket{
		streamer:             streamer,
		notificationStreamer: notificationStreamer,
	}

	e.GET("/ws/timer", socket.TimerWS)

}

// WSReadStream godoc
//
//	@Summary	Websocket
//	@Description
//	@Tags		ws
//	@Param		vk_user_id	query	int64	true	"user id"
//	@Param		debug		query	string	false	"you can add secret key to query for debug requests"
//	@Produce	json
//	@Param		subscribe\unsubscribe	body		timerevent.SubscribeEvent		true	"event to add\remove timers from event stream"
//	@Success	200						{object}	notification.NotificationDTO	"notification"
//	@Success	201						{object}	timerevent.ResetEvent			"reset event"
//	@Success	202						{object}	timerevent.StopEvent			"stop event"
//	@Success	203						{object}	timerevent.StartEvent			"start event"
//	@Success	204						{object}	timerevent.UpdateEvent			"update event"
//	@Router		/ws/timer [get]
func (s *TimerSocket) TimerWS(c echo.Context) error {
	// parse user id from query
	userId, err := strconv.ParseInt(c.QueryParam("vk_user_id"), 10, 64)
	if err != nil {
		return exception.Wrap(err, exception.NewCause("parse user id from request", "Websocket", _PROVIDER))
	}
	// create context with defer cancel
	ctx, cancel := context.WithCancel(c.Request().Context())
	defer cancel()

	// create websocket connection
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return exception.Wrap(err, exception.NewCause("listen service stream", "Websocker", _PROVIDER))
	}
	defer ws.Close()

	// create stream for stream timer events
	/*
		Stop
		Start
		Delete
		Update
	*/
	timerEventStream := s.streamer.NewStream()

	// chan for listen all events from client
	readStream := WSReadStream(ctx, ws)

	// stream of expired timers by user
	notficationStream := s.notificationStreamer.NewUserStream(userId)

	// close func
	closeFunc := func() {
		cancel()
		timerEventStream.Close()
		notficationStream.Close()
	}
	defer closeFunc()

	// set close handler
	ws.SetCloseHandler(func(code int, text string) error {
		closeFunc()
		return nil
	})

	// loop listen
	for {
		select {

		// ctx done listener
		case <-ctx.Done():
			return nil
		// handle notification stream
		case n, ok := <-notficationStream.Stream():
			if !ok {
				continue
			}
			// on handle send expired event to client
			ws.WriteJSON(n)

		// read stream
		case event, ok := <-readStream:
			if !ok {
				continue
			}
			// call subscribe/unsubscribe function in substream struct
			switch event.Type {
			case timerevent.Subscribe:
				timerEventStream.Subscribe(event.TimerIds...)
			case timerevent.Unsubscribe:
				timerEventStream.Unsubscribe(event.TimerIds...)
			}
		// listen timers events from timers
		case event, ok := <-timerEventStream.Stream():
			if !ok {
				continue
			}
			// send event from stream
			ws.WriteJSON(event)
		}
	}
}

// stream which read all events from websocket connection and write it to created chan
func WSReadStream(ctx context.Context, ws *websocket.Conn) <-chan *timerevent.SubscribeEvent {
	ech := make(chan *timerevent.SubscribeEvent)
	go func() {
	Loop:
		for {
			select {
			case <-ctx.Done():
				break Loop
			default:
				var event timerevent.SubscribeEvent
				err := ws.ReadJSON(&event)
				if err != nil {
					continue
				}
				ech <- &event
			}
		}
		close(ech)
	}()
	return ech
}
