package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/abeme/go_sm_api/controller"
	"github.com/abeme/go_sm_api/entity"
	"github.com/abeme/go_sm_api/middleware"
	"github.com/abeme/go_sm_api/service"
	"github.com/abeme/go_sm_api/ws"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	r := gin.Default()

	// init DB (SQLite via GORM)
	dbFile := os.Getenv("DB_FILE")
	if dbFile == "" {
		dbFile = "dev.db"
	}
	log.Printf("Opening SQLite database file %s", dbFile)
	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to open sqlite db: %v", err)
	}

	// ✅ Migrate all tables at once
	if err := db.AutoMigrate(
		&entity.User{},
		//&entity.PrivateMessage{},
		&entity.Group{},
		&entity.GroupMember{},
		//&entity.GroupMessage{},
	); err != nil {
		log.Fatalf("migrate failed: %v", err)
	}

	// init redis
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	// services
	userSvc := service.NewUserService(db)
	groupSvc := service.NewGroupService(db, rdb)

	// controllers
	authCtrl := controller.NewAuthController(userSvc)
	groupCtrl := controller.NewGroupController(groupSvc)

	// ws hub
	hub := ws.NewHub(rdb)

	r.POST("/signup", authCtrl.SignUp)
	r.POST("/login", authCtrl.Login)

	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	protected.POST("/groups", groupCtrl.Create)
	protected.POST("/groups/:id/join", groupCtrl.Join)
	protected.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "You are authenticated"})
	})

	// ws endpoint
	r.GET("/ws", func(c *gin.Context) {
		ws.ServeWS(hub, c)
	})

	log.Println("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
