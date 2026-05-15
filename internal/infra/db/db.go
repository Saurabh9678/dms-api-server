package db

import (
	"fmt"

	"gorm.io/gorm"
)

type Provider interface {
	Open() (*gorm.DB, error)
}

func Connect(provider Provider) (*gorm.DB, error) {
	connection, err := provider.Open()
	if err != nil {
		return nil, fmt.Errorf("couldn't connect to db: %w", err)
	}
	return connection, nil
}
