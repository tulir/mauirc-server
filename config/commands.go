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

// Package config contains configurations
package config

import (
	msg "github.com/sorcix/irc"
	"maunium.net/go/mauirc-common/messages"
	"maunium.net/go/mauirc-server/database"
)

// HandleCommand handles mauIRC commands from clients
func (user *userImpl) HandleCommand(data messages.Container) {
	switch data.Type {
	case messages.MsgRaw:
		user.rawMessage(messages.ParseRawMessage(data.Object))
	case messages.MsgMessage:
		user.cmdMessage(messages.ParseMessage(data.Object))
	case messages.MsgKick:
		user.cmdKick(messages.ParseKick(data.Object))
	case messages.MsgMode:
		user.cmdMode(messages.ParseMode(data.Object))
	case messages.MsgClear:
		user.cmdClearHistory(messages.ParseClearHistory(data.Object))
	case messages.MsgClose:
		user.cmdCloseChannel(messages.ParseOpenCloseChannel(data.Object))
	case messages.MsgOpen:
		user.cmdOpenChannel(messages.ParseOpenCloseChannel(data.Object))
	case messages.MsgDelete:
		user.cmdDeleteMessage(messages.ParseDeleteMessage(data.Object))
	}
}

func (user *userImpl) rawMessage(data messages.RawMessage) {
	if len(data.Network) == 0 || len(data.Message) == 0 {
		return
	}

	net := user.GetNetwork(data.Network)
	if net == nil {
		return
	}

	net.Tunnel().Send(msg.ParseMessage(data.Message))
}

func (user *userImpl) cmdDeleteMessage(data messages.DeleteMessage) {
	id := int64(data)
	err := database.DeleteMessage(user.Email, id)
	if err != nil {
		log.Warnf("<%s> Failed to delete message #%d: %s\n", user.Email, id, err)
		return
	}

	user.NewMessages <- messages.Container{Type: messages.MsgDelete, Object: data}
}

func (user *userImpl) cmdClearHistory(data messages.ClearHistory) {
	if len(data.Network) == 0 || len(data.Channel) == 0 {
		return
	}

	err := database.ClearChannel(user.GetEmail(), data.Network, data.Channel)
	if err != nil {
		log.Warnf("<%s> Failed to clear history of %s@%s: %s", user.GetEmail(), data.Network, data.Channel, err)
		return
	}

	user.NewMessages <- messages.Container{Type: messages.MsgClear, Object: data}
}

func (user *userImpl) cmdCloseChannel(data messages.OpenCloseChannel) {
	if len(data.Network) == 0 || len(data.Channel) == 0 {
		return
	}

	network := user.GetNetwork(data.Network)
	if network == nil {
		return
	}

	network.GetActiveChannels().Remove(data.Channel)
}

func (user *userImpl) cmdOpenChannel(data messages.OpenCloseChannel) {
	if len(data.Network) == 0 || len(data.Channel) == 0 {
		return
	}

	network := user.GetNetwork(data.Network)
	if network == nil {
		return
	}

	network.GetActiveChannels().Put(&chanDataImpl{Network: network.GetName(), Name: data.Channel})
}

func (user *userImpl) cmdMessage(data messages.Message) {
	if len(data.Network) == 0 || len(data.Channel) == 0 || len(data.Command) == 0 || len(data.Message) == 0 {
		return
	}

	net := user.GetNetwork(data.Network)
	if net == nil {
		return
	}

	if len(data.Channel) == 0 || len(data.Command) == 0 || len(data.Message) == 0 {
		return
	}
	net.SendMessage(data.Channel, data.Command, data.Message)
}

func (user *userImpl) cmdKick(data messages.Kick) {
	if len(data.Network) == 0 || len(data.Channel) == 0 || len(data.User) == 0 {
		return
	} else if len(data.Message) == 0 {
		data.Message = "Bye bye"
	}

	net := user.GetNetwork(data.Network)
	if net == nil {
		return
	}

	net.Tunnel().Kick(data.Channel, data.User, data.Message)
}

func (user *userImpl) cmdMode(data messages.Mode) {
	if len(data.Network) == 0 || len(data.Channel) == 0 || len(data.Message) == 0 {
		return
	}

	net := user.GetNetwork(data.Network)
	if net == nil {
		return
	}

	net.Tunnel().Mode(data.Channel, data.Message, data.Args)
}
