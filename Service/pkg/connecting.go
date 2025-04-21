package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"time"
)

type Postgres struct {
	maxAttempts int

	Pool *pgxpool.Pool
}

func New(connectionString string) *Postgres {
	return &Postgres{
		Pool: newPool(context.Background(), 10, connectionString),
	}
}

func newPool(ctx context.Context, maxAttempts int, connectionString string) (connectionPool *pgxpool.Pool) {
	var err error

	err = doWithTries(func() error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		connectionPool, err = pgxpool.Connect(ctx, connectionString)
		if err != nil {
			return err
		}

		return nil
	}, maxAttempts, 5*time.Second)
	if err != nil {
		(fmt.Println(err.Error()))
		return nil
	}

	return connectionPool
}

func doWithTries(fn func() error, attempts int, delay time.Duration) (err error) {
	for attempts > 0 {
		if err = fn(); err != nil {
			time.Sleep(delay)
			attempts--

			continue
		}
		return nil
	}
	return err
}
