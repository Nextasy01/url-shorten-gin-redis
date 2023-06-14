package repository

import (
	"context"
	"os"

	"github.com/go-redis/redis/v8"
)

type Database struct {
	Conn context.Context
}

func NewDatabase() Database {
	return Database{
		Conn: context.Background(),
	}
}

func (db *Database) CreateConnection(dbNo int) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("DB_ADDR"),
		Password: os.Getenv("DB_PASS"),
		DB:       dbNo,
	})
	return rdb
}
