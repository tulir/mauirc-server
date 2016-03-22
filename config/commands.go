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
	"maunium.net/go/mauircd/database"
	"strings"
)

func (net *Network) handleCommand(sender, msg string) {
	split := strings.SplitN(msg, " ", 2)
	command := strings.ToLower(split[0])
	args := strings.Split(split[1], " ")

	switch command {
	case "clearbuffer":
		if len(args) > 0 {
			database.ClearChannel(net.Owner.Email, net.Name, args[0])
			net.message("*mauirc", "mauIRCd", "privmsg", "Successfully cleared buffer of "+args[0]+" on "+net.Name)
		}
	}
}
