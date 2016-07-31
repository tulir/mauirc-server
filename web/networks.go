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
	//	"maunium.net/go/mauircd/plugin"
	"net/http"
	"strings"
)

func network(w http.ResponseWriter, r *http.Request) {
	authd, user := checkAuth(w, r)
	if !authd {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	args := strings.Split(r.RequestURI, "/")[2:]
	switch r.Method {
	case http.MethodDelete:
		deleteNetwork(w, r, args, user)
	case http.MethodPut:
		addNetwork(w, r, args, user)
	case http.MethodPost:
		editNetwork(w, r, args, user)
	default:
		w.Header().Add("Allow", http.MethodDelete+","+http.MethodPut+","+http.MethodPost)
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func deleteNetwork(w http.ResponseWriter, r *http.Request, args []string, user mauircdi.User) {
	net := user.GetNetwork(args[0])
	if net != nil {
		go net.Disconnect()
		user.DeleteNetwork(args[0])
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func addNetwork(w http.ResponseWriter, r *http.Request, args []string, user mauircdi.User) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	net, ok := user.CreateNetwork(args[0], data)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	oldNet := user.GetNetwork(args[0])
	if oldNet != nil {
		oldNet.Disconnect()
		user.DeleteNetwork(args[0])
	}

	if !user.AddNetwork(net) {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user.InitNetworks()
	user.SendNetworkData(net)
}

type editResponse struct {
	New mauircdi.NetData `json:"new"`
	Old mauircdi.NetData `json:"old"`
}

func editNetwork(w http.ResponseWriter, r *http.Request, args []string, user mauircdi.User) {
	net := user.GetNetwork(args[0])
	if net == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	defer r.Body.Close()
	var data mauircdi.NetData
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var oldData = net.GetNetData()
	connectedUpdate(net, data, oldData)
	nameUpdates(net, data, oldData)
	addrUpdates(net, data, oldData)

	enc := json.NewEncoder(w)
	err = enc.Encode(editResponse{New: net.GetNetData(), Old: oldData})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func connectedUpdate(net mauircdi.Network, data mauircdi.NetData, oldData mauircdi.NetData) {
	if data.Connected != oldData.Connected {
		if data.Connected {
			net.Connect()
		} else {
			net.Disconnect()
		}
	}
}

func nameUpdates(net mauircdi.Network, data mauircdi.NetData, oldData mauircdi.NetData) {
	if data.Nick != oldData.Nick {
		net.SetNick(data.Nick)
	}

	if data.Realname != oldData.Realname {
		net.SetRealname(data.Realname)
	}

	if data.User != oldData.User {
		net.SetUser(data.User)
	}
}

func addrUpdates(net mauircdi.Network, data mauircdi.NetData, oldData mauircdi.NetData) {
	if data.IP != oldData.IP {
		net.SetIP(data.IP)
	}

	if data.Port != oldData.Port {
		net.SetPort(data.Port)
	}

	if data.SSL != oldData.SSL {
		net.SetSSL(data.SSL)
	}
}
