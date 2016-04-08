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

// Package plugin contains Lua plugin executing stuff
package plugin

import (
	"github.com/mattn/anko/vm"
	"maunium.net/go/mauircd/database"
	"maunium.net/go/mauircd/interfaces"
	"maunium.net/go/mauircd/util/preview"
)

// LoadAll load all the bindings into the given Anko VM environment
func LoadAll(env *vm.Env, evt *mauircdi.Event) {
	LoadEvent(env.NewModule("event"), evt)
	LoadNetwork(env.NewModule("network"), evt)
	LoadUser(env.NewModule("user"), evt)
}

// LoadEvent loads event things into the given Anko VM environment
func LoadEvent(env *vm.Env, evt *mauircdi.Event) {
	env.Define("network", evt.Message.Network)
	env.Define("channel", evt.Message.Channel)
	env.Define("timestamp", evt.Message.Timestamp)
	env.Define("sender", evt.Message.Sender)
	env.Define("command", evt.Message.Command)
	env.Define("message", evt.Message.Message)
	env.Define("ownmsg", evt.Message.OwnMsg)
	env.Define("cancelled", evt.Cancelled)
	LoadPreview(env.NewModule("preview"), evt)
}

// LoadPreview loads preview things into the given Anko VM environment
func LoadPreview(env *vm.Env, evt *mauircdi.Event) {
	env.Define("HasPreview", func() bool {
		return evt.Message.Preview != nil
	})

	env.Define("RemovePreview", func() {
		evt.Message.Preview = nil
	})

	env.Define("SetPreviewURL", func(url string) bool {
		newPreview, err := preview.GetPreview(url)
		if err != nil {
			return false
		}
		evt.Message.Preview = newPreview
		return true
	})

	env.Define("SetPreviewImage", func(url, typ string) {
		if len(url) == 0 && len(typ) == 0 {
			evt.Message.Preview.Image = nil
			if evt.Message.Preview.Text == nil {
				evt.Message.Preview = nil
			}
		}
		imgPreview := &preview.Image{URL: url, Type: typ}
		if evt.Message.Preview == nil {
			evt.Message.Preview = &preview.Preview{}
		}
		evt.Message.Preview.Image = imgPreview
	})

	env.Define("SetPreviewText", func(title, description, sitename string) {
		if len(title) == 0 && len(description) == 0 && len(sitename) == 0 {
			evt.Message.Preview.Text = nil
			if evt.Message.Preview.Image == nil {
				evt.Message.Preview = nil
			}
		} else if title == description {
			description = ""
		}
		textPreview := &preview.Text{Title: title, Description: description, SiteName: sitename}
		if evt.Message.Preview == nil {
			evt.Message.Preview = &preview.Preview{}
		}
		evt.Message.Preview.Text = textPreview
	})
}

// LoadNetwork loads network things into the given Anko VM environment
func LoadNetwork(env *vm.Env, evt *mauircdi.Event) {
	env.Define("GetNick", evt.Network.GetNick)
	env.Define("GetTopic", func(channel string) string {
		ch, ok := evt.Network.GetActiveChannels().Get(channel)
		if !ok {
			return ""
		}
		return ch.GetTopic()
	})
	env.Define("GetChannels", func() []string {
		var channels []string
		evt.Network.GetActiveChannels().ForEach(func(ch mauircdi.ChannelData) {
			channels = append(channels, ch.GetName())
		})
		return channels
	})
	env.Define("GetAllChannels", evt.Network.GetAllChannels)
	env.Define("SendFakeMessage", evt.Network.SendMessage)
	env.Define("ReceiveFakeMessage", evt.Network.ReceiveMessage)

	var irc = env.NewModule("irc")
	{
		irc.Define("Nick", func(nick string) {
			evt.Network.SendRaw("NICK %s", nick)
		})
		irc.Define("Join", func(channel string, keys string) {
			evt.Network.SendRaw("JOIN %s %s", channel, keys)
		})
		irc.Define("Part", func(channel string, reason string) {
			evt.Network.SendRaw("PART %s :%s", channel, reason)
		})
		irc.Define("Topic", func(channel string, topic string) {
			evt.Network.SendRaw("TOPIC %s :%s", channel, topic)
		})
	}
}

// LoadUser loads user things into the given Anko VM environment
func LoadUser(env *vm.Env, evt *mauircdi.Event) {
	env.Define("email", evt.Network.GetOwner().GetEmail())
	env.Define("SendMessage", func(id int64, network, channel string, timestamp int64, sender, command, message string, ownmsg bool) {
		evt.Network.InsertAndSend(database.Message{
			ID:        id,
			Network:   network,
			Channel:   channel,
			Timestamp: timestamp,
			Sender:    sender,
			Command:   command,
			Message:   message,
			OwnMsg:    ownmsg,
		})
	})
	env.Define("SendDirectMessage", func(id int64, network, channel string, timestamp int64, sender, command, message string, ownmsg bool) {
		evt.Network.GetOwner().GetMessageChan() <- mauircdi.Message{
			Type: "message",
			Object: database.Message{
				ID:        id,
				Network:   network,
				Channel:   channel,
				Timestamp: timestamp,
				Sender:    sender,
				Command:   command,
				Message:   message,
				OwnMsg:    ownmsg,
			},
		}
	})
	env.Define("SendRawMessage", func(typ string, data string) {
		evt.Network.GetOwner().GetMessageChan() <- mauircdi.Message{Type: typ, Object: data}
	})
	env.Define("GetNetworks", func() []string {
		var networks []string
		evt.Network.GetOwner().GetNetworks().ForEach(func(net mauircdi.Network) {
			networks = append(networks, evt.Network.GetName())
		})
		return networks
	})
}
