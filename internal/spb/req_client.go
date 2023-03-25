package spb

import (
	"fmt"
	"github.com/goioc/di"
	"github.com/imroc/req/v3"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/config"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"net/http"
	"strings"
	"time"
)

const (
	BeanId = "SpbClient"
)

type ReqClient struct {
	client   *req.Client
	clientId string
	secret   string
}

func RegisterBean(conf *config.Config) {
	client := newReqClient(conf)
	_ = lo.Must(di.RegisterBeanInstance(BeanId, client))
}

func newReqClient(conf *config.Config) Client {
	client := req.C().
		SetBaseURL("https://gorod.gov.spb.ru").
		// this is to pretend to be an official client
		SetUserAgent("okhttp/2.5.0").
		EnableDumpEachRequest().
		SetTimeout(5 * time.Second)

	return &ReqClient{
		client:   client,
		clientId: conf.OurSpbClientId,
		secret:   conf.OurSpbSecret,
	}
}

func (r *ReqClient) GetNearestBuildings(latitude float64, longitude float64) (*NearestBuildingResponse, error) {
	var result NearestBuildingResponse
	var errorResponse ErrorResponse
	request := r.client.R()
	request.SetSuccessResult(&result)
	request.SetErrorResult(&result)
	response, err := request.
		SetQueryParam("latitude", fmt.Sprint(latitude)).
		SetQueryParam("longitude", fmt.Sprint(longitude)).
		Get("/public_api/maps/get_nearest/")
	if err != nil {
		r.printDebugDump(response)
		return nil, errorx.EnhanceStackTrace(err, "failed to get reasons")
	}
	if response.IsErrorState() || !response.IsSuccessState() {
		return nil, ErrBadRequest.New("failed to get nearest buildings: status=%v, response=%v", response.StatusCode, errorResponse.String())
	}

	return &result, nil
}

func (r *ReqClient) GetReasons() ([]CityResponse, error) {
	var result []CityResponse
	var errorResponse ErrorResponse
	request := r.client.R()
	request.SetSuccessResult(&result)
	request.SetErrorResult(&result)
	response, err := request.Get("/api/v4.0/classifier")
	if err != nil {
		r.printDebugDump(response)
		return nil, errorx.EnhanceStackTrace(err, "failed to get reasons")
	}

	if response.IsErrorState() || !response.IsSuccessState() {
		return nil, ErrBadRequest.New("failed to get reasons: status=%v, response=%v", response.StatusCode, errorResponse.String())
	}

	return result, nil
}

func (r *ReqClient) Send(token string, fields map[string]string, files map[string][]byte) error {
	var errorResponse ErrorResponse
	request := r.client.R()
	request.SetErrorResult(&errorResponse)
	request.SetHeader("Authorization", "Bearer "+token)
	request.SetFormData(fields)
	for fileName, fileBytes := range files {
		request.SetFileBytes("files", fileName, fileBytes)
	}

	response, err := request.Post("/api/v4.0/problems/")
	if err != nil {
		r.printDebugDump(response)
		return errorx.EnhanceStackTrace(err, "failed to send a message")
	}

	if response.IsErrorState() || !response.IsSuccessState() {
		if strings.Contains(errorResponse.String(), "Выберите не дом.") {
			return ErrExpectingNotBuildingCoords.New("failed to send a message, expecting not a building coordinates")
		}
		if strings.Contains(errorResponse.String(), "Вы отправили 10 сообщений за сутки.") {
			return ErrTooManyRequests.Wrap(err, "too many requests")
		}
		if response.StatusCode == http.StatusUnauthorized {
			return ErrUnauthorized.Wrap(err, "token expired")
		}
		return ErrBadRequest.New("failed to send a message: status=%v, response=%v", response.StatusCode, errorResponse.String())
	}

	return nil
}

func (r *ReqClient) CreateSendProblemRequest(reasonId int64, body string, latitude float64, longitude float64) (map[string]string, error) {
	reason, err := r.getReason(reasonId)
	if err != nil {
		return nil, ErrBadRequest.Wrap(err, "failed to get reason")
	}

	result := map[string]string{}

	result["body"] = body
	result["reason"] = fmt.Sprint(reasonId)
	result["latitude"] = fmt.Sprint(latitude)
	result["longitude"] = fmt.Sprint(longitude)
	result["manually_selected_reason"] = "true"

	switch reason.PositionType {
	case PositionTypeBuilding:
		nearestBuilding, err := r.getNearestBuilding(latitude, longitude)
		if err != nil {
			return nil, errorx.EnhanceStackTrace(err, "failed to get nearest building")
		}

		result["building"] = fmt.Sprint(nearestBuilding.Id)
	case PositionTypeStreet:
		//do nothing, no additional fields required
	case PositionTypeNearBuilding:
		fallthrough
	case PositionTypeNearBuilding2:
		nearestBuilding, err := r.getNearestBuilding(latitude, longitude)
		if err != nil {
			return nil, errorx.EnhanceStackTrace(err, "failed to get nearest building")
		}

		result["nearest_building"] = fmt.Sprint(nearestBuilding.Id)
	default:
		return nil, errorx.IllegalArgument.New("unsupported position type: type=%v", reason.PositionType)
	}

	return result, nil

}

func (r *ReqClient) Login(login string, password string) (*TokenResponse, error) {
	var result TokenResponse
	var responseError ErrorResponse
	request := r.client.R()
	request.SetFormData(
		map[string]string{
			"username":      login,
			"password":      password,
			"grant_type":    "password",
			"client_id":     r.clientId,
			"client_secret": r.secret,
		},
	)
	request.SetSuccessResult(&result)
	request.SetErrorResult(&responseError)
	response, err := request.Post("/api/v4.0/token/")
	if err != nil {
		r.printDebugDump(response)
		return nil, errorx.EnhanceStackTrace(err, "failed to login")
	}

	if response.IsErrorState() || !response.IsSuccessState() {
		return nil, ErrUnauthorized.New("failed to login: status=%v, response=%v", response.StatusCode, responseError.String())
	}

	return &result, nil
}

func (r *ReqClient) printDebugDump(response *req.Response) {
	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		println(response.Dump())
	}
}

func (r *ReqClient) getReason(id int64) (*ReasonResponse, error) {
	cityResponses, err := r.GetReasons()
	if err != nil {
		return nil, errorx.EnhanceStackTrace(err, "failed to get reasons")
	}

	for _, cityResponse := range cityResponses {
		for _, categoryResponse := range cityResponse.Categories {
			for _, reason := range categoryResponse.Reasons {
				if reason.Id == id {
					return &reason, nil
				}
			}
		}
	}

	return nil, errorx.IllegalArgument.New("can't find reason: id=%v", id)
}

func (r *ReqClient) getNearestBuilding(latitude float64, longitude float64) (*BuildingResponse, error) {
	nearestBuildings, err := r.GetNearestBuildings(latitude, longitude)
	if err != nil {
		return nil, errorx.EnhanceStackTrace(err, "failed to get nearest buildings")
	}

	if len(nearestBuildings.Buildings) == 0 {
		return nil, errorx.IllegalState.New("no nearest buildings found")
	}

	return &nearestBuildings.Buildings[0], nil
}

type ErrorResponse map[string]any

func (r *ErrorResponse) String() string {
	out, _ := yaml.Marshal(r)
	return string(out)
}
