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

// Package auth contains the authentication system
package auth

import (
	"github.com/gorilla/sessions"
	"maunium.net/go/mauircd/interfaces"
	"maunium.net/go/mauircd/web/errors"
	"maunium.net/go/mauircd/web/util"
	"maunium.net/go/maulogger"
	"net/http"
)

type authform struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

var store *sessions.CookieStore
var config mauircdi.Configuration
var log = maulogger.CreateSublogger("Web/Auth", maulogger.LevelInfo)

// InitStore initializes the cookie store
func InitStore(cfg mauircdi.Configuration) {
	config = cfg
	store = sessions.NewCookieStore(config.GetCookieSecret())
	store.Options = &sessions.Options{
		Domain:   config.GetExternalAddr(),
		Path:     "/",
		MaxAge:   86400 * 30,
		Secure:   true,
		HttpOnly: true,
	}
}

// Check authentication
func Check(w http.ResponseWriter, r *http.Request) (bool, mauircdi.User) {
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

// HTTPCheck handler
func HTTPCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Add("Allow", http.MethodGet)
		errors.Write(w, errors.InvalidMethod)
		return
	}

	success, _ := Check(w, r)
	log.Debugf("%s checked authentication (Authenticated: %s)\n", util.GetIP(r), success)
	w.WriteHeader(http.StatusOK)
	if !success {
		w.Write([]byte("{\"authenticated\": \"false\"}"))
	} else {
		w.Write([]byte("{\"authenticated\": \"true\"}"))
	}
}
