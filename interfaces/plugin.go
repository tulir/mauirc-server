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
	"maunium.net/go/mauirc-server/database"
)

// Script wraps the name and code of a script
type Script interface {
	GetName() string
	GetScript() string
	Run(evt *Event)
}

// Event is a plugin event
type Event struct {
	Network   Network
	Message   database.Message
	Cancelled bool
}
