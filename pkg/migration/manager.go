package migration

import "go.uber.org/zap"

type Manager struct {
	logger     *zap.Logger
	migrations []Migration
}

func NewManager(logger *zap.Logger, migrations []Migration) *Manager {
	return &Manager{
		logger:     logger,
		migrations: migrations,
	}
}

func (m *Manager) RunAllMigrations() error {
	for _, migration := range m.migrations {
		err := migration.Migrate()
		if err != nil {
			m.logger.Error("failed to run migrations", zap.Error(err))
			return err
		}
	}

	return nil
}
