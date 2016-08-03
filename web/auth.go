// mauIRCd - The IRC bouncer/backend system for mauIRC clients.
// Copyright (C) 2016 Tulir Asokan

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

// Package web contains the HTTP server
package web

import (
	"encoding/json"
	"github.com/gorilla/sessions"
	"maunium.net/go/mauircd/interfaces"
	"net/http"
)

type authform struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

var store *sessions.CookieStore

func initStore(address string) {
	store = sessions.NewCookieStore(config.GetCookieSecret())
	store.Options = &sessions.Options{
		Domain:   address,
		Path:     "/",
		MaxAge:   86400 * 30,
		Secure:   true,
		HttpOnly: true,
	}
}

func checkAuth(w http.ResponseWriter, r *http.Request) (bool, mauircdi.User) {
	session, err := store.Get(r, "mauIRC")
	if err != nil {
		return false, nil
	}

	emailI := session.Values["email"]
	tokenI := session.Values["token"]
	if emailI == nil || tokenI == nil {
		return false, nil
	}
	email := emailI.(string)
	token := tokenI.(string)

	user := config.GetUser(email)
	if user == nil {
		return false, nil
	}

	if !user.CheckAuthToken(token) {
		return false, user
	}
	return true, user
}

func httpAuthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Add("Allow", http.MethodGet)
		WriteError(w, ErrInvalidMethod)
		return
	}

	success, _ := checkAuth(w, r)
	w.WriteHeader(http.StatusOK)
	if !success {
		w.Write([]byte("{\"authenticated\": \"false\"}"))
	} else {
		w.Write([]byte("{\"authenticated\": \"true\"}"))
	}
}

func register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Add("Allow", http.MethodPost)
		WriteError(w, ErrInvalidMethod)
		return
	}

	dec := json.NewDecoder(r.Body)
	var af authform
	err := dec.Decode(&af)

	if err != nil || len(af.Email) == 0 || len(af.Password) == 0 {
		WriteError(w, ErrMissingFields)
		return
	}

	user := config.CreateUser(af.Email, af.Password)
	if user == nil {
		WriteError(w, ErrEmailUsed)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Add("Allow", http.MethodPost)
		WriteError(w, ErrInvalidMethod)
		return
	}

	dec := json.NewDecoder(r.Body)
	var af authform
	err := dec.Decode(&af)

	if err != nil || len(af.Email) == 0 || len(af.Password) == 0 {
		WriteError(w, ErrMissingFields)
		return
	}

	user := config.GetUser(af.Email)
	if user == nil {
		WriteError(w, ErrInvalidCredentials)
		return
	} else if !user.CheckPassword(af.Password) {
		WriteError(w, ErrInvalidCredentials)
		return
	}

	session, err := store.Get(r, "mauIRC")
	if err != nil {
		session, err = store.New(r, "mauIRC")
		if err != nil {
			WriteError(w, ErrCookieFail)
			return
		}
	}

	session.Values["token"] = user.NewAuthToken()
	session.Values["email"] = user.GetEmail()

	session.Save(r, w)
}

func password(w http.ResponseWriter, r *http.Request) {

}

func emailConfirm(w http.ResponseWriter, r *http.Request) {

}
