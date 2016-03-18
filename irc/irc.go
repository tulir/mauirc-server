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

// Package irc contains the IRC client
package irc

import (
	"fmt"

	goirc "github.com/thoj/go-ircevent"
	"maunium.net/go/mauircd/database"
	"maunium.net/go/mauircd/plugin"
)

// Network is a mauircd network connection
type Network struct {
	IRC     *goirc.Connection
	Owner   string
	Name    string
	Nick    string
	Scripts []plugin.Script
}

// Create an IRC connection
func Create(name, nick, user, password, ip string, port int, ssl bool) *Network {
	i := goirc.IRC(nick, user)

	i.UseTLS = ssl
	if len(password) > 0 {
		i.Password = password
	}
	i.Connect(fmt.Sprintf("%s:%d", ip, port))
	mauirc := &Network{IRC: i, Owner: user, Name: name}

	i.AddCallback("PRIVMSG", mauirc.privmsg)
	i.AddCallback("CTCP_ACTION", mauirc.action)

	return mauirc
}

func (net *Network) message(channel, sender, command, message string) {
	for _, s := range net.Scripts {
		channel, sender, command, message = s.Run(channel, sender, command, message)
	}

	database.Insert(net.Owner, net.Name, channel, sender, command, message)
}

func (net *Network) sendMessage(channel, message string) {
	command := "privmsg"
	sender := net.Nick
	for _, s := range net.Scripts {
		channel, sender, command, message = s.Run(channel, sender, command, message)
	}

	net.IRC.Privmsg(channel, message)
	database.Insert(net.Owner, net.Name, channel, sender, command, message)
}

func (net *Network) privmsg(evt *goirc.Event) {
	net.message(evt.Arguments[0], evt.Nick, "privmsg", evt.Message())
}

func (net *Network) action(evt *goirc.Event) {
	net.message(evt.Arguments[0], evt.Nick, "action", evt.Message())
}
