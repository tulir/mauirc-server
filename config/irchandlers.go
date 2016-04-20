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
	"maunium.net/go/mauircd/interfaces"
	"maunium.net/go/mauircd/util/userlist"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (net *netImpl) mode(evt *irc.Event) {
	if evt.Arguments[0][0] == '#' {
		net.IRC.SendRawf("NAMES %s", evt.Arguments[0])
	}
	// TODO add proper MODE handling
	fmt.Println(evt.Arguments)
	fmt.Println(evt.User, evt.Source, evt.Nick, evt.Host, evt.Code)
	net.ReceiveMessage(evt.Arguments[0], evt.Nick, "mode", strings.Join(evt.Arguments[1:], " "))
}

func (net *netImpl) nick(evt *irc.Event) {
	if evt.Nick == net.Nick {
		net.Owner.NewMessages <- mauircdi.Message{Type: "nickchange", Object: mauircdi.NickChange{Network: net.Name, Nick: evt.Message()}}
		net.Nick = evt.Message()
	}
	for _, ci := range net.ChannelInfo {
		if b, i := ci.UserList.Contains(evt.Nick); b {
			ci.UserList[i] = evt.Message()
			sort.Sort(ci.UserList)

			net.ReceiveMessage(ci.Name, evt.Nick, "nick", evt.Message())
			net.Owner.NewMessages <- mauircdi.Message{Type: "chandata", Object: ci}
		}
	}
}

func (net *netImpl) userlist(evt *irc.Event) {
	ciInt, ok := net.ChannelInfo.Get(evt.Arguments[2])
	if !ok {
		return
	}

	ci, ok := ciInt.(*chanDataImpl)
	if !ok {
		return
	}

	users := strings.Split(evt.Message(), " ")
	if len(users[len(users)-1]) == 0 {
		users = users[:len(users)-1]
	}

	if ci.ReceivingUserList {
		ci.UserList = ci.UserList.Merge(users)
	} else {
		ci.UserList = userlist.List(users)
		ci.ReceivingUserList = true
	}
}

func (net *netImpl) userlistend(evt *irc.Event) {
	ciInt, ok := net.ChannelInfo.Get(evt.Arguments[1])
	if !ok {
		return
	}

	ci, ok := ciInt.(*chanDataImpl)
	if !ok {
		return
	}

	ci.ReceivingUserList = false
	sort.Sort(ci.UserList)
	net.Owner.NewMessages <- mauircdi.Message{Type: "chandata", Object: ci}
}

func (net *netImpl) chanlist(evt *irc.Event) {
	net.ChannelList = append(net.ChannelList, evt.Arguments[1])
}

func (net *netImpl) chanlistend(evt *irc.Event) {
	net.Owner.NewMessages <- mauircdi.Message{Type: "chanlist", Object: mauircdi.ChanList{Network: net.Name, List: net.ChannelList}}
}

func (net *netImpl) topic(evt *irc.Event) {
	ci := net.ChannelInfo[evt.Arguments[0]]
	if ci != nil {
		ci.Topic = evt.Message()
		ci.TopicSetBy = evt.Nick
		ci.TopicSetAt = time.Now().Unix()
		net.ReceiveMessage(ci.Name, evt.Nick, "topic", evt.Message())
		net.Owner.NewMessages <- mauircdi.Message{Type: "chandata", Object: ci}
	}
}

func (net *netImpl) topicresp(evt *irc.Event) {
	ci := net.ChannelInfo[evt.Arguments[1]]
	if ci != nil {
		ci.Topic = evt.Message()
		net.Owner.NewMessages <- mauircdi.Message{Type: "chandata", Object: ci}
	}
}

func (net *netImpl) noperms(evt *irc.Event) {
	net.Owner.NewMessages <- mauircdi.Message{
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

func (net *netImpl) topicset(evt *irc.Event) {
	ci := net.ChannelInfo[evt.Arguments[1]]
	if ci != nil {
		ci.TopicSetBy = evt.Arguments[2]
		setAt, err := strconv.ParseInt(evt.Arguments[3], 10, 64)
		if err != nil {
			ci.TopicSetAt = setAt
		}
		net.Owner.NewMessages <- mauircdi.Message{Type: "chandata", Object: ci}
	}
}

func (net *netImpl) quit(evt *irc.Event) {
	for _, ci := range net.ChannelInfo {
		if b, i := ci.UserList.Contains(evt.Nick); b {
			ci.UserList[i] = ci.UserList[len(ci.UserList)-1]
			ci.UserList = ci.UserList[:len(ci.UserList)-1]
			sort.Sort(ci.UserList)

			net.ReceiveMessage(ci.Name, evt.Nick, "quit", evt.Message())
			net.Owner.NewMessages <- mauircdi.Message{Type: "chandata", Object: ci}
		}
	}
}

func (net *netImpl) join(evt *irc.Event) {
	net.ReceiveMessage(evt.Arguments[0], evt.Nick, "join", evt.Message())
	net.joinpart(evt.Nick, evt.Arguments[0], false)
}

func (net *netImpl) part(evt *irc.Event) {
	net.ReceiveMessage(evt.Arguments[0], evt.Nick, "part", evt.Message())
	net.joinpart(evt.Nick, evt.Arguments[0], true)
}

func (net *netImpl) kick(evt *irc.Event) {
	net.ReceiveMessage(evt.Arguments[0], evt.Nick, "kick", evt.Arguments[1]+":"+evt.Message())
	net.joinpart(evt.Nick, evt.Arguments[0], true)
}

func (net *netImpl) privmsg(evt *irc.Event) {
	net.ReceiveMessage(evt.Arguments[0], evt.Nick, "privmsg", evt.Message())
}

func (net *netImpl) action(evt *irc.Event) {
	net.ReceiveMessage(evt.Arguments[0], evt.Nick, "action", evt.Message())
}

func (net *netImpl) connected(evt *irc.Event) {
	net.IRC.SendRaw("LIST")
	for channel := range net.ChannelInfo {
		if strings.HasPrefix(channel, "#") {
			net.IRC.Join(channel)
		}
	}
	net.GetOwner().GetMessageChan() <- mauircdi.Message{Type: "netdata", Object: mauircdi.NetData{Name: net.GetName(), Connected: true}}
}

func (net *netImpl) disconnected(event *irc.Event) {
	fmt.Printf("Disconnected from %s:%d\n", net.IP, net.Port)
	net.GetOwner().GetMessageChan() <- mauircdi.Message{Type: "netdata", Object: mauircdi.NetData{Name: net.GetName(), Connected: false}}
}

func (net *netImpl) joinpart(user, channel string, part bool) {
	if user == net.Nick {
		net.joinpartMe(channel, part)
	} else {
		net.joinpartOther(user, channel, part)
	}
}

func (net *netImpl) joinpartMe(channel string, part bool) {
	for ch := range net.ChannelInfo {
		if ch == channel {
			if part {
				net.ChannelInfo.Remove(ch)
			} else {
				if net.ChannelInfo[channel] == nil {
					net.ChannelInfo.Put(&chanDataImpl{Name: channel, Network: net.Name})
				}
				return
			}
		}
	}
	if !part {
		net.ChannelInfo.Put(&chanDataImpl{Name: channel, Network: net.Name})
	}
}

func (net *netImpl) joinpartOther(user, channel string, part bool) {
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
	net.Owner.NewMessages <- mauircdi.Message{Type: "chandata", Object: ci}
}
