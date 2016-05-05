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
	msg "github.com/sorcix/irc"
	"maunium.net/go/mauircd/database"
	"maunium.net/go/mauircd/interfaces"
	"strconv"
)

// HandleCommand handles mauIRC commands from clients
func (user *userImpl) HandleCommand(data map[string]string) {
	switch data["type"] {
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

func (user *userImpl) rawMessage(data map[string]string) {
	if len(data["network"]) == 0 || len(data["message"]) == 0 {
		return
	}

	net := user.GetNetwork(data["network"])
	if net == nil {
		return
	}

	net.Tunnel().Send(msg.ParseMessage(data["message"]))
}

func (user *userImpl) cmdDeleteMessage(data map[string]string) {
	if len(data["id"]) == 0 {
		return
	}

	id, err := strconv.ParseInt(data["id"], 10, 64)
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

func (user *userImpl) cmdClearHistory(data map[string]string) {
	if len(data["network"]) == 0 || len(data["channel"]) == 0 {
		return
	}

	err := database.ClearChannel(user.GetEmail(), data["network"], data["channel"])
	if err != nil {
		fmt.Printf("<%s> Failed to clear history of %s@%s: %s", user.GetEmail(), data["network"], data["channel"], err)
		return
	}

	user.NewMessages <- mauircdi.Message{Type: "clear", Object: mauircdi.ClearHistory{Channel: data["channel"], Network: data["network"]}}
}

func (user *userImpl) cmdCloseChannel(data map[string]string) {
	if len(data["network"]) == 0 || len(data["channel"]) == 0 {
		return
	}

	network := user.GetNetwork(data["network"])
	if network == nil {
		return
	}

	network.GetActiveChannels().Remove(data["channel"])
}

func (user *userImpl) cmdOpenChannel(data map[string]string) {
	if len(data["network"]) == 0 || len(data["channel"]) == 0 {
		return
	}

	network := user.GetNetwork(data["network"])
	if network == nil {
		return
	}

	network.GetActiveChannels().Put(&chanDataImpl{Network: network.GetName(), Name: data["channel"]})
}

func (user *userImpl) cmdMessage(data map[string]string) {
	if len(data["network"]) == 0 || len(data["channel"]) == 0 || len(data["command"]) == 0 || len(data["message"]) == 0 {
		return
	}

	net := user.GetNetwork(data["network"])
	if net == nil {
		return
	}

	if len(data["channel"]) == 0 || len(data["command"]) == 0 || len(data["message"]) == 0 {
		return
	}
	net.SendMessage(data["channel"], data["command"], data["message"])
}

func (user *userImpl) cmdKick(data map[string]string) {
	if len(data["network"]) == 0 || len(data["channel"]) == 0 || len(data["user"]) == 0 || len(data["message"]) == 0 {
		return
	}

	net := user.GetNetwork(data["network"])
	if net == nil {
		return
	}

	net.Tunnel().Kick(data["channel"], data["user"], data["message"])
}

func (user *userImpl) cmdMode(data map[string]string) {
	if len(data["network"]) == 0 || len(data["channel"]) == 0 || len(data["message"]) == 0 {
		return
	}

	net := user.GetNetwork(data["network"])
	if net == nil {
		return
	}

	net.Tunnel().Mode(data["channel"], data["message"], "")
}
