package postgres

import (
	"context"
	"fmt"
	"os"

	logs "github.com/ellexo2456/FilmLib/internal/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetDbParams() string {
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USER")
	pass := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")
	params := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, pass, dbname)

	return params
}

func Connect(ctx context.Context, params string) *pgxpool.Pool {
	dbpool, err := pgxpool.New(ctx, params)
	if err != nil {
		logs.LogFatal(logs.Logger, "postgres_connector", "Connect", err, err.Error())
	}

	err = dbpool.Ping(ctx)
	if err != nil {
		logs.LogFatal(logs.Logger, "postgres_connector", "Connect", err, err.Error())
	}
	logs.Logger.Info("Connected to postgres")
	logs.Logger.Debug("postgres client :", dbpool)

	return dbpool
}
