package spb

type Client interface {
	Login(login string, password string) (*TokenResponse, error)
}
