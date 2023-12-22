package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// this file contains the handlers for the api endpoints

// struct for login request format
type Login struct {
	Username string
	Password string
}

// struct for transfer request format
type TransferRequest struct {
	To     string  `form:"to" binding:"required"`
	Amount float32 `form:"amount" binding:"required"`
}

// handler to invalidate a session
func PostInvalidate(sessions *Sessions) gin.HandlerFunc {
	return func(c *gin.Context) {
		// get the token from the cookie
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

		// remove the session from the sessions map
		sessions.Lock.Lock()
		delete(sessions.Sessions, token.Value)
		sessions.Lock.Unlock()

		c.Status(http.StatusOK)
		c.Redirect(http.StatusFound, "/login")
	}
}

// check if a session is valid
func GetLoginHandler(sessions *Sessions) gin.HandlerFunc {
	return func(c *gin.Context) {
		// get the token from the cookie
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

		// check if the session exists
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

// login handler to create a session
func PostLoginHandler(sessions *Sessions, logins *Users) gin.HandlerFunc {
	return func(c *gin.Context) {
		loginInfo := Login{}

		if err := c.BindJSON(&loginInfo); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// check the user exists
		logins.Lock.Lock()
		if logins.Logins[loginInfo.Username] != loginInfo.Password {
			c.Status(http.StatusUnauthorized)
			logins.Lock.Unlock()
			return
		}
		logins.Lock.Unlock()

		// create new session, expiry is 24 hours from now
		session := new(Session)
		session.Expiry = time.Now().Add(24 * time.Hour)
		session.Username = loginInfo.Username
		userId := uuid.NewString()

		sessions.Lock.Lock()
		sessions.Sessions[userId] = *session
		sessions.Lock.Unlock()

		// set the cookie and redirect to the account page
		c.SetCookie("token", userId, 24*3600, "/", "www2.pr.iva.cy", false, false)
		c.Redirect(http.StatusFound, "/account")
	}
}

// handler to get the current user's account
func GetAccountHandler(sessions *Sessions, accounts *Accounts) gin.HandlerFunc {
	return func(c *gin.Context) {
		// get the token from the cookie
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

		// get the session from store
		sessions.Lock.Lock()
		session, ok := sessions.Sessions[token.Value]
		sessions.Lock.Unlock()

		if !ok {
			c.SetCookie("token", "", 0, "/", "www2.pr.iva.cy", false, false)
			c.Status(http.StatusUnauthorized)
			return
		}

		// get the account matching the username
		accounts.Lock.Lock()
		account, ok := accounts.Accounts[session.Username]
		accounts.Lock.Unlock()

		// return the account information
		c.JSON(http.StatusOK, account)
	}
}

// handler to transfer money between accounts
func PostTransferHandler(sessions *Sessions, accounts *Accounts) gin.HandlerFunc {
	return func(c *gin.Context) {
		// get the token from the cookie
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

		// get the session from store
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

		// get transfer information from request
		if err := c.Bind(&transferInfo); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// transfer money between accounts and update balances
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
		// print updated info about the beneficiary account for debugging
		fmt.Println(toAccount)

		// redirect to the account page to update balance
		c.Redirect(http.StatusFound, "/account")
	}
}
