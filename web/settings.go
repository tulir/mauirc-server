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
	"net/http"
	"strings"
)

func settings(w http.ResponseWriter, r *http.Request) {
	authd, user := checkAuth(w, r)
	if !authd {
		WriteError(w, ErrNotAuthenticated)
		return
	}

	if r.Method == http.MethodGet {
		getSettings(w, r, user)
	} else if r.Method == http.MethodPut {
		putSettings(w, r, user)
	} else {
		w.Header().Add("Allow", strings.Join([]string{http.MethodGet, http.MethodPut}, ","))
		WriteError(w, ErrInvalidMethod)
	}
}

func putSettings(w http.ResponseWriter, r *http.Request, user mauircdi.User) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		WriteError(w, ErrBodyNotFound)
		return
	}
	var settings = new(interface{})
	json.Unmarshal(data, &settings)
	user.SetSettings(settings)
}

func getSettings(w http.ResponseWriter, r *http.Request, user mauircdi.User) {
	data, err := json.Marshal(user.GetSettings())
	if err != nil {
		WriteError(w, ErrInternal)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
