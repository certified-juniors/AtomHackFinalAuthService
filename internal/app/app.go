package app

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/certified-juniors/AtomHack/docs"
	auth_http "github.com/certified-juniors/AtomHack/internal/auth/delivery/http"
	auth_postgres "github.com/certified-juniors/AtomHack/internal/auth/repository/postgresql"
	auth_redis "github.com/certified-juniors/AtomHack/internal/auth/repository/redis"
	auth_usecase "github.com/certified-juniors/AtomHack/internal/auth/usecase"
	"github.com/certified-juniors/AtomHack/internal/connectors/postgres"
	"github.com/certified-juniors/AtomHack/internal/connectors/redis"
	logs "github.com/certified-juniors/AtomHack/internal/logger"
	"github.com/certified-juniors/AtomHack/internal/middleware"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger"
)

func StartServer() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	accessLogger := middleware.AccessLogger{
		LogrusLogger: logs.Logger,
	}

	dbParams := postgres.GetDbParams()
	pc := postgres.Connect(ctx, dbParams)
	defer pc.Close()

	rc := redis.Connect()
	defer rc.Close()

	mainRouter := mux.NewRouter()

	authMiddlewareRouter := mainRouter.PathPrefix("/api").Subrouter()

	sr := auth_redis.NewSessionRedisRepository(rc)
	ar := auth_postgres.NewAuthPostgresqlRepository(pc, ctx)

	jwtSecret := []byte(os.Getenv("JWT_SECRET"))

	au := auth_usecase.NewAuthUsecase(ar, sr, jwtSecret)

	auth_http.NewAuthHandler(authMiddlewareRouter, mainRouter, au)

	docs.SwaggerInfo.Host = os.Getenv("SWAGGER_ADDR")
	docs.SwaggerInfo.Schemes = []string{os.Getenv("SWAGGER_SCHEME")}
	mainRouter.PathPrefix("/swagger").Handler(httpSwagger.WrapHandler)

	mw := middleware.NewAuth(au)

	authMiddlewareRouter.Use(mw.IsAuth)
	mainRouter.Use(accessLogger.AccessLogMiddleware)
	mainRouter.Use(middleware.CORS)

	serverPort := ":" + os.Getenv("HTTP_SERVER_PORT")
	logs.Logger.Info("starting server at ", serverPort)

	err = http.ListenAndServe(serverPort, mainRouter)
	if err != nil {
		logs.LogFatal(logs.Logger, "main", "main", err, err.Error())
	}
	logs.Logger.Info("server stopped")
}
