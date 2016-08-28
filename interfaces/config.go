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
	"maunium.net/go/mauirc-common/messages"
	"time"
)

// Configuration contains the main config
type Configuration interface {
	Load() error
	Connect()
	Save() error

	GetIDENTConfig() IdentConf

	GetSQLString() string
	GetPath() string

	GetMail() Mail

	GetUsers() UserList
	GetUser(name string) User
	PurgeUnverifiedUsers()
	CreateUser(email, password string) (User, string, time.Time)

	GetAddr() string
	GetExternalAddr() string
	TrustHeaders() bool

	GetCookieSecret() []byte
}

// Mail (er)
type Mail interface {
	Validate() error
	IsEnabled() bool
	Send(to, subject, template string, args map[string]interface{})
	LoadTemplates(path string)
}

// IdentConf tells the IDENT server what ip and port to bind to
type IdentConf struct {
	Enabled bool   `json:"enabled"`
	IP      string `json:"ip"`
	Port    int    `json:"port"`
}

// UserList is a list of users that can be looped through
type UserList interface {
	ForEach(func(user User))
}

// User contains the authentication and network data of an user
type User interface {
	Save()

	InitNetworks()
	GetNetworks() NetworkList
	GetNetwork(name string) Network
	DeleteNetwork(name string) bool
	AddNetwork(nw Network) bool
	CreateNetwork(name string, data []byte) (Network, bool)
	SendNetworkData(net Network)

	GetEmail() string
	GetNameFromEmail() string

	IsVerified() bool
	SetVerified()

	NewResetToken() (string, time.Time)
	CheckResetToken(token string) bool
	ClearResetToken()

	CheckPassword(password string) bool
	SetPassword(newPassword string) error

	NewAuthToken() string
	CheckAuthToken(token string) bool

	HandleCommand(data messages.Container)

	GetGlobalScripts() []Script
	AddGlobalScript(s Script) bool
	RemoveGlobalScript(name string) bool
	SaveGlobalScripts(path string) error

	GetMessageChan() chan messages.Container

	GetSettings() interface{}
	SetSettings(val interface{})
}

// NetworkList is a list of networks that can be looped through
type NetworkList interface {
	ForEach(func(net Network))
}
