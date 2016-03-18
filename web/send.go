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
	"fmt"
	"maunium.net/go/mauircd/irc"
	"net/http"
	"strings"
)

// SendForm ...
type SendForm struct {
	Channel string `json:"channel"`
	Command string `json:"command"`
	Message string `json:"message"`
}

func send(w http.ResponseWriter, r *http.Request) {
	var ip = strings.Split(r.RemoteAddr, ":")[0]
	if r.Method != "POST" {
		w.Header().Add("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var sf SendForm
	err := decoder.Decode(&sf)
	if err != nil || len(sf.Channel) == 0 || len(sf.Command) == 0 || len(sf.Message) == 0 {
		fmt.Printf("%[1]s sent an invalid send request.", ip)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	irc.TmpNet.SendMessage(sf.Channel, sf.Message)
}
