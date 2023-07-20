package notificationerror

import (
	"net/http"

	"github.com/Tap-Team/timerapi/pkg/exception"
)

const ETypeNotification = "notification"

var (
	ExceptionNotificationNotFound  = exception.New(http.StatusNotFound, ETypeNotification, "not_found")
	ExceptionDuplicateNotification = exception.New(http.StatusBadRequest, ETypeNotification, "duplicate")
	ExceptionTypeNotFound          = exception.New(http.StatusNotFound, ETypeNotification, "type_not_found")
)
