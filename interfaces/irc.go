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

// Package mauircdi contains interfaces
package mauircdi

import (
	"maunium.net/go/mauircd/database"
)

// Network is a single IRC network
type Network interface {
	Open()
	ReceiveMessage(channel, sender, command, message string)
	SendMessage(channel, sender, message string)
	SwitchMessageNetwork(msg database.Message, receiving bool) bool
	InsertAndSend(msg database.Message)
	SendRaw(msg string, args ...interface{})
	Close()

	GetOwner() User
	GetName() string
	GetNick() string

	SaveScripts(path string) error
	LoadScripts(path string) error

	GetActiveChannels() ChannelDataList
	GetAllChannels() []string

	GetScripts() []Script
	AddScript(s Script)
}

// ChannelDataList contains a list of channel data objects
type ChannelDataList interface {
	Get(channel string) (ChannelData, bool)
	Put(data ChannelData)
	ForEach(do func(ChannelData))
}

// ChannelData has basic channel data (topic, user list, etc)
type ChannelData interface {
	GetUsers() []string
	GetName() string
	GetTopic() string
	GetNetwork() string
}

// Script wraps the name and code of a script
type Script interface {
	GetName() string
	GetScript() string
	Run(net Network, msg database.Message, cancelled bool) (database.Message, bool)
}
