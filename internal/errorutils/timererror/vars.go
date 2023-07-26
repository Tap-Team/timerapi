package timererror

import (
	"net/http"

	"github.com/Tap-Team/timerapi/pkg/exception"
)

const timerErrType = "timer"

var (
	ExceptionTimerNotFound          = func() exception.Exception { return exception.New(http.StatusNotFound, timerErrType, "not_found") }
	ExceptionTimerExists            = func() exception.Exception { return exception.New(http.StatusBadRequest, timerErrType, "exists") }
	ExceptionCountDownTimerNotFound = func() exception.Exception {
		return exception.New(http.StatusNotFound, timerErrType, "countdown_not_found")
	}
	ExceptionTimerSubscribersNotFound = func() exception.Exception {
		return exception.New(http.StatusNotFound, timerErrType, "subscribers_not_found")
	}
	ExceptionWrongTimerTime = func() exception.Exception { return exception.New(http.StatusBadRequest, timerErrType, "wrong_time") }
	ExceptionNilID          = func() exception.Exception { return exception.New(http.StatusBadRequest, timerErrType, "nil_id") }

	ExceptionUserForbidden = func() exception.Exception { return exception.New(http.StatusForbidden, timerErrType, "user_forbidden") }

	ExceptionColorNotFound = func() exception.Exception { return exception.New(http.StatusNotFound, timerErrType, "color_not_found") }
	ExceptionTypeNotFound  = func() exception.Exception { return exception.New(http.StatusNotFound, timerErrType, "type_not_found") }

	ExceptionStatusNotFound = func() exception.Exception {
		return exception.New(http.StatusNotFound, timerErrType, "status_not_found")
	}

	ExceptionTimerIsPaused  = func() exception.Exception { return exception.New(http.StatusBadRequest, timerErrType, "is_paused") }
	ExceptionTimerIsPlaying = func() exception.Exception { return exception.New(http.StatusBadRequest, timerErrType, "is_playing") }

	ExceptionUserAlreadySubscriber = func() exception.Exception {
		return exception.New(http.StatusBadRequest, timerErrType, "user_already_subscriber")
	}

	ExceptionCreatorUnsubscribe = func() exception.Exception {
		return exception.New(http.StatusBadRequest, timerErrType, "creator_unsubscribe")
	}
)
