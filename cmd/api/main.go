package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"

	"github.com/tfenng/scaffold/internal/api/http"
	"github.com/tfenng/scaffold/internal/cache"
	"github.com/tfenng/scaffold/internal/db"
	"github.com/tfenng/scaffold/internal/repo"
	"github.com/tfenng/scaffold/internal/service"
)

func main() {
	ctx := context.Background()

	dsn := "postgres://xmap:xmap@localhost:5432/app?sslmode=disable"
	pool, err := db.NewPostgres(ctx, dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	rdb := cache.NewRedis("127.0.0.1:6379")
	if err := cache.Ping(ctx, rdb); err != nil {
		log.Println("redis unavailable, continue without cache:", err)
		// 你也可以在这里决定直接退出
	}

	txMgr := repo.PgxTxManager{Pool: pool}
	userRepo := repo.NewUserRepo(pool)
	userQueryRepo := repo.NewUserQueryRepo(pool)
	userCache := cache.NewUserCache(rdb)

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
