package spb

import (
	"github.com/goioc/di"
	"github.com/imroc/req/v3"
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/config"
	"github.com/samber/lo"
	"strings"
	"time"
)

const (
	BeanId = "SpbClient"
)

var (
	OurSpbHttp = errorx.CommonErrors.NewType("OurSpbHttp")
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
		SetBaseURL("https://gorod.gov.spb.ru/api/v4.0").
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
	response, err := request.Post("/token/")
	if err != nil {
		return nil, errorx.EnhanceStackTrace(err, "failed to login: dump=%v", response.Dump())
	}

	if response.IsErrorState() || !response.IsSuccessState() {
		return nil, OurSpbHttp.New("failed to login: error=%v dump=%v", responseError.String(), response.Dump())
	}

	return &result, nil
}

type ErrorResponse struct {
	NonFieldErrors []string `json:"non_field_errors"`
}

func (e *ErrorResponse) String() string {
	return strings.Join(e.NonFieldErrors, "\n")
}
