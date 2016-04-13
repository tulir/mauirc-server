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
	"io/ioutil"
	"maunium.net/go/mauircd/interfaces"
	"maunium.net/go/mauircd/plugin"
	"net/http"
	"strings"
)

const global = "global"

func script(w http.ResponseWriter, r *http.Request) {
	authd, user := checkAuth(w, r)
	if !authd {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	args := strings.Split(r.RequestURI, "/")[2:]
	if r.Method == http.MethodGet {
		getScripts(w, r, args, user)
	} else if r.Method == http.MethodDelete {
		deleteScript(w, r, args, user)
	} else if r.Method == http.MethodPut {
		putScript(w, r, args, user)
	} else {
		w.Header().Add("Allow", http.MethodGet+","+http.MethodDelete+","+http.MethodPut)
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func putScript(w http.ResponseWriter, r *http.Request, args []string, user mauircdi.User) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	script := plugin.Script{Name: args[1], TheScript: string(data)}

	if args[0] == global {
		user.AddGlobalScript(script)
	} else {
		net := user.GetNetwork(args[0])
		if net == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		net.AddScript(script)
	}
}

func getScripts(w http.ResponseWriter, r *http.Request, args []string, user mauircdi.User) {
	var scripts []mauircdi.Script
	if args[0] == "all" {
		scripts = user.GetGlobalScripts()
		user.GetNetworks().ForEach(func(net mauircdi.Network) {
			scripts = append(scripts, net.GetScripts()...)
		})
	} else if args[0] == global {
		scripts = user.GetGlobalScripts()
	} else {
		net := user.GetNetwork(args[0])
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

func deleteScript(w http.ResponseWriter, r *http.Request, args []string, user mauircdi.User) {
	if args[0] == "global" {
		if !user.RemoveGlobalScript(args[1]) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	} else {
		net := user.GetNetwork(args[0])
		if net == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if !net.RemoveScript(args[1]) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
