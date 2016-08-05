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

// Package interfaces contains interfaces
package interfaces

import (
	"maunium.net/go/libmauirc"
	"maunium.net/go/mauirc-common/messages"
	"maunium.net/go/mauirc-server/database"
	"maunium.net/go/mauirc-server/util/userlist"
)

// Network is a single IRC network
type Network interface {
	Save()
	Open()
	ReceiveMessage(channel, sender, command, message string)
	SendMessage(channel, command, message string)
	SwitchMessageNetwork(msg database.Message, receiving bool) bool
	InsertAndSend(msg database.Message)
	Tunnel() libmauirc.Tunnel

	Connect() error
	Disconnect()
	ForceDisconnect()
	IsConnected() bool

	GetOwner() User
	GetName() string
	GetNick() string
	GetNetData() messages.NetData

	SetName(name string)
	SetNick(nick string)
	SetRealname(realname string)
	SetUser(user string)
	SetIP(ip string)
	SetPort(port uint16)
	SetSSL(ssl bool)

	SaveScripts(path string) error
	LoadScripts(path string) error

	GetActiveChannels() ChannelDataList
	GetAllChannels() []string

	GetScripts() []Script
	AddScript(s Script) bool
	RemoveScript(name string) bool
}

// ChannelDataList contains a list of channel data objects
type ChannelDataList interface {
	Get(channel string) (ChannelData, bool)
	Put(data ChannelData)
	Remove(channel string)
	Has(channel string) bool
	ForEach(do func(ChannelData))
}

// ChannelData has basic channel data (topic, user list, etc)
type ChannelData interface {
	GetUsers() []string
	GetName() string
	GetTopic() string
	GetNetwork() string
	Modes() ModeList
}

// ModeList is a list of Modes
type ModeList []Mode

// Mode contains a channel mode rune and the target.
type Mode struct {
	Mode   rune   `json:"mode"`
	Target string `json:"target"`
}

// HasMode checks if the given mode list contains the given rune with the given target.
func (ml ModeList) HasMode(r rune, target string) bool {
	for _, mode := range ml {
		if mode.Mode == r && mode.Target == target {
			return true
		}
	}
	return false
}

// AddMode adds the given mode with the given target
func (ml ModeList) AddMode(r rune, target string) ModeList {
	if !ml.HasMode(r, target) {
		ml = append(ml, Mode{Mode: r, Target: target})
	}
	return ml
}

// RemoveMode removes the given mode with the given target
func (ml ModeList) RemoveMode(r rune, target string) ModeList {
	for i, mode := range ml {
		if mode.Mode == r && mode.Target == target {
			ml[i] = ml[len(ml)-1]
			ml = ml[:len(ml)-1]
		}
	}
	return ml
}

// PrefixOf gets the prefix of the given user
func (ml ModeList) PrefixOf(user string) rune {
	level := 0
	for _, mode := range ml {
		if mode.Target == user {
			lvl := ml.levelOfMode(mode.Mode)
			if lvl > level {
				level = lvl
			}
		}
	}
	return userlist.PrefixOf(level)
}

func (ml ModeList) levelOfMode(r rune) int {
	switch r {
	case 'q':
		return 5
	case 'a':
		return 4
	case 'o':
		return 3
	case 'h':
		return 2
	case 'v':
		return 1
	default:
		return 0
	}
}
