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
	"maunium.net/go/mauircd/interfaces"
	"net/http"
	"strings"
)

func getScripts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Add("Allow", http.MethodGet)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	network := string(strings.ToLower(r.RequestURI)[len("/scripts/get/")])

	authd, user := checkAuth(w, r)
	if !authd {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var scripts []mauircdi.Script
	if network == "all" {
		scripts = user.GetGlobalScripts()
		user.GetNetworks().ForEach(func(net mauircdi.Network) {
			scripts = append(scripts, net.GetScripts()...)
		})
	} else if network == "global" {
		scripts = user.GetGlobalScripts()
	} else {
		net := user.GetNetwork(network)
		if net == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		scripts = net.GetScripts()
	}

	data, err := json.Marshal(scripts)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
