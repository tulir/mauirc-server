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

// User is a single mauIRCd user
type User struct {
	Networks      []*Network       `json:"networks"`
	Email         string           `json:"email"`
	Password      string           `json:"password"`
	User          string           `json:"user"`
	Nick          string           `json:"nick"`
	Realname      string           `json:"realname"`
	AuthTokens    []AuthToken      `json:"authtokens,omitempty"`
	NewMessages   chan interface{} `json:"-"`
	GlobalScripts []plugin.Script  `json:"-"`
}

// AuthToken is a simple wrapper for an auth token string and a timestamp
type AuthToken struct {
	Token string
	Time  int64
}

// UserList is a wrapper for sorting user lists
type UserList []string

func (s UserList) Len() int {
	return len(s)
}
func (s UserList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s UserList) levelOf(r byte) int {
	switch r {
	case '~':
		return 5
	case '&':
		return 4
	case '@':
		return 3
	case '%':
		return 2
	case '+':
		return 1
	default:
		return 0
	}
}

func (s UserList) Less(i, j int) bool {
	levelI := s.levelOf(s[i][0])
	levelJ := s.levelOf(s[j][0])
	if levelI < levelJ {
		return true
	} else if levelI > levelJ {
		return false
	} else {
		return s[i] < s[j]
	}
}

// ChannelData contains information about a channeö
type ChannelData struct {
	UserList   UserList `json:"user-list"`
	Topic      string   `json:"topic"`
	TopicSetBy string   `json:"topic-set-by"`
	TopicSetAt int64    `json:"topic-set-at"`
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
