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
	"maunium.net/go/mauircd/config"
	"maunium.net/go/mauircd/database"
	"net/http"
)

type historyform struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	N        int    `json:"n"`
}

func history(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Add("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	dec := json.NewDecoder(r.Body)
	var hf historyform
	err := dec.Decode(&hf)

	if err != nil || len(hf.Email) == 0 || len(hf.Password) == 0 || hf.N <= 0 {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user := config.GetUser(hf.Email)
	if !user.CheckPassword(hf.Password) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	results, err := database.GetHistory(user.Email, hf.N)
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
