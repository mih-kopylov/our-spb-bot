package spb

import (
	"github.com/joomcode/errorx"
)

type Client interface {
	Login(login string, password string) (*TokenResponse, error)
	GetNearestBuildings(latitude float64, longitude float64) (*NearestBuildingResponse, error)
	GetReasons() ([]CityResponse, error)
	Send(token string, fields map[string]string, files map[string][]byte) (*SentMessageResponse, error)
	CreateSendProblemRequest(reasonId int64, body string, latitude float64, longitude float64) (map[string]string, error)
}

var (
	Errors                        = errorx.NewNamespace("OurSpbHttp")
	ErrTooManyRequests            = Errors.NewType("TooManyRequests")
	ErrBadRequest                 = Errors.NewType("BadRequest")
	ErrFailedRequest              = Errors.NewType("FailedRequest")
	ErrUnauthorized               = Errors.NewType("Unauthorized")
	ErrExpectingNotBuildingCoords = Errors.NewType("ExpectingNotBuildingCoords")
)
