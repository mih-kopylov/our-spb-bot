package migration

import (
	"github.com/joomcode/errorx"
	"github.com/mih-kopylov/our-spb-bot/internal/state"
	"github.com/mih-kopylov/our-spb-bot/internal/util"
	"github.com/mih-kopylov/our-spb-bot/pkg/migration"
	"go.uber.org/zap"
	"time"
)

type AccountTimeMigration struct {
	logger       *zap.Logger
	stateManager state.Manager
}

func NewAccountTimeMigration(logger *zap.Logger, stateManager state.Manager) migration.Migration {
	return &AccountTimeMigration{
		logger:       logger,
		stateManager: stateManager,
	}
}

func (m *AccountTimeMigration) Migrate() error {
	m.logger.Info("running account time migration")

	allUserStates, err := m.stateManager.GetAllStates()
	if err != nil {
		return errorx.EnhanceStackTrace(err, "failed to migrate account time")
	}

	for _, userState := range allUserStates {
		migrated := false
		for i := range userState.Accounts {
			if userState.Accounts[i].RateLimitNextDayTime.Equal(time.Time{}) {
				userState.Accounts[i].RateLimitNextDayTime = util.DefaultSendTime
				migrated = true
			}
		}

		if migrated {
			err = m.stateManager.SetState(userState)
			if err != nil {
				return errorx.EnhanceStackTrace(err, "failed to migrate account time")
			}
		}
	}

	m.logger.Info("account time migration completed")
	return nil
}
