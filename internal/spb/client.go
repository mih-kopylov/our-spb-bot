package spb

import (
	"github.com/joomcode/errorx"
)

type Client interface {
	Login(login string, password string) (*TokenResponse, error)
	GetNearestBuildings(latitude float64, longitude float64) (*NearestBuildingResponse, error)
	GetReasons() ([]CityResponse, error)
	Send(token string, fields map[string]string, files map[string][]byte) error
	CreateSendProblemRequest(reasonId int64, body string, latitude float64, longitude float64) (map[string]string, error)
}

var (
	OurSpbHttpErrors              = errorx.NewNamespace("OurSpbHttp")
	ErrTooManyRequests            = OurSpbHttpErrors.NewType("TooManyRequests")
	ErrBadRequest                 = OurSpbHttpErrors.NewType("BadRequest")
	ErrUnauthorized               = OurSpbHttpErrors.NewType("Unauthorized")
	ErrExpectingNotBuildingCoords = OurSpbHttpErrors.NewType("ExpectingNotBuildingCoords")
)
