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
	"io/ioutil"
	"net/http"
	"strings"

	"maunium.net/go/mauirc-server/common/errors"
	"maunium.net/go/mauirc-server/common/messages"
	"maunium.net/go/mauirc-server/interfaces"
	"maunium.net/go/mauirc-server/web/auth"
)

// Network HTTP handler
func Network(w http.ResponseWriter, r *http.Request) {
	authd, user := auth.Check(w, r)
	if !authd {
		errors.Write(w, errors.NotAuthenticated)
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
		errors.Write(w, errors.InvalidMethod)
	}
}

func deleteNetwork(w http.ResponseWriter, r *http.Request, args []string, user interfaces.User) {
	net := user.GetNetwork(args[0])
	if net != nil {
		if net.IsConnected() {
			net.ForceDisconnect()
		}
		if user.DeleteNetwork(args[0]) {
			log.Debugf("%s deleted network %s of %s\n", getIP(r), net.GetName(), user.GetEmail())
			w.WriteHeader(http.StatusOK)
		} else {
			log.Debugf("%s tried to delete network %s of %s, but failed for unknown reasons\n", getIP(r), net.GetName(), user.GetEmail())
			errors.Write(w, errors.Internal)
		}
	} else {
		errors.Write(w, errors.NetworkNotFound)
	}
}

func addNetwork(w http.ResponseWriter, r *http.Request, args []string, user interfaces.User) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errors.Write(w, errors.BodyNotFound)
		return
	}

	net, ok := user.CreateNetwork(args[0], data)
	if !ok {
		errors.Write(w, errors.RequestNotJSON)
		return
	}

	oldNet := user.GetNetwork(args[0])
	if oldNet != nil {
		oldNet.Disconnect()
		user.DeleteNetwork(args[0])
	}

	if !user.AddNetwork(net) {
		log.Debugf("%s tried to create network %s for %s, but failed for unknown reasons\n", getIP(r), net.GetName(), user.GetEmail())
		errors.Write(w, errors.Internal)
		return
	}

	user.InitNetworks()
	user.SendNetworkData(net)
	log.Debugf("%s created network %s for %s\n", getIP(r), net.GetName(), user.GetEmail())
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
	New messages.NetData `json:"new"`
	Old messages.NetData `json:"old"`
}

func editNetwork(w http.ResponseWriter, r *http.Request, args []string, user interfaces.User) {
	net := user.GetNetwork(args[0])
	if net == nil {
		errors.Write(w, errors.NetworkNotFound)
		return
	}

	defer r.Body.Close()
	var data editRequest
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&data)
	if err != nil {
		errors.Write(w, errors.RequestNotJSON)
		return
	}

	var oldData = net.GetNetData()
	nameUpdates(net, data, oldData)
	addrUpdates(net, data, oldData)
	connectedUpdate(net, data, oldData)

	log.Debugf("%s edited network %s of %s\n", getIP(r), net.GetName(), user.GetEmail())

	enc := json.NewEncoder(w)
	err = enc.Encode(editResponse{New: net.GetNetData(), Old: oldData})
	if err != nil {
		errors.Write(w, errors.Internal)
		return
	}
}

func connectedUpdate(net interfaces.Network, data editRequest, oldData messages.NetData) {
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

func nameUpdates(net interfaces.Network, data editRequest, oldData messages.NetData) {
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

func addrUpdates(net interfaces.Network, data editRequest, oldData messages.NetData) {
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
