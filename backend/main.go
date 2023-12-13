package main

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
)

type Session struct {
	Expiry   time.Time
	Username string
}

type Users struct {
	Logins map[string]string
	Lock   sync.Mutex
}

type Sessions struct {
	Sessions map[string]Session
	Lock     sync.Mutex
}

func main() {
	router := gin.Default()

	users := Users{Lock: sync.Mutex{}, Logins: make(map[string]string)}
	sessions := Sessions{Lock: sync.Mutex{}, Sessions: make(map[string]Session)}

	users.Logins["test"] = "password"

	router.Use(static.Serve("/", static.LocalFile("../frontend/out", true)))
	api := router.Group("/api")
	{
		api.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "test",
			})
		})
		api.GET("/login", GetLoginHandler(&sessions))
		api.POST("/login", PostLoginHandler(&sessions, &users))
		api.POST("/invalidate", PostInvalidate(&sessions))
	}

	router.Run(":8123")
}
