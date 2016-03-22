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

	"github.com/thoj/go-ircevent"
	"maunium.net/go/mauircd/database"
	"maunium.net/go/mauircd/util"
	"strings"
	"time"
)

// Open an IRC connection
func (net *Network) Open(user *User) {
	i := irc.IRC(user.Nick, user.User)

	i.UseTLS = net.SSL
	i.QuitMessage = "mauIRCd shutting down..."
	if len(net.Password) > 0 {
		i.Password = net.Password
	}
	err := i.Connect(fmt.Sprintf("%s:%d", net.IP, net.Port))
	if err != nil {
		panic(err)
	}

	net.IRC = i
	net.Owner = user
	net.Nick = user.Nick
	net.NewMessages = make(chan database.Message, 256)

	i.AddCallback("PRIVMSG", net.privmsg)
	i.AddCallback("NOTICE", net.privmsg)
	i.AddCallback("CPRIVMSG", net.privmsg)
	i.AddCallback("CNOTICE", net.privmsg)
	i.AddCallback("CTCP_ACTION", net.action)
	i.AddCallback("JOIN", net.join)
	i.AddCallback("PART", net.part)
	i.AddCallback("001", func(evt *irc.Event) {
		for _, channel := range net.Channels {
			i.Join(channel)
		}
	})

	i.AddCallback("NICK", func(evt *irc.Event) {
		if evt.Nick == net.Nick {
			net.Nick = evt.Message()
		}
	})

	i.AddCallback("DISCONNECTED", func(event *irc.Event) {
		fmt.Printf("Disconnected from %s:%d\n", net.IP, net.Port)
	})
}

func (net *Network) joinpart(channel string, part bool) {
	for i, ch := range net.Channels {
		if ch == channel {
			if part {
				net.Channels[i] = net.Channels[len(net.Channels)-1]
				net.Channels = net.Channels[:len(net.Channels)-1]
				database.ClearChannel(net.Owner.Email, net.Name, ch)
			}
			return
		}
	}
	net.Channels = append(net.Channels, channel)
}

func (net *Network) message(channel, sender, command, message string) {
	if channel == "AUTH" {
		return
	}

	for _, s := range net.Scripts {
		channel, sender, command, message = s.Run(channel, sender, command, message)
	}

	if channel == net.Nick {
		channel = sender
	}

	if len(channel) == 0 || len(command) == 0 {
		return
	}

	msg := database.Message{Network: net.Name, Channel: channel, Timestamp: time.Now().Unix(), Sender: sender, Command: command, Message: message}
	net.NewMessages <- msg
	database.Insert(net.Owner.Email, msg)

	if sender == net.Nick && command == "join" {
		net.joinpart(channel, false)
	} else if sender == net.Nick && command == "part" {
		net.joinpart(channel, true)
	}
}

// SendMessage sends the given message to the given channel
func (net *Network) SendMessage(channel, command, message string) {
	splitted := util.Split(message)
	if len(splitted) > 1 {
		for _, piece := range splitted {
			net.SendMessage(channel, command, piece)
		}
		return
	}

	sender := net.Nick
	cancelled := false
	for _, s := range net.Scripts {
		channel, sender, command, message, cancelled = s.Run(channel, sender, command, message, cancelled)
	}

	if !strings.HasPrefix(channel, "*") {
		switch command {
		case "privmsg":
			net.IRC.Privmsg(channel, message)
		case "action":
			net.IRC.Action(channel, message)
		case "join":
			net.IRC.Join(channel)
			return
		case "part":
			net.IRC.Part(channel)
			return
		}
	} else if channel == "*mauirc" && command == "privmsg" {
		handleCommand(sender, command, args)
	}

	msg := database.Message{Network: net.Name, Channel: channel, Timestamp: time.Now().Unix(), Sender: sender, Command: command, Message: message}
	net.NewMessages <- msg
	database.Insert(net.Owner.Email, msg)
}

func (net *Network) handleCommand(sender, msg string) {
	split := strings.SplitN(msg, " ", 2)
	command = strings.ToLower(split[0])
	args = strings.Split(split[1], " ")

	switch command {
	case "clearbuffer":
		if len(args) > 0 {
			database.ClearChannel(net.Owner.Email, net.Name, args[0])
			net.message("*mauirc", "mauIRCd", "privmsg", "Successfully cleared buffer of "+args[0]+" on "+net.Name)
		}
	}
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
