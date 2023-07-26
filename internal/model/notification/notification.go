package notification

import (
	"github.com/Tap-Team/timerapi/internal/model/timermodel"
	"github.com/google/uuid"
)

type NotificationType string

const (
	Expired NotificationType = "notification_expired"
	Delete  NotificationType = "notification_delete"
)

type Notification interface {
	Type() NotificationType
	TimerId() uuid.UUID
	Timer() timermodel.Timer
}

type NotificationSubscribers interface {
	Notification
	Subscribers() []int64
}

type NotificationDTO struct {
	Ntype  NotificationType `json:"type"`
	NTimer timermodel.Timer `json:"timer"`
}

func (n NotificationDTO) TimerId() uuid.UUID {
	return n.NTimer.ID
}

func (n NotificationDTO) Type() NotificationType {
	return n.Ntype
}

func (n NotificationDTO) Timer() timermodel.Timer {
	return n.NTimer
}

func NewExpired(timer timermodel.Timer) Notification {
	return &NotificationDTO{NTimer: timer, Ntype: Expired}
}

func NewDelete(timer timermodel.Timer) Notification {
	return &NotificationDTO{NTimer: timer, Ntype: Delete}
}

type NotificationDTOSubscribers struct {
	NotificationDTO
	Subs []int64 `json:"subscribers"`
}

func (ns *NotificationDTOSubscribers) Subscribers() []int64 {
	return ns.Subs
}

func NewWithSubscribers(notification Notification, subscribers []int64) NotificationSubscribers {
	return &NotificationDTOSubscribers{
		NotificationDTO: NotificationDTO{
			Ntype:  notification.Type(),
			NTimer: notification.Timer(),
		},
		Subs: subscribers,
	}
}
