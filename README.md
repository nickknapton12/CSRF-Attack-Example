# CSRF-Attack-Example

A simple, vulnerable application and a CSRF attack exploiting it.

The vulnerable application is a simple "banking" application in which users have balances, can login, logout and transfer money to other accounts. It is served with go, and the code for it lives in the backed + pages directories. A key part of the backend is the use of cookies. Cookies are given to the user when they log in, and revoked when they log out. Cookies are also presented in every request requiring authentication so the user doesn't have to re enter their password.

The Attack code is a simple html website, consisting of a hidden form that on page load submits. This form contains everything needed to submit a transfer to a predefined user. If the user currently holds a valid session cookie from the "bank", then the vulnerable application allows the attack website to make the transfer request on behalf of the user.

## To run

From backend dir
```zsh
go run main.go handlers.go
```
Visit http://localhost:8123/

## To attack

Open the index.html page in the `attack` folder in your browser.
