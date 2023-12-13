package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Login struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
}

func PostInvalidate(sessions *Sessions) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Request.Cookie("token")

		if err != nil {
			c.SetCookie("token", "", 0, "/", "localhost", false, false)
			c.Status(http.StatusBadRequest)
			return
		}

		if token.Value == "" {
			c.SetCookie("token", "", 0, "/", "localhost", false, false)
			c.Status(http.StatusBadRequest)
			return
		}

		sessions.Lock.Lock()
		delete(sessions.Sessions, token.Value)
		sessions.Lock.Unlock()

		c.Status(http.StatusOK)
	}
}

func GetLoginHandler(sessions *Sessions) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Request.Cookie("token")

		if err != nil {
			c.SetCookie("token", "", 0, "/", "localhost", false, false)
			c.Status(http.StatusUnauthorized)
			return
		}

		if token.Value == "" {
			c.SetCookie("token", "", 0, "/", "localhost", false, false)
			c.Status(http.StatusUnauthorized)
			return
		}

		sessions.Lock.Lock()
		session, ok := sessions.Sessions[token.Value]
		sessions.Lock.Unlock()

		if !ok {
			c.Status(http.StatusUnauthorized)
			return
		}

		c.JSON(http.StatusOK, session)
	}
}

func PostLoginHandler(sessions *Sessions, logins *Users) gin.HandlerFunc {
	return func(c *gin.Context) {
		loginInfo := Login{}

		if err := c.Bind(&loginInfo); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// check if user exists
		logins.Lock.Lock()
		if logins.Logins[loginInfo.Username] == "" {
			c.Status(http.StatusBadRequest)
			logins.Lock.Unlock()
			return
		}
		logins.Lock.Unlock()

		session := new(Session)
		session.Expiry = time.Now().Add(24 * time.Hour)
		session.Username = loginInfo.Username
		userId := uuid.NewString()

		sessions.Lock.Lock()
		sessions.Sessions[userId] = *session
		sessions.Lock.Unlock()

		c.SetCookie("token", userId, 24*3600, "/", "localhost", false, false)
		c.Status(200)
	}
}
