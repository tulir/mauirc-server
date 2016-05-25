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

// Message is a basic wrapper for a type string and the actual message object
type Message struct {
	Type   string      `json:"type"`
	Object interface{} `json:"object"`
}

// RawMessage is a raw IRC message
type RawMessage struct {
	Network string `json:"network"`
	Message string `json:"message"`
}

// NickChange is the message for IRC nick changes
type NickChange struct {
	Network string `json:"network"`
	Nick    string `json:"nick"`
}

// NetData contains basic network data
type NetData struct {
	Name      string `json:"name"`
	Connected bool   `json:"connected"`
}

// ChanList contains a channel list and network name
type ChanList struct {
	Network string   `json:"network"`
	List    []string `json:"list"`
}

// Invite an user to a channel
type Invite struct {
	Network string `json:"network"`
	Channel string `json:"channel"`
	Sender  string `json:"sender"`
}

// CommandResponse is a response to an user-sent internal command
type CommandResponse struct {
	Success       bool   `json:"success"`
	SimpleMessage string `json:"simple-message"`
	Message       string `json:"message"`
}

// ClearHistory tells the client to clear the specific channel
type ClearHistory struct {
	Network string `json:"network"`
	Channel string `json:"channel"`
}
