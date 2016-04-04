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
	"maunium.net/go/mauircd/database"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ChanList contains a list of channels
type ChanList struct {
	Network string   `json:"network"`
	List    []string `json:"list"`
}

func (net *Network) joinpart(user, channel string, part bool) {
	if user == net.Nick {
		net.joinpartMe(channel, part)
	} else {
		net.joinpartOther(user, channel, part)
	}
}

func (net *Network) mode(evt *irc.Event) {
	if evt.Arguments[0][0] == '#' {
		net.IRC.SendRawf("NAMES %s", evt.Arguments[0])
	}
	// TODO Send message to user about mode changes
}

func (net *Network) nick(evt *irc.Event) {
	if evt.Nick == net.Nick {
		net.Owner.NewMessages <- MauMessage{Type: "nickchange", Object: NickChange{Network: net.Name, Nick: evt.Message()}}
		net.Nick = evt.Message()
	}
	for _, ci := range net.ChannelInfo {
		if b, i := ci.UserList.Contains(evt.Nick); b {
			ci.UserList[i] = evt.Message()
			sort.Sort(ci.UserList)

			net.ReceiveMessage(ci.Name, evt.Nick, "nick", evt.Message())
			net.Owner.NewMessages <- MauMessage{Type: "chandata", Object: ci}
		}
	}
}

func (net *Network) userlist(evt *irc.Event) {
	ci := net.ChannelInfo[evt.Arguments[2]]
	if ci != nil {
		users := strings.Split(evt.Message(), " ")
		if len(users[len(users)-1]) == 0 {
			users = users[:len(users)-1]
		}

		if ci.ReceivingUserList {
			ci.UserList.Merge(users)
		} else {
			ci.UserList = UserList(users)
			ci.ReceivingUserList = true
		}
	}
}

func (net *Network) userlistend(evt *irc.Event) {
	ci := net.ChannelInfo[evt.Arguments[1]]
	if ci != nil {
		ci.ReceivingUserList = false
		sort.Sort(ci.UserList)
		net.Owner.NewMessages <- MauMessage{Type: "chandata", Object: ci}
	}
}

func (net *Network) chanlist(evt *irc.Event) {
	net.ChannelList = append(net.ChannelList, evt.Arguments[1])
}

func (net *Network) chanlistend(evt *irc.Event) {
	net.Owner.NewMessages <- MauMessage{Type: "chanlist", Object: ChanList{Network: net.Name, List: net.ChannelList}}
}

func (net *Network) topic(evt *irc.Event) {
	ci := net.ChannelInfo[evt.Arguments[0]]
	if ci != nil {
		ci.Topic = evt.Message()
		ci.TopicSetBy = evt.Nick
		ci.TopicSetAt = time.Now().Unix()
		net.ReceiveMessage(ci.Name, evt.Nick, "topic", evt.Message())
		net.Owner.NewMessages <- MauMessage{Type: "chandata", Object: ci}
	}
}

func (net *Network) topicresp(evt *irc.Event) {
	ci := net.ChannelInfo[evt.Arguments[1]]
	if ci != nil {
		ci.Topic = evt.Message()
		net.Owner.NewMessages <- MauMessage{Type: "chandata", Object: ci}
	}
}

func (net *Network) noperms(evt *irc.Event) {
	net.Owner.NewMessages <- MauMessage{
		Type: "message",
		Object: database.Message{
			ID:        -1,
			Network:   net.Name,
			Channel:   evt.Arguments[1],
			Timestamp: time.Now().Unix(),
			Sender:    "[" + net.Name + "]",
			Command:   "privmsg",
			Message:   evt.Message(),
			OwnMsg:    false,
		},
	}
}

func (net *Network) topicset(evt *irc.Event) {
	ci := net.ChannelInfo[evt.Arguments[1]]
	if ci != nil {
		ci.TopicSetBy = evt.Arguments[2]
		setAt, err := strconv.ParseInt(evt.Arguments[3], 10, 64)
		if err != nil {
			ci.TopicSetAt = setAt
		}
		net.Owner.NewMessages <- MauMessage{Type: "chandata", Object: ci}
	}
}

func (net *Network) quit(evt *irc.Event) {
	for _, ci := range net.ChannelInfo {
		if b, i := ci.UserList.Contains(evt.Nick); b {
			ci.UserList[i] = ci.UserList[len(ci.UserList)-1]
			ci.UserList = ci.UserList[:len(ci.UserList)-1]
			sort.Sort(ci.UserList)

			net.ReceiveMessage(ci.Name, evt.Nick, "quit", evt.Message())
			net.Owner.NewMessages <- MauMessage{Type: "chandata", Object: ci}
		}
	}
}

func (net *Network) join(evt *irc.Event) {
	net.ReceiveMessage(evt.Arguments[0], evt.Nick, "join", evt.Message())
	net.joinpart(evt.Nick, evt.Arguments[0], false)
}

func (net *Network) part(evt *irc.Event) {
	net.ReceiveMessage(evt.Arguments[0], evt.Nick, "part", evt.Message())
	net.joinpart(evt.Nick, evt.Arguments[0], true)
}

func (net *Network) privmsg(evt *irc.Event) {
	net.ReceiveMessage(evt.Arguments[0], evt.Nick, "privmsg", evt.Message())
}

func (net *Network) action(evt *irc.Event) {
	net.ReceiveMessage(evt.Arguments[0], evt.Nick, "action", evt.Message())
}
