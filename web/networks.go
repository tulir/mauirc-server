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
		WriteError(w, ErrNotAuthenticated)
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
		WriteError(w, ErrInvalidMethod)
	}
}

func deleteNetwork(w http.ResponseWriter, r *http.Request, args []string, user mauircdi.User) {
	net := user.GetNetwork(args[0])
	if net != nil {
		if net.IsConnected() {
			net.ForceDisconnect()
		}
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

type editRequest struct {
	Name            string `json:"name"`
	User            string `json:"user"`
	Realname        string `json:"realname"`
	Nick            string `json:"nick"`
	Connected       string `json:"connected"`
	SSL             string `json:"ssl"`
	IP              string `json:"ip"`
	Port            uint16 `json:"port"`
	ForceDisconnect bool   `json:"forcedisconnect"`
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
	var data editRequest
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&data)
	if err != nil {
		WriteError(w, ErrRequestNotJSON)
		return
	}

	var oldData = net.GetNetData()
	nameUpdates(net, data, oldData)
	addrUpdates(net, data, oldData)
	connectedUpdate(net, data, oldData)

	enc := json.NewEncoder(w)
	err = enc.Encode(editResponse{New: net.GetNetData(), Old: oldData})
	if err != nil {
		WriteError(w, ErrInternal)
		return
	}
}

func connectedUpdate(net mauircdi.Network, data editRequest, oldData mauircdi.NetData) {
	if len(data.Connected) == 0 {
		return
	}
	connected := strings.ToLower(data.Connected) == "true"
	if connected != oldData.Connected {
		if connected {
			net.Connect()
		} else {
			net.Disconnect()
		}
	} else if data.ForceDisconnect {
		net.ForceDisconnect()
	}
}

func nameUpdates(net mauircdi.Network, data editRequest, oldData mauircdi.NetData) {
	if len(data.Name) > 0 && data.Name != oldData.Name {
		net.SetName(data.Name)
	}

	if len(data.Nick) > 0 && data.Nick != oldData.Nick {
		net.SetNick(data.Nick)
	}

	if len(data.Realname) > 0 && data.Realname != oldData.Realname {
		net.SetRealname(data.Realname)
	}

	if len(data.User) > 0 && data.User != oldData.User {
		net.SetUser(data.User)
	}
}

func addrUpdates(net mauircdi.Network, data editRequest, oldData mauircdi.NetData) {
	if len(data.IP) > 0 && data.IP != oldData.IP {
		net.SetIP(data.IP)
	}

	if data.Port > 0 && data.Port != oldData.Port {
		net.SetPort(data.Port)
	}

	if len(data.SSL) > 0 {
		ssl := strings.ToLower(data.SSL) == "true"
		if ssl != oldData.SSL {
			net.SetSSL(ssl)
		}
	}
}
