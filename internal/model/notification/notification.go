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
