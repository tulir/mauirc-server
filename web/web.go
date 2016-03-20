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
	"fmt"
	"net/http"
)

// Load the web server
func Load(ip string, port int) {
	http.HandleFunc("/history", history)
	http.HandleFunc("/socket", serveWs)
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", ip, port), nil)
	if err != nil {
		panic(err)
	}
}
