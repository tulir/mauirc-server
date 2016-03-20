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

	"github.com/thoj/go-ircevent"
	"maunium.net/go/mauircd/database"
	"maunium.net/go/mauircd/plugin"
	"time"
)

// TmpNet ...
var TmpNet *Network

// Network is a mauircd network connection
type Network struct {
	IRC         *irc.Connection
	Owner       string
	Name        string
	Nick        string
	Scripts     []plugin.Script
	NewMessages chan database.Message
}

// Create an IRC connection
func Create(name, nick, user, email, password, ip string, port int, ssl bool) *Network {
	i := irc.IRC(nick, user)

	i.UseTLS = ssl
	i.QuitMessage = "mauIRCd shutting down..."
	if len(password) > 0 {
		i.Password = password
	}
	err := i.Connect(fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		panic(err)
	}

	mauirc := &Network{IRC: i, Owner: email, Name: name, Nick: nick, NewMessages: make(chan database.Message, 256)}

	i.AddCallback("PRIVMSG", mauirc.privmsg)
	i.AddCallback("CTCP_ACTION", mauirc.action)
	i.AddCallback("JOIN", mauirc.join)
	i.AddCallback("PART", mauirc.part)
	i.AddCallback("001", func(evt *irc.Event) {
		i.Join("#mau")
	})

	i.AddCallback("NICK", func(evt *irc.Event) {
		if evt.Nick == mauirc.Nick {
			mauirc.Nick = evt.Message()
		}
	})

	i.AddCallback("DISCONNECTED", func(event *irc.Event) {
		fmt.Printf("Disconnected from %s:%d\n", ip, port)
	})

	return mauirc
}

func (net *Network) message(channel, sender, command, message string) {
	for _, s := range net.Scripts {
		channel, sender, command, message = s.Run(channel, sender, command, message)
	}

	if len(channel) == 0 || len(command) == 0 {
		return
	}

	msg := database.Message{Network: net.Name, Channel: channel, Timestamp: time.Now().Unix(), Sender: sender, Command: command, Message: message}
	net.NewMessages <- msg
	database.Insert(net.Owner, msg)
}

// SendMessage sends the given message to the given channel
func (net *Network) SendMessage(channel, command, message string) {
	splitted := split(message)
	if splitted != nil && len(splitted) > 1 {
		for _, piece := range splitted {
			net.SendMessage(channel, command, piece)
		}
		return
	}

	sender := net.Nick
	for _, s := range net.Scripts {
		channel, sender, command, message = s.Run(channel, sender, command, message)
	}

	if command == "privmsg" {
		net.IRC.Privmsg(channel, message)
	} else if command == "action" {
		net.IRC.Action(channel, message)
	}
	msg := database.Message{Network: net.Name, Channel: channel, Timestamp: time.Now().Unix(), Sender: sender, Command: command, Message: message}
	net.NewMessages <- msg
	database.Insert(net.Owner, msg)
}

// Close the IRC connection.
func (net *Network) Close() {
	net.IRC.Quit()
}

func (net *Network) join(evt *irc.Event) {
	go net.message(evt.Arguments[0], evt.Nick, "join", evt.Message())
}

func (net *Network) part(evt *irc.Event) {
	go net.message(evt.Arguments[0], evt.Nick, "part", evt.Message())
}

func (net *Network) privmsg(evt *irc.Event) {
	go net.message(evt.Arguments[0], evt.Nick, "privmsg", evt.Message())
}

func (net *Network) action(evt *irc.Event) {
	go net.message(evt.Arguments[0], evt.Nick, "action", evt.Message())
}
