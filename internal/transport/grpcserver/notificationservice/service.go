package notificationservice

import (
	"context"

	"github.com/Tap-Team/timerapi/internal/model/notification"
	"github.com/Tap-Team/timerapi/proto/notificationservicepb"
	"google.golang.org/protobuf/types/known/emptypb"
)

type NotificationStream interface {
	NewStream() interface {
		Stream() <-chan notification.NotificationSubscribers
		Close()
	}
}

type Service struct {
	notificationStream NotificationStream
	notificationservicepb.UnimplementedNotificationServiceServer
}

func New(
	nstream NotificationStream,
) *Service {
	return &Service{notificationStream: nstream}
}

func (s *Service) NotificationStream(e *emptypb.Empty, stream notificationservicepb.NotificationService_NotificationStreamServer) error {
	notificationStream := s.notificationStream.NewStream()
	defer notificationStream.Close()
	for ntion := range notificationStream.Stream() {
		timer := ntion.Timer()
		stream.Send(
			&notificationservicepb.Notification{
				Type: string(ntion.Type()),
				Timer: &notificationservicepb.Timer{
					Name:        string(timer.Name),
					Description: string(timer.Description),
					Type:        string(timer.Type),
				},
				Subscribers: ntion.Subscribers(),
			},
		)
	}
	return nil
}

func (s *Service) Notifications(ctx context.Context, ids *notificationservicepb.Ids) (*notificationservicepb.RepeatedNotification, error) {
	return s.UnimplementedNotificationServiceServer.Notifications(ctx, ids)
}
