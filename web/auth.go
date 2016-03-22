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
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"github.com/gorilla/sessions"
	"maunium.net/go/mauircd/config"
	"net/http"
	"time"
)

type authform struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

var store *sessions.CookieStore

func initStore(address string) {
	store = sessions.NewCookieStore([]byte("lfoWqz9lTNgkM12KLZ23cT7FXeBlXWFL"))
	store.Options = &sessions.Options{
		Domain:   address,
		Path:     "/",
		MaxAge:   86400 * 30,
		Secure:   true,
		HttpOnly: true,
	}
}

func generateAuthToken() string {
	var authToken string
	b := make([]byte, 32)
	// Fill the byte array with cryptographically random bytes.
	n, err := rand.Read(b)
	if n == len(b) && err == nil {
		authToken = base64.RawStdEncoding.EncodeToString(b)
		return authToken
	}

	return ""
}

func checkAuth(w http.ResponseWriter, r *http.Request) (bool, *config.User) {
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

	var found = false
	var newAt = user.AuthTokens
	for i := 0; i < len(user.AuthTokens); i++ {
		if user.AuthTokens[i].Time < time.Now().Unix() {
			newAt[i] = newAt[len(newAt)-1]
			newAt = newAt[:len(newAt)-1]
		} else if user.AuthTokens[i].Token == token {
			found = true
			break
		}
	}
	user.AuthTokens = newAt

	if !found {
		return false, user
	}
	return true, user
}

func httpAuthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.Header().Add("Allow", "GET")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	success, _ := checkAuth(w, r)
	w.WriteHeader(http.StatusOK)
	if !success {
		w.Write([]byte("false"))
	} else {
		w.Write([]byte("true"))
	}
}

func auth(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Add("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	dec := json.NewDecoder(r.Body)
	var af authform
	err := dec.Decode(&af)

	if err != nil || len(af.Email) == 0 || len(af.Password) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user := config.GetUser(af.Email)
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	} else if !user.CheckPassword(af.Password) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	session, err := store.Get(r, "mauIRC")
	if err != nil {
		session, err = store.New(r, "mauIRC")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	at := generateAuthToken()
	user.AuthTokens = append(user.AuthTokens, config.AuthToken{Token: at, Time: time.Now().Add(30 * 24 * time.Hour).Unix()})

	session.Values["token"] = at
	session.Values["email"] = user.Email

	session.Save(r, w)
}
