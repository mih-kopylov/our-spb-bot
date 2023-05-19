package migration

import "go.uber.org/zap"

type Migrations struct {
	logger     *zap.Logger
	migrations []Migration
}

func NewMigrations(logger *zap.Logger, migrations []Migration) *Migrations {
	return &Migrations{
		logger:     logger,
		migrations: migrations,
	}
}

func (m *Migrations) RunAll() error {
	for _, migration := range m.migrations {
		err := migration.Migrate()
		if err != nil {
			m.logger.Error("failed to run migrations", zap.Error(err))
			return err
		}
	}

	return nil
}
