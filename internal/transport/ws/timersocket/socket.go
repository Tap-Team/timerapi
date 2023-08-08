package timersocket

import (
	"context"
	"net/http"
	"strconv"
	"sync"

	"github.com/Tap-Team/timerapi/internal/model/notification"
	"github.com/Tap-Team/timerapi/internal/model/timerevent"
	"github.com/Tap-Team/timerapi/pkg/exception"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
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
//	@Param		event	body		timerevent.SubscribeEvent		true	"event to add\remove timers from event stream"
//	@Success	200		{object}	notification.NotificationDTO	"notification"
//	@Success	201		{object}	timerevent.ResetEvent			"reset event"
//	@Success	202		{object}	timerevent.StopEvent			"stop event"
//	@Success	203		{object}	timerevent.StartEvent			"start event"
//	@Success	204		{object}	timerevent.UpdateEvent			"update event"
//	@Router		/ws/timer [get]
func (s *TimerSocket) TimerWS(c echo.Context) error {
	// parse user id from query
	userId, err := strconv.ParseInt(c.QueryParam("vk_user_id"), 10, 64)
	if err != nil {
		return exception.Wrap(err, exception.NewCause("parse user id from request", "TimerWS", _PROVIDER))
	}
	// create websocket connection
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return exception.Wrap(err, exception.NewCause("listen service stream", "TimerWS", _PROVIDER))
	}
	defer ws.Close()
	// create context with defer cancel
	ctx, cancel := context.WithCancel(c.Request().Context())
	// log.Printf("WEBSOCKET OPEN, %s", ws.RemoteAddr())
	// create stream for stream timer events
	/*
		Stop
		Start
		Delete
		Update
	*/
	timerEventStream := s.streamer.NewStream()
	defer timerEventStream.Close()

	// stream of expired timers by user
	notificationStream := s.notificationStreamer.NewUserStream(userId)
	defer notificationStream.Close()

	var mu sync.Mutex
	// set close handler
	ws.SetCloseHandler(func(code int, text string) error {
		mu.Lock()
		err := ws.WriteMessage(code, []byte(text))
		mu.Unlock()
		cancel()
		return err
	})
	// chan for listen all events from client
	readStream := WSReadStream(ctx, ws)
	// loop listen
Loop:
	for {
		select {
		// ctx done listener
		case <-ctx.Done():
			break Loop
		// handle notification stream
		case n, ok := <-notificationStream.Stream():
			// log.Printf("WEBSOCKET %s, client notification %s", ws.LocalAddr(), n.Type())
			if !ok {
				break Loop
			}
			// on handle send expired event to client
			mu.Lock()
			ws.WriteJSON(n)
			mu.Unlock()
		// read stream
		case event, ok := <-readStream:
			if !ok {
				break Loop
			}
			// call subscribe/unsubscribe function in substream struct
			// log.Printf("WEBSOCKET %s, client event %s", ws.LocalAddr(), event.Type)
			switch event.Type {
			case timerevent.Subscribe:
				timerEventStream.Subscribe(event.TimerIds...)
			case timerevent.Unsubscribe:
				timerEventStream.Unsubscribe(event.TimerIds...)
			}
		// listen timers events from timers
		case event, ok := <-timerEventStream.Stream():
			// log.Printf("WEBSOCKET %s, event: %s, ok: %t", ws.LocalAddr(), event.Type(), ok)
			if !ok {
				break Loop
			}
			// log.Printf("WEBSOCKET EVENT %s, %s", event.Type(), ws.RemoteAddr())
			// send event from stream
			mu.Lock()
			ws.WriteJSON(event)
			mu.Unlock()
		}
	}
	c.Logger().Infof("WEBSOCKET CLOSED, %s", ws.RemoteAddr())
	return nil
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
				if _, ok := err.(*websocket.CloseError); ok {
					break Loop
				}
				if err != nil {
					continue
				}
				ech <- &event
			}
		}
		close(ech)
		// log.Printf("READ CLOSED, %s", ws.RemoteAddr())
	}()
	return ech
}
