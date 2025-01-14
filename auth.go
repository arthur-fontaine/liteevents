package main

import "net/http"

type Auth struct {
	Passphrase string
}

func (a *Auth) CheckRaw(passphrase string) bool {
	return a.Passphrase == passphrase
}

func (a *Auth) Check(req *http.Request) bool {
	cookie, err := req.Cookie("passphrase")
	if err != nil {
		return false
	}
	return a.CheckRaw(cookie.Value)
}
