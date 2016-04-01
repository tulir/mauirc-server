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
	"maunium.net/go/mauircd/preview"
	"maunium.net/go/mauircd/util"
	"sort"
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

	net.IRC = i
	net.Owner = user
	net.Nick = user.Nick

	i.AddCallback("PRIVMSG", net.privmsg)
	i.AddCallback("NOTICE", net.privmsg)
	i.AddCallback("CPRIVMSG", net.privmsg)
	i.AddCallback("CNOTICE", net.privmsg)
	i.AddCallback("CTCP_ACTION", net.action)
	i.AddCallback("JOIN", net.join)
	i.AddCallback("PART", net.part)
	i.AddCallback("MODE", net.mode)
	i.AddCallback("353", net.userlist)
	i.AddCallback("366", net.userlistend)
	i.AddCallback("322", net.chanlist)
	i.AddCallback("323", net.chanlistend)
	i.AddCallback("332", net.topic)
	i.AddCallback("333", net.topicset)
	i.AddCallback("NICK", net.nick)
	i.AddCallback("QUIT", net.quit)

	i.AddCallback("001", func(evt *irc.Event) {
		i.SendRaw("LIST") // TODO update channel list properly
		for _, channel := range net.Channels {
			i.Join(channel)
		}
	})

	i.AddCallback("DISCONNECTED", func(event *irc.Event) {
		fmt.Printf("Disconnected from %s:%d\n", net.IP, net.Port)
	})

	err := i.Connect(fmt.Sprintf("%s:%d", net.IP, net.Port))
	if err != nil {
		panic(err)
	}
}

// ReceiveMessage stores the message and sends it to the client
func (net *Network) ReceiveMessage(channel, sender, command, message string) {
	msg := database.Message{Network: net.Name, Channel: channel, Timestamp: time.Now().Unix(), Sender: sender, Command: command, Message: message}
	cancelled := false
	if msg.Channel == "AUTH" || msg.Channel == "*" {
		return
	} else if msg.Channel == net.Nick {
		msg.Channel = msg.Sender
	}

	msg, cancelled = net.RunScripts(msg, cancelled, true)
	if cancelled {
		return
	}

	if len(msg.Channel) == 0 || len(msg.Command) == 0 {
		return
	}

	net.InsertAndSend(msg)
}

// SendMessage sends the given message to the given channel
func (net *Network) SendMessage(channel, command, message string) {
	msg := database.Message{Network: net.Name, Channel: channel, Timestamp: time.Now().Unix(), Sender: net.Nick, Command: command, Message: message}
	cancelled := false

	msg, cancelled = net.RunScripts(msg, cancelled, false)
	if cancelled {
		return
	}

	if splitted := util.Split(msg.Message); len(splitted) > 1 {
		for _, piece := range splitted {
			net.SendMessage(msg.Channel, msg.Command, piece)
		}
		return
	}

	if net.sendToIRC(msg) {
		net.InsertAndSend(msg)
	}
}

// SendRaw sends the raw message to IRC.
func (net *Network) SendRaw(message string) {
	net.IRC.SendRaw(message)
}

func (net *Network) sendToIRC(msg database.Message) bool {
	if !strings.HasPrefix(msg.Channel, "*") {
		switch msg.Command {
		case "privmsg":
			net.IRC.Privmsg(msg.Channel, msg.Message)
		case "action":
			net.IRC.Action(msg.Channel, msg.Message)
		case "join":
			net.IRC.Join(msg.Channel)
			return false
		case "part":
			net.IRC.Part(msg.Channel)
			return false
		}
	}
	return true
}

// SwitchNetwork sends the given message to another network
func (net *Network) SwitchNetwork(msg database.Message, receiving bool) bool {
	newNet := net.Owner.GetNetwork(msg.Network)
	if newNet != nil {
		if receiving {
			newNet.ReceiveMessage(msg.Channel, msg.Sender, msg.Command, msg.Message)
		} else {
			newNet.SendMessage(msg.Channel, msg.Command, msg.Message)
		}
		return true
	}
	return false
}

// InsertAndSend inserts the given message into the database and sends it to the client
func (net *Network) InsertAndSend(msg database.Message) {
	msg.Preview, _ = preview.GetPreview(msg.Message)
	msg.ID = database.Insert(net.Owner.Email, msg)
	net.Owner.NewMessages <- MauMessage{Type: "message", Object: msg}
}

// Close the IRC connection.
func (net *Network) Close() {
	if net.IRC.Connected() {
		net.IRC.Quit()
	}
}

func (net *Network) joinpartMe(channel string, part bool) {
	for i, ch := range net.Channels {
		if ch == channel {
			if part {
				net.Channels[i] = net.Channels[len(net.Channels)-1]
				net.Channels = net.Channels[:len(net.Channels)-1]
				database.ClearChannel(net.Owner.Email, net.Name, ch)
				delete(net.ChannelInfo, channel)
			} else {
				if net.ChannelInfo[channel] == nil {
					net.ChannelInfo[channel] = &ChannelData{Name: channel, Network: net.Name}
				}
				return
			}
		}
	}
	if !part {
		net.ChannelInfo[channel] = &ChannelData{Name: channel, Network: net.Name}
		net.Channels = append(net.Channels, channel)
	}
}

func (net *Network) joinpartOther(user, channel string, part bool) {
	ci := net.ChannelInfo[channel]
	if ci == nil {
		return
	}

	contains, i := ci.UserList.Contains(user)
	if contains {
		if part {
			ci.UserList[i] = ci.UserList[len(ci.UserList)-1]
			ci.UserList = ci.UserList[:len(ci.UserList)-1]
		} else {
			return
		}
	} else if !part {
		ci.UserList = append(ci.UserList, user)
	}
	sort.Sort(ci.UserList)
	net.Owner.NewMessages <- MauMessage{Type: "chandata", Object: ci}
}
