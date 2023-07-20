package echoconfig

import (
	"github.com/Tap-Team/timerapi/pkg/exception"
	"github.com/labstack/echo/v4"
)

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func ErrorHandler(e error, c echo.Context) {
	c.Logger().Error(e)
	httpCode := 500
	response := ErrorResponse{
		Code:    "common_internal",
		Message: "internal server error",
	}
	if e, ok := e.(exception.HttpError); ok {
		httpCode = e.HttpCode()
	}
	if e, ok := e.(exception.CodeTypedError); ok {
		response.Code = exception.MakeCode(e)
	}
	response.Message = e.Error()

	c.JSON(httpCode, response)
}
