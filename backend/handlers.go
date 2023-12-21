package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Login struct {
	Username string
	Password string
}

func PostInvalidate(sessions *Sessions) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Request.Cookie("token")

		if err != nil {
			c.SetCookie("token", "", 0, "/", "www2.pr.iva.cy", false, false)
			c.Status(http.StatusBadRequest)
			return
		}

		if token.Value == "" {
			c.SetCookie("token", "", 0, "/", "www2.pr.iva.cy", false, false)
			c.Status(http.StatusBadRequest)
			return
		}

		sessions.Lock.Lock()
		delete(sessions.Sessions, token.Value)
		sessions.Lock.Unlock()

		c.Status(http.StatusOK)
		c.Redirect(http.StatusFound, "/login")
	}
}

func GetLoginHandler(sessions *Sessions) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Request.Cookie("token")

		if err != nil {
			c.SetCookie("token", "", 0, "/", "www2.pr.iva.cy", false, false)
			c.Status(http.StatusUnauthorized)
			return
		}

		if token.Value == "" {
			c.SetCookie("token", "", 0, "/", "www2.pr.iva.cy", false, false)
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

		if err := c.BindJSON(&loginInfo); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// check if user exists
		logins.Lock.Lock()
		if logins.Logins[loginInfo.Username] != loginInfo.Password {
			c.Status(http.StatusUnauthorized)
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

		c.SetCookie("token", userId, 24*3600, "/", "www2.pr.iva.cy", false, false)
		c.Redirect(http.StatusFound, "/account")
	}
}

func GetAccountHandler(sessions *Sessions, accounts *Accounts) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Request.Cookie("token")

		if err != nil {
			c.SetCookie("token", "", 0, "/", "www2.pr.iva.cy", false, false)
			c.Status(http.StatusUnauthorized)
			return
		}

		if token.Value == "" {
			c.SetCookie("token", "", 0, "/", "www2.pr.iva.cy", false, false)
			c.Status(http.StatusUnauthorized)
			return
		}

		sessions.Lock.Lock()
		session, ok := sessions.Sessions[token.Value]
		sessions.Lock.Unlock()

		if !ok {
			c.SetCookie("token", "", 0, "/", "www2.pr.iva.cy", false, false)
			c.Status(http.StatusUnauthorized)
			return
		}

		accounts.Lock.Lock()
		account, ok := accounts.Accounts[session.Username]
		accounts.Lock.Unlock()

		c.JSON(http.StatusOK, account)
	}
}

type TransferRequest struct {
	To     string  `form:"to" binding:"required"`
	Amount float32 `form:"amount" binding:"required"`
}

func PostTransferHandler(sessions *Sessions, accounts *Accounts) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Request.Cookie("token")

		if err != nil {
			c.SetCookie("token", "", 0, "/", "www2.pr.iva.cy", false, false)
			c.Status(http.StatusUnauthorized)
			return
		}

		if token.Value == "" {
			c.SetCookie("token", "", 0, "/", "www2.pr.iva.cy", false, false)
			c.Status(http.StatusUnauthorized)
			return
		}

		sessions.Lock.Lock()
		session, ok := sessions.Sessions[token.Value]
		sessions.Lock.Unlock()

		if !ok {
			c.SetCookie("token", "", 0, "/", "www2.pr.iva.cy", false, false)
			c.Status(http.StatusUnauthorized)
			return
		}

		if session.Expiry.Compare(time.Now()) <= 0 {
			c.SetCookie("token", "", 0, "/", "www2.pr.iva.cy", false, false)
			c.Status(http.StatusUnauthorized)
			return
		}

		transferInfo := TransferRequest{}

		if err := c.Bind(&transferInfo); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		accounts.Lock.Lock()
		fromAccount, ok := accounts.Accounts[session.Username]
		if !ok {
			c.Status(http.StatusBadRequest)
			accounts.Lock.Unlock()
			return
		}
		toAccount, ok := accounts.Accounts[transferInfo.To]
		if !ok {
			c.Status(http.StatusBadRequest)
			accounts.Lock.Unlock()
			return
		}
		if fromAccount.Balance < transferInfo.Amount {
			c.Status(http.StatusBadRequest)
			accounts.Lock.Unlock()
			return
		}
		fromAccount.Balance -= transferInfo.Amount
		toAccount.Balance += transferInfo.Amount
		accounts.Accounts[session.Username] = fromAccount
		accounts.Accounts[transferInfo.To] = toAccount
		accounts.Lock.Unlock()
		fmt.Println(toAccount)
		c.Redirect(http.StatusFound, "/account")
	}
}
