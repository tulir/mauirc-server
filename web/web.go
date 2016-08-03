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
	"github.com/gorilla/context"
	"maunium.net/go/mauircd/interfaces"
	"maunium.net/go/maulogger"
	"net/http"
	"os"
)

type webError struct {
	HTTP      int    `json:"http"`
	Simple    string `json:"error"`
	Human     string `json:"error-humanreadable"`
	ExtraInfo string `json:"error-extrainfo,omitempty"`
}

// Web errors
var (
	ErrInvalidMethod      = webError{HTTP: http.StatusMethodNotAllowed, Simple: "methodnotallowed", Human: "The request method is not allowed.", ExtraInfo: "See the Allow header for a list of allowed headers."}
	ErrInvalidCredentials = webError{HTTP: http.StatusUnauthorized, Simple: "invalidcredentials", Human: "Invalid username or password"}
	ErrEmailUsed          = webError{HTTP: http.StatusForbidden, Simple: "emailused", Human: "The given email is already in use"}
	ErrCookieFail         = webError{HTTP: http.StatusInternalServerError, Simple: "cookiefail", Human: "Failed to find or create cookie store", ExtraInfo: "Try removing all cookies for this site."}
	ErrMissingFields      = webError{HTTP: http.StatusBadRequest, Simple: "badrequest", Human: "Your request is missing one or more required fields"}
	ErrFieldFormatting    = webError{HTTP: http.StatusBadRequest, Simple: "badrequest", Human: "Your request has one or more fields with an invalid format"}
)

// WriteError writes a webError to a http.ResponseWriter
func WriteError(w http.ResponseWriter, err webError) error {
	enc := json.NewEncoder(w)
	w.WriteHeader(err.HTTP)
	return enc.Encode(err)
}

var config mauircdi.Configuration
var log = maulogger.CreateSublogger("Web", maulogger.LevelInfo)

// Load the web server
func Load(c mauircdi.Configuration) {
	log.Debugln("Loading HTTP server")
	config = c
	initStore(config.GetExternalAddr())

	http.HandleFunc("/history/", history)
	http.HandleFunc("/script/", script)
	http.HandleFunc("/network/", network)
	http.HandleFunc("/settings/", settings)
	http.HandleFunc("/auth/login", login)
	http.HandleFunc("/auth/confirm", emailConfirm)
	http.HandleFunc("/auth/password/", password)
	http.HandleFunc("/auth/register", register)
	http.HandleFunc("/auth/check", httpAuthCheck)
	http.HandleFunc("/socket", serveWs)
	err := http.ListenAndServe(config.GetAddr(), context.ClearHandler(http.DefaultServeMux))
	if err != nil {
		log.Fatalf("Failed to listen to %s: %s", config.GetAddr(), err)
		log.Parent.Close()
		os.Exit(4)
	}
}
