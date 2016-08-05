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

// Package misc contains HTTP-only misc handlers
package misc

import (
	"encoding/json"
	"maunium.net/go/mauirc-common/messages"
	"maunium.net/go/mauirc-server/database"
	"maunium.net/go/mauirc-server/web/auth"
	"maunium.net/go/mauirc-server/web/errors"
	"maunium.net/go/mauirc-server/web/util"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// History HTTP handler
func History(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.Header().Add("Allow", "GET")
		errors.Write(w, errors.InvalidMethod)
		return
	}

	authd, user := auth.Check(w, r)
	if !authd {
		errors.Write(w, errors.NotAuthenticated)
		return
	}

	n, err := strconv.Atoi(r.URL.Query().Get("n"))
	if err != nil || n <= 0 {
		n = 256
	}

	args := strings.Split(r.RequestURI, "/")[2:]
	if len(args) > 0 && len(args[len(args)-1]) == 0 {
		args = args[:len(args)-1]
	}

	results, err := getHistory(user.GetEmail(), getIP(r), n, args)

	if err != nil {
		errors.Write(w, errors.Internal)
		return
	}

	json, err := json.Marshal(results)
	if err != nil {
		log.Errorln("Error while processing /history request by %s: %s", util.GetIP(r), err)
		errors.Write(w, errors.Internal)
		return
	}
	w.Write(json)
}

func getHistory(email, ip string, n int, args []string) ([]messages.Message, error) {
	if len(args) == 0 {
		log.Debugf("%s requested %d messages of history for %s\n", ip, n, email)
		return database.GetHistory(email, n)
	} else if len(args) == 1 {
		log.Debugf("%s requested %d messages of the history of %s for %s\n", ip, n, args[0], email)
		return database.GetNetworkHistory(email, args[0], n)
	} else {
		channel, _ := url.QueryUnescape(args[1])
		log.Debugf("%s requested %d messages of the history of %s @ %s for %s\n", ip, n, channel, args[0], email)
		return database.GetChannelHistory(email, args[0], channel, n)
	}
}
