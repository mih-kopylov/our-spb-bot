package state

import (
	"github.com/joomcode/errorx"
	"time"
)

type AccountState string

const (
	AccountStateEnabled  AccountState = "enabled"
	AccountStateDisabled AccountState = "disabled"
)

type Account struct {
	Login                string       `firestore:"login"`
	Password             string       `firestore:"password"`
	Token                string       `firestore:"token"`
	RateLimitedUntil     time.Time    `firestore:"rateLimitedUntil"`
	RateLimitNextDayTime time.Time    `firestore:"rateLimitNextDayTime"`
	State                AccountState `firestore:"state"`
}

func (a *Account) GetStateName() (string, error) {
	switch a.State {
	case AccountStateDisabled:
		return "Отключён", nil
	case AccountStateEnabled:
		return "Включён", nil
	default:
		return "", errorx.IllegalState.New("unsupported account state: %v", a.State)
	}

}
