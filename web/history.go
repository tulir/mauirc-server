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
	"maunium.net/go/mauircd/database"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func history(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.Header().Add("Allow", "GET")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	authd, user := checkAuth(w, r)
	if !authd {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	n, err := strconv.Atoi(r.URL.Query().Get("n"))
	if err != nil || n <= 0 {
		n = 256
	}

	var results []database.Message

	args := strings.Split(r.RequestURI, "/")[2:]
	if len(args) > 0 && len(args[len(args)-1]) == 0 {
		args = args[:len(args)-1]
	}

	if len(args) == 0 {
		results, err = database.GetHistory(user.GetEmail(), n)
	} else if len(args) == 1 {
		results, err = database.GetNetworkHistory(user.GetEmail(), args[0], n)
	} else {
		channel, _ := url.QueryUnescape(args[1])
		results, err = database.GetChannelHistory(user.GetEmail(), args[0], channel, n)
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json, err := json.Marshal(results)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}
