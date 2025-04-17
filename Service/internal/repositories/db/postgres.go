package db

import "sync"

type PostgresRepository struct {
	// TODO Pool соединений

	mu sync.RWMutex
}

func New() *PostgresRepository {
	return &PostgresRepository{}
}
