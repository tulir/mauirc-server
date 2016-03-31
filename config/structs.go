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
	"github.com/thoj/go-ircevent"
	"maunium.net/go/mauircd/plugin"
)

// Configuration is the base configuration for mauIRCd
type Configuration struct {
	Path         string    `json:"-"`
	SQL          SQLConfig `json:"sql"`
	Users        []*User   `json:"users"`
	IP           string    `json:"ip"`
	Port         int       `json:"port"`
	Address      string    `json:"external-address"`
	CSecretB64   string    `json:"cookie-secret"`
	CookieSecret []byte    `json:"-"`
}

// SQLConfig contains sql connection info
type SQLConfig struct {
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
}

// MauMessage wraps a generic interface and a type string
type MauMessage struct {
	Type   string      `json:"type"`
	Object interface{} `json:"object"`
}

// User is a single mauIRCd user
type User struct {
	Networks      []*Network      `json:"networks"`
	Email         string          `json:"email"`
	Password      string          `json:"password"`
	User          string          `json:"user"`
	Nick          string          `json:"nick"`
	Realname      string          `json:"realname"`
	AuthTokens    []AuthToken     `json:"authtokens,omitempty"`
	NewMessages   chan MauMessage `json:"-"`
	GlobalScripts []plugin.Script `json:"-"`
}

// AuthToken is a simple wrapper for an auth token string and a timestamp
type AuthToken struct {
	Token string
	Time  int64
}

// ChannelData contains information about a channeö
type ChannelData struct {
	Network           string   `json:"network"`
	Name              string   `json:"name"`
	UserList          UserList `json:"userlist"`
	UserCount         int      `json:"usercount"`
	Topic             string   `json:"topic"`
	TopicSetBy        string   `json:"topicsetby"`
	TopicSetAt        int64    `json:"topicsetat"`
	ReceivingUserList bool     `json:"-"`
}

// NickChange tells the client about nick changes
type NickChange struct {
	Network string `json:"network"`
	Nick    string `json:"nick"`
}

// Network is a single IRC network owned by a single mauIRCd user
type Network struct {
	Name     string   `json:"name"`
	IP       string   `json:"ip"`
	Port     int      `json:"port"`
	Password string   `json:"password"`
	SSL      bool     `json:"ssl"`
	Channels []string `json:"channels"`

	Owner       *User                   `json:"-"`
	IRC         *irc.Connection         `json:"-"`
	Nick        string                  `json:"-"`
	Scripts     []plugin.Script         `json:"-"`
	ChannelInfo map[string]*ChannelData `json:"-"`
}
