package db

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresProvider struct {
	dsn string
}

func NewPostgresProvider(rawDSN string) *PostgresProvider {
	return &PostgresProvider{
		dsn: rawDSN,
	}
}

func (p *PostgresProvider) Open() (*gorm.DB, error) {
	conn, err := gorm.Open(postgres.Open(p.dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	return conn, nil
}
