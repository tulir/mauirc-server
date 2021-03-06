// mauIRC-server - The IRC bouncer/backend system for mauIRC clients.
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
	"net/http"
	"os"

	"github.com/gorilla/context"
	"maunium.net/go/mauirc-server/interfaces"
	"maunium.net/go/mauirc-server/web/auth"
	"maunium.net/go/mauirc-server/web/misc"
	"maunium.net/go/mauirc-server/web/socket"
	"maunium.net/go/mauirc-server/web/util"
	"maunium.net/go/maulogger"
)

var config interfaces.Configuration
var log = maulogger.CreateSublogger("Web", maulogger.LevelInfo)

// Load the web server
func Load(c interfaces.Configuration) {
	log.Debugln("Loading HTTP server")
	config = c
	auth.InitStore(config)
	misc.Init(config)
	util.Init(config)

	http.HandleFunc("/history/", misc.History)
	http.HandleFunc("/script/", misc.Script)
	http.HandleFunc("/network/", misc.Network)
	http.HandleFunc("/settings/", misc.Settings)
	http.HandleFunc("/auth/login", auth.Login)
	http.HandleFunc("/auth/confirm", auth.EmailConfirm)
	http.HandleFunc("/auth/password/reset", auth.PasswordReset)
	http.HandleFunc("/auth/password/forgot", auth.PasswordForgot)
	http.HandleFunc("/auth/password/change", auth.PasswordChange)
	http.HandleFunc("/auth/register", auth.Register)
	http.HandleFunc("/auth/check", auth.HTTPCheck)
	http.HandleFunc("/socket", socket.Serve)
	err := http.ListenAndServe(config.GetAddr(), context.ClearHandler(http.DefaultServeMux))
	if err != nil {
		log.Fatalf("Failed to listen to %s: %s", config.GetAddr(), err)
		log.Parent.Close()
		os.Exit(4)
	}
}
