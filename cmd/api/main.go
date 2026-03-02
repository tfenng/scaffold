package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/tfenng/scaffold/internal/api/http"
	"github.com/tfenng/scaffold/internal/cache"
	"github.com/tfenng/scaffold/internal/db"
	"github.com/tfenng/scaffold/internal/repo"
	"github.com/tfenng/scaffold/internal/service"
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	ctx := context.Background()

	if err := godotenv.Load(); err != nil {
		log.Println("warning: .env file not found, using environment variables")
	}

	host := getEnv("POSTGRES_HOST", "localhost")
	dbName := getEnv("POSTGRES_DB", "app")
	user := getEnv("POSTGRES_USER", "xmap")
	password := getEnv("POSTGRES_PASSWORD", "xmap")
	dsn := fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=disable", user, password, host, dbName)

	pool, err := db.NewPostgres(ctx, dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	redisAddr := getEnv("REDIS_ADDR", "127.0.0.1:6379")
	rdb := cache.NewRedis(redisAddr)
	defer func() { _ = rdb.Close() }()

	var userCache *cache.UserCache
	if err := cache.Ping(ctx, rdb); err != nil {
		log.Println("redis unavailable, continue without cache:", err)
		log.Println("cache_mode=no-cache")
	} else {
		userCache = cache.NewUserCache(rdb)
		log.Println("cache_mode=redis")
	}

	txMgr := repo.PgxTxManager{Pool: pool}
	userRepo := repo.NewUserRepo(pool)
	userQueryRepo := repo.NewUserQueryRepo(pool)

	userSvc := &service.UserService{
		Tx: txMgr, Users: userRepo, Query: userQueryRepo, UCache: userCache,
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(cors.Default())
	r.Use(http.ErrorMiddleware())

	h := &http.UserHandler{Svc: userSvc}
	r.GET("/users/:id", h.Get)
	r.POST("/users", h.Create)
	r.GET("/users", h.List)
	r.PUT("/users/:id", h.Update)
	r.DELETE("/users/:id", h.Delete)

	log.Fatal(r.Run(":8080"))
}
