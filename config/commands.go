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
	"strconv"
)

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
	case "kick":
		user.cmdKick(data)
	case "mode":
		user.cmdMode(data)
	case "clear":
		user.cmdClearHistory(data)
	case "close":
		user.cmdCloseChannel(data)
	case "open":
		user.cmdOpenChannel(data)
	case "delete":
		user.cmdDeleteMessage(data)
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

	channel, ok := data.Path("channel").Data().(string)
	if !ok {
		return
	}

	err := database.ClearChannel(user.GetEmail(), network, channel)
	if err != nil {
		fmt.Printf("<%s> Failed to clear history of %s@%s: %s", user.GetEmail(), channel, network, err)
		return
	}
	user.NewMessages <- mauircdi.Message{Type: "clear", Object: mauircdi.ClearHistory{Channel: channel, Network: network}}
}

func (user *userImpl) cmdCloseChannel(data *gabs.Container) {
	net, ok := data.Path("network").Data().(string)
	if !ok {
		return
	}

	network := user.GetNetwork(net)
	if network == nil {
		return
	}

	channel, ok := data.Path("channel").Data().(string)
	if !ok {
		return
	}

	network.GetActiveChannels().Remove(channel)
}

func (user *userImpl) cmdOpenChannel(data *gabs.Container) {
	net, ok := data.Path("network").Data().(string)
	if !ok {
		return
	}

	network := user.GetNetwork(net)
	if network == nil {
		return
	}

	channel, ok := data.Path("channel").Data().(string)
	if !ok {
		return
	}

	network.GetActiveChannels().Put(&chanDataImpl{Network: network.GetName(), Name: channel})
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

func (user *userImpl) cmdKick(data *gabs.Container) {
	network, ok := data.Path("network").Data().(string)
	if !ok {
		return
	}

	net := user.GetNetwork(network)
	if net == nil {
		return
	}

	channel, okChan := data.Path("channel").Data().(string)
	usr, okUser := data.Path("user").Data().(string)
	message, okMsg := data.Path("message").Data().(string)
	if !okChan || !okUser || !okMsg {
		return
	}

	if len(channel) > 0 && len(usr) > 0 && len(message) > 0 {
		net.SendRaw("KICK %s %s :%s", channel, usr, message)
	}
}

func (user *userImpl) cmdMode(data *gabs.Container) {
	network, ok := data.Path("network").Data().(string)
	if !ok {
		return
	}

	net := user.GetNetwork(network)
	if net == nil {
		return
	}

	channel, okChan := data.Path("channel").Data().(string)
	message, okMsg := data.Path("message").Data().(string)
	if !okChan || !okMsg {
		return
	}

	if len(channel) > 0 && len(message) > 0 {
		net.SendRaw("MODE %s %s", channel, message)
	}
}
