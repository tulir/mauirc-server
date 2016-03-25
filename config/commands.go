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
	"maunium.net/go/mauircd/database"
	"maunium.net/go/mauircd/plugin"
	"strconv"
	"strings"
)

// TODO command handlers should not be network-specific
func (net *Network) handleCommand(sender, msg string) {
	split := strings.SplitN(msg, " ", 2)
	command := strings.ToLower(split[0])
	args := strings.Split(split[1], " ")

	switch command {
	case "clearbuffer":
		if len(args) > 0 {
			database.ClearChannel(net.Owner.Email, net.Name, args[0])
			net.ReceiveMessage("*mauirc", "mauIRCd", "privmsg", "Successfully cleared buffer of "+args[0]+" on "+net.Name)
		}
	case "deletemessage":
		if len(args) > 0 {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				net.ReceiveMessage("*mauirc", "mauIRCd", "privmsg", "Couldn't parse int from "+args[0])
				return
			}
			database.DeleteMessage(net.Owner.Email, int64(id))
		}
	case "importscript":
		if len(args) > 1 {
			args[0] = strings.ToLower(args[0])
			data, err := download(args[1])
			if err != nil {
				fmt.Println(err)
				net.ReceiveMessage("*mauirc", "mauIRCd", "privmsg", "Failed to download script from http://pastebin.com/raw/"+args[1])
				return
			}
			for i := 0; i < len(net.Scripts); i++ {
				if net.Scripts[i].Name == args[0] {
					net.Scripts[i].TheScript = data
					net.ReceiveMessage("*mauirc", "mauIRCd", "privmsg", "Successfully updated script with name "+args[0])
					return
				}
			}
			net.Scripts = append(net.Scripts, plugin.Script{TheScript: data, Name: args[0]})
			net.ReceiveMessage("*mauirc", "mauIRCd", "privmsg", "Successfully loaded script with name "+args[0])
		}
	}
}
