package main

import (
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Session struct {
	Expiry   time.Time
	Username string
}

type Account struct {
	Username string
	Balance  float32
	IsOpen   bool
}

type Users struct {
	Logins map[string]string
	Lock   sync.Mutex
}

type Sessions struct {
	Sessions map[string]Session
	Lock     sync.Mutex
}

type Accounts struct {
	Accounts map[string]Account
	Lock     sync.Mutex
}

func main() {
	router := gin.Default()

	users := Users{Lock: sync.Mutex{}, Logins: make(map[string]string)}
	sessions := Sessions{Lock: sync.Mutex{}, Sessions: make(map[string]Session)}
	accounts := Accounts{Lock: sync.Mutex{}, Accounts: make(map[string]Account)}

	users.Logins["test"] = "password"
	users.Logins["test2"] = "password"
	users.Logins["alice"] = "crypto"
	users.Logins["bob"] = "secret"
	users.Logins["cryptolicious"] = "uint64_t"

	for u := range users.Logins {
		accounts.Accounts[u] = Account{Username: u, Balance: rand.Float32() * 10000 * rand.Float32(), IsOpen: true}
	}

	accounts.Accounts["shhhhhh"] = Account{Username: "", Balance: 1000000000, IsOpen: false}

	// router.Use(static.Serve("/", static.LocalFile("../frontend/out", true)))
	// remove these for next usage if routes conflict
	router.StaticFile("/", "../pages/home.html")
	router.StaticFile("/account", "../pages/account.html")
	router.StaticFile("/login", "../pages/login.html")
	router.StaticFile("/signup", "../pages/signup.html")

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
		api.GET("/account", GetAccountHandler(&sessions, &accounts))
		api.POST("/transfer", PostTransferHandler(&sessions, &accounts))
	}

	router.Run(":8123")
}
