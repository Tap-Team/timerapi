package timererror

import (
	"net/http"

	"github.com/Tap-Team/timerapi/pkg/exception"
)

const timerErrType = "timer"

var (
	ExceptionTimerNotFound            = exception.New(http.StatusNotFound, timerErrType, "not_found")
	ExceptionTimerExists              = exception.New(http.StatusBadRequest, timerErrType, "exists")
	ExceptionCountDownTimerNotFound   = exception.New(http.StatusNotFound, timerErrType, "countdown_not_found")
	ExceptionTimerSubscribersNotFound = exception.New(http.StatusNotFound, timerErrType, "subscribers_not_found")
	ExceptionWrongTimerTime           = exception.New(http.StatusBadRequest, timerErrType, "wrong_time")
	ExceptionNilID                    = exception.New(http.StatusBadRequest, timerErrType, "nil_id")

	ExceptionUserForbidden = exception.New(http.StatusForbidden, timerErrType, "user_forbidden")

	ExceptionColorNotFound  = exception.New(http.StatusNotFound, timerErrType, "color_not_found")
	ExceptionTypeNotFound   = exception.New(http.StatusNotFound, timerErrType, "type_not_found")
	ExceptionStatusNotFound = exception.New(http.StatusNotFound, timerErrType, "status_not_found")

	ExceptionTimerIsPaused  = exception.New(http.StatusBadRequest, timerErrType, "is_paused")
	ExceptionTimerIsPlaying = exception.New(http.StatusBadRequest, timerErrType, "is_playing")

	ExceptionUserAlreadySubscriber = exception.New(http.StatusBadRequest, timerErrType, "user_already_subscriber")

	ExceptionCreatorUnsubscribe = exception.New(http.StatusBadRequest, timerErrType, "creator_unsubscribe")
)
