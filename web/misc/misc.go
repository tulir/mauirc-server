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

// Package misc contains HTTP-only misc handlers
package misc

import (
	"maunium.net/go/maulogger"
	"net/http"
	"strings"
)

var log = maulogger.CreateSublogger("Web/HTTP", maulogger.LevelInfo)

func getIP(r *http.Request) string {
	//if config.TrustHeaders() {
	//	return r.Header.Get("X-Forwarded-For")
	//}
	return strings.Split(r.RemoteAddr, ":")[0]
}
