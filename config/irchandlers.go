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
	msg "github.com/sorcix/irc"
	"maunium.net/go/mauircd/database"
	"maunium.net/go/mauircd/interfaces"
	"maunium.net/go/mauircd/util/userlist"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (net *netImpl) mode(evt *msg.Message) {
	if len(evt.Host) == 0 {
		return
	}
	if evt.Params[0][0] == '#' {
		ci := net.ChannelInfo.get(evt.Params[0])
		if ci == nil {
			net.ChannelInfo.Put(&chanDataImpl{Network: net.Name, Name: evt.Params[0]})
			ci = net.ChannelInfo.get(evt.Params[0])
		}

		var targets = evt.Params[2:]

		var add = true
		var ii = 0
		for _, r := range evt.Params[1] {
			if r == '-' {
				add = false
			} else if r == '+' {
				add = true
			} else {
				var target string
				if len(targets) > ii {
					target = targets[ii]
				} else {
					target = ""
				}

				if add {
					ci.ModeList = ci.ModeList.AddMode(r, target)
				} else {
					ci.ModeList = ci.ModeList.RemoveMode(r, target)
				}

				if len(target) > 0 {
					ci.UserList.SetPrefix(target, fmt.Sprintf("%c", ci.ModeList.PrefixOf(target)))
					sort.Sort(ci.UserList)
				}
				ii++
			}
		}

		net.Owner.NewMessages <- mauircdi.Message{Type: "chandata", Object: ci}
	}
	net.ReceiveMessage(evt.Params[0], evt.Name, "mode", strings.Join(evt.Params[1:], " "))
}

func (net *netImpl) nick(evt *msg.Message) {
	if evt.Name == net.IRC.GetNick() {
		net.Owner.NewMessages <- mauircdi.Message{Type: "nickchange", Object: mauircdi.NickChange{Network: net.Name, Nick: evt.Trailing}}
		net.Nick = evt.Trailing
	}
	for _, ci := range net.ChannelInfo {
		if b, i := ci.UserList.Contains(evt.Name); b {
			ci.UserList[i] = evt.Trailing
			sort.Sort(ci.UserList)

			net.ReceiveMessage(ci.Name, evt.Name, "nick", evt.Trailing)
			net.Owner.NewMessages <- mauircdi.Message{Type: "chandata", Object: ci}
		}
	}
}

func (net *netImpl) userlist(evt *msg.Message) {
	ci := net.ChannelInfo.get(evt.Params[2])
	if ci == nil {
		net.ChannelInfo.Put(&chanDataImpl{Network: net.Name, Name: evt.Params[2]})
		ci = net.ChannelInfo.get(evt.Params[2])
	}

	users := strings.Split(evt.Trailing, " ")
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

func (net *netImpl) userlistend(evt *msg.Message) {
	ci := net.ChannelInfo.get(evt.Params[1])
	if ci == nil {
		net.ChannelInfo.Put(&chanDataImpl{Network: net.Name, Name: evt.Params[1]})
		ci = net.ChannelInfo.get(evt.Params[1])
	}

	ci.ReceivingUserList = false
	sort.Sort(ci.UserList)
	net.Owner.NewMessages <- mauircdi.Message{Type: "chandata", Object: ci}
}

func (net *netImpl) chanlist(evt *msg.Message) {
	net.ChannelList = append(net.ChannelList, evt.Params[1])
}

func (net *netImpl) chanlistend(evt *msg.Message) {
	net.Owner.NewMessages <- mauircdi.Message{Type: "chanlist", Object: mauircdi.ChanList{Network: net.Name, List: net.ChannelList}}
}

func (net *netImpl) topic(evt *msg.Message) {
	ci := net.ChannelInfo.get(evt.Params[0])
	if ci == nil {
		net.ChannelInfo.Put(&chanDataImpl{Network: net.Name, Name: evt.Params[0]})
		ci = net.ChannelInfo.get(evt.Params[0])
	}
	ci.Topic = evt.Trailing
	ci.TopicSetBy = evt.Name
	ci.TopicSetAt = time.Now().Unix()
	net.ReceiveMessage(ci.Name, evt.Name, "topic", evt.Trailing)
	net.Owner.NewMessages <- mauircdi.Message{Type: "chandata", Object: ci}
}

func (net *netImpl) topicresp(evt *msg.Message) {
	ci := net.ChannelInfo.get(evt.Params[1])
	if ci == nil {
		net.ChannelInfo.Put(&chanDataImpl{Network: net.Name, Name: evt.Params[1]})
		ci = net.ChannelInfo.get(evt.Params[1])
	}
	ci.Topic = evt.Trailing
	net.Owner.NewMessages <- mauircdi.Message{Type: "chandata", Object: ci}
}

func (net *netImpl) noperms(evt *msg.Message) {
	net.Owner.NewMessages <- mauircdi.Message{
		Type: "message",
		Object: database.Message{
			ID:        -1,
			Network:   net.Name,
			Channel:   evt.Params[1],
			Timestamp: time.Now().Unix(),
			Sender:    "[" + net.Name + "]",
			Command:   "privmsg",
			Message:   evt.Trailing,
			OwnMsg:    false,
		},
	}
}

func (net *netImpl) topicset(evt *msg.Message) {
	ci := net.ChannelInfo.get(evt.Params[1])
	if ci == nil {
		net.ChannelInfo.Put(&chanDataImpl{Network: net.Name, Name: evt.Params[1]})
		ci = net.ChannelInfo.get(evt.Params[1])
	}
	ci.TopicSetBy = evt.Params[2]
	setAt, err := strconv.ParseInt(evt.Params[3], 10, 64)
	if err != nil {
		ci.TopicSetAt = setAt
	}
	net.Owner.NewMessages <- mauircdi.Message{Type: "chandata", Object: ci}
}

func (net *netImpl) quit(evt *msg.Message) {
	for _, ci := range net.ChannelInfo {
		if b, i := ci.UserList.Contains(evt.Name); b {
			ci.UserList[i] = ci.UserList[len(ci.UserList)-1]
			ci.UserList = ci.UserList[:len(ci.UserList)-1]
			sort.Sort(ci.UserList)

			net.ReceiveMessage(ci.Name, evt.Name, "quit", evt.Trailing)
			net.Owner.NewMessages <- mauircdi.Message{Type: "chandata", Object: ci}
		}
	}
}

func (net *netImpl) join(evt *msg.Message) {
	net.ReceiveMessage(evt.Params[0], evt.Name, "join", evt.Trailing)
	net.joinpart(evt.Name, evt.Params[0], false)
}

func (net *netImpl) part(evt *msg.Message) {
	net.ReceiveMessage(evt.Params[0], evt.Name, "part", evt.Trailing)
	net.joinpart(evt.Name, evt.Params[0], true)
}

func (net *netImpl) kick(evt *msg.Message) {
	net.ReceiveMessage(evt.Params[0], evt.Name, "kick", evt.Params[1]+":"+evt.Trailing)
	net.joinpart(evt.Params[1], evt.Params[0], true)
}

func (net *netImpl) privmsg(evt *msg.Message) {
	if evt.IsServer() {
		return
	} else if evt.Params[0][0] == '@' {
		evt.Params[0] = evt.Params[0][1:]
	}
	net.ReceiveMessage(evt.Params[0], evt.Name, "privmsg", evt.Trailing)
}

func (net *netImpl) action(evt *msg.Message) {
	net.ReceiveMessage(evt.Params[0], evt.Name, "action", evt.Trailing)
}

func (net *netImpl) invite(evt *msg.Message) {
	net.GetOwner().GetMessageChan() <- mauircdi.Message{Type: "invite", Object: mauircdi.Invite{
		Network: net.Name,
		Channel: evt.Params[1],
		Sender:  evt.Name,
	}}
}

func (net *netImpl) connected(evt *msg.Message) {
	net.IRC.List()
	for channel := range net.ChannelInfo {
		if strings.HasPrefix(channel, "#") {
			net.IRC.Join(channel, "")
		}
	}
	net.GetOwner().GetMessageChan() <- mauircdi.Message{Type: "netdata", Object: mauircdi.NetData{Name: net.GetName(), Connected: true}}
}

func (net *netImpl) disconnected(evt *msg.Message) {
	log.Warnf("Disconnected from %s:%d\n", net.IP, net.Port)
	net.GetOwner().GetMessageChan() <- mauircdi.Message{Type: "netdata", Object: mauircdi.NetData{Name: net.GetName(), Connected: false}}
}

func (net *netImpl) joinpart(user, channel string, part bool) {
	ci := net.ChannelInfo.get(channel)
	if ci == nil {
		net.ChannelInfo.Put(&chanDataImpl{Network: net.Name, Name: channel})
		ci = net.ChannelInfo.get(channel)
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

	if user == net.IRC.GetNick() && part {
		net.ChannelInfo.Remove(channel)
	}
}

func (net *netImpl) isAway(evt *msg.Message) {
	data := net.GetWhoisDataIfExists(evt.Params[1])
	if data != nil {
		data.Away = evt.Trailing
	}
}

func (net *netImpl) whoisUser(evt *msg.Message) {
	data := net.GetWhoisData(evt.Params[1])
	data.User = evt.Params[2]
	data.Host = evt.Params[3]
	data.RealName = evt.Trailing

}

func (net *netImpl) whoisServer(evt *msg.Message) {
	data := net.GetWhoisData(evt.Params[1])
	data.Server = evt.Params[2]
	data.ServerInfo = evt.Trailing
}

func (net *netImpl) whoisSecure(evt *msg.Message) {
	data := net.GetWhoisData(evt.Params[1])
	data.SecureConn = true
}

func (net *netImpl) whoisOperator(evt *msg.Message) {
	data := net.GetWhoisData(evt.Params[1])
	data.Operator = true
}

func (net *netImpl) whoisIdle(evt *msg.Message) {
	data := net.GetWhoisData(evt.Params[1])
	time, _ := strconv.ParseInt(evt.Params[2], 10, 64)
	data.IdleTime = time
}

func (net *netImpl) whoisChannels(evt *msg.Message) {
	data := net.GetWhoisData(evt.Params[1])
	for _, ch := range strings.Split(evt.Trailing, " ") {
		if len(ch) <= 0 {
			continue
		}
		var prefix string
		if userlist.LevelOfByte(ch[0]) > 0 {
			prefix = userlist.NameOf(userlist.LevelOfByte(ch[0]))
			ch = ch[1:]
		}
		data.Channels[ch] = prefix
	}
}

func (net *netImpl) whoisEnd(evt *msg.Message) {
	data := net.GetWhoisData(evt.Params[1])
	net.Owner.NewMessages <- mauircdi.Message{Type: "whois", Object: data}
	net.RemoveWhoisData(evt.Params[1])
}

func (net *netImpl) rawHandler(evt *msg.Message) {
	// libmauirc adds the trailing text as a param, remove it.
	evt.Params = evt.Params[:len(evt.Params)-1]
	net.Owner.NewMessages <- mauircdi.Message{Type: "raw", Object: mauircdi.RawMessage{Network: net.GetName(), Message: evt.String()}}
}
