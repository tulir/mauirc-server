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

// Package config contains configurations
package config

import (
	"fmt"
	"github.com/Jeffail/gabs"
	"maunium.net/go/mauircd/database"
	"maunium.net/go/mauircd/interfaces"
	"maunium.net/go/mauircd/plugin"
	"strconv"
	"strings"
)

type cmdResponse struct {
	Success       bool   `json:"success"`
	SimpleMessage string `json:"simple-message"`
	Message       string `json:"message"`
}

type clearhistory struct {
	Network string `json:"network"`
	Channel string `json:"channel"`
}

func (user *userImpl) respond(success bool, simple, message string, args ...interface{}) {
	user.NewMessages <- mauircdi.Message{
		Type: "cmdresponse",
		Object: cmdResponse{
			Success:       success,
			SimpleMessage: simple,
			Message:       fmt.Sprintf(message, args...),
		},
	}
}

// HandleCommand handles mauIRC commands from clients
func (user *userImpl) HandleCommand(data *gabs.Container) {
	typ, ok := data.Path("type").Data().(string)
	if !ok {
		return
	}

	switch typ {
	case "raw":
		user.rawMessage(data)
	case "message":
		user.cmdMessage(data)
	case "userlist":
		user.cmdUserlist(data)
	case "clear":
		user.cmdClearHistory(data)
	case "delete":
		user.cmdDeleteMessage(data)
	case "importscript":
		user.cmdImportScript(data)
	default:
		user.respond(false, "unknown-type", "Unknown message type: %s", typ)
	}
}

func (user *userImpl) rawMessage(data *gabs.Container) {
	network, ok := data.Path("network").Data().(string)
	if !ok {
		return
	}

	net := user.GetNetwork(network)
	if net == nil {
		return
	}

	message, ok := data.Path("message").Data().(string)
	if !ok {
		return
	}

	net.SendRaw(message)
}

func (user *userImpl) cmdImportScript(data *gabs.Container) {
	name, ok := data.Path("name").Data().(string)
	if !ok {
		return
	}

	url, ok := data.Path("url").Data().(string)
	if !ok {
		return
	}

	var network = ""
	if data.Exists("network") {
		network, _ = data.Path("network").Data().(string)
	}

	name = strings.ToLower(name)
	scriptData, err := download(url)

	if err != nil {
		user.respond(false, "download-failed", "Failed to download script from http://pastebin.com/raw/%s", url)
		return
	}

	var net mauircdi.Network

	if len(network) != 0 {
		net = user.GetNetwork(network)
		if net == nil {
			user.respond(false, "no-such-network", "No such network: %s", network)
			return
		}
	}

	/*for i := 0; i < len(scriptList); i++ {
		if scriptList[i].Name == name {
			scriptList[i].TheScript = scriptData
			return
		}
	}
	scriptList = append(scriptList, plugin.Script{TheScript: scriptData, Name: name})*/

	if net != nil {
		net.AddScript(plugin.Script{TheScript: scriptData, Name: name})
		user.respond(true, "script-loaded-network", "Successfully loaded script %s on %s", name, net.GetName())
	} else {
		user.AddGlobalScript(plugin.Script{TheScript: scriptData, Name: name})
		user.respond(true, "script-loaded-global", "Successfully loaded global script %s", name)
	}
}

func (user *userImpl) cmdDeleteMessage(data *gabs.Container) {
	idS, ok := data.Path("id").Data().(string)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(idS, 10, 64)
	if err != nil {
		return
	}

	err = database.DeleteMessage(user.Email, id)
	if err != nil {
		fmt.Printf("<%s> Failed to delete message #%d: %s", user.Email, id, err)
		return
	}

	user.NewMessages <- mauircdi.Message{Type: "delete", Object: id}
}

func (user *userImpl) cmdClearHistory(data *gabs.Container) {
	network, ok := data.Path("network").Data().(string)
	if !ok {
		return
	}

	net := user.GetNetwork(network)
	if net == nil {
		return
	}

	channel, ok := data.Path("channel").Data().(string)
	if !ok {
		return
	}

	err := database.ClearChannel(user.GetEmail(), net.GetName(), channel)
	if err != nil {
		fmt.Printf("<%s> Failed to clear history of %s@%s: %s", user.GetEmail(), channel, net.GetName(), err)
		return
	}
	user.NewMessages <- mauircdi.Message{Type: "clear", Object: clearhistory{Channel: channel, Network: net.GetName()}}
}

func (user *userImpl) cmdUserlist(data *gabs.Container) {
	network, ok := data.Path("network").Data().(string)
	if !ok {
		return
	}

	net := user.GetNetwork(network)
	if net == nil {
		return
	}

	channel, ok := data.Path("channel").Data().(string)
	if !ok {
		return
	}

	cd, ok := net.GetActiveChannels().Get(channel)

	if ok {
		user.NewMessages <- mauircdi.Message{Type: "userlist", Object: cd.GetUsers()}
	}
}

func (user *userImpl) cmdMessage(data *gabs.Container) {
	network, ok := data.Path("network").Data().(string)
	if !ok {
		return
	}

	net := user.GetNetwork(network)
	if net == nil {
		return
	}

	channel, okChan := data.Path("channel").Data().(string)
	command, okCmd := data.Path("command").Data().(string)
	message, okMsg := data.Path("message").Data().(string)
	if !okChan || !okCmd || !okMsg {
		return
	}
	if len(channel) > 0 && len(command) > 0 && len(message) > 0 {
		net.SendMessage(channel, command, message)
	}
}
