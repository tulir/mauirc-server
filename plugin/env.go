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
	"fmt"
	anko_encoding_json "github.com/mattn/anko/builtins/encoding/json"
	anko_flag "github.com/mattn/anko/builtins/flag"
	//anko_io "github.com/mattn/anko/builtins/io"
	//anko_io_ioutil "github.com/mattn/anko/builtins/io/ioutil"
	anko_math "github.com/mattn/anko/builtins/math"
	anko_net "github.com/mattn/anko/builtins/net"
	anko_net_http "github.com/mattn/anko/builtins/net/http"
	anko_net_url "github.com/mattn/anko/builtins/net/url"
	//anko_os "github.com/mattn/anko/builtins/os"
	//anko_os_exec "github.com/mattn/anko/builtins/os/exec"
	anko_path "github.com/mattn/anko/builtins/path"
	anko_path_filepath "github.com/mattn/anko/builtins/path/filepath"
	anko_regexp "github.com/mattn/anko/builtins/regexp"
	anko_sort "github.com/mattn/anko/builtins/sort"
	anko_strings "github.com/mattn/anko/builtins/strings"
	"github.com/mattn/anko/vm"
	"maunium.net/go/mauircd/database"
	"maunium.net/go/mauircd/interfaces"
	"maunium.net/go/mauircd/util/preview"
)

// LoadImport loads the import() function into the given environment to allow importing some Go packages
func LoadImport(env *vm.Env) {
	tbl := map[string]func(env *vm.Env) *vm.Env{
		"encoding/json": anko_encoding_json.Import,
		"flag":          anko_flag.Import,
		//"io":            anko_io.Import,
		//"io/ioutil":     anko_io_ioutil.Import,
		"math":     anko_math.Import,
		"net":      anko_net.Import,
		"net/http": anko_net_http.Import,
		"net/url":  anko_net_url.Import,
		//"os":            anko_os.Import,
		//"os/exec":       anko_os_exec.Import,
		"path":          anko_path.Import,
		"path/filepath": anko_path_filepath.Import,
		"regexp":        anko_regexp.Import,
		"sort":          anko_sort.Import,
		"strings":       anko_strings.Import,
	}

	env.Define("import", func(s string) interface{} {
		if loader, ok := tbl[s]; ok {
			return loader(env)
		}
		panic(fmt.Sprintf("package '%s' not found", s))
	})
}

// LoadAll load all the bindings into the given Anko VM environment
func LoadAll(env *vm.Env, evt *mauircdi.Event) {
	LoadEvent(env.NewModule("event"), evt)
	LoadNetwork(env.NewModule("network"), evt)
	LoadUser(env.NewModule("user"), evt)
}

// LoadEvent loads event things into the given Anko VM environment
func LoadEvent(env *vm.Env, evt *mauircdi.Event) {
	env.Define("GetID", func() int64 {
		return evt.Message.ID
	})
	env.Define("SetID", func(val int64) {
		evt.Message.ID = val
	})
	env.Define("GetNetwork", func() string {
		return evt.Message.Network
	})
	env.Define("SetNetwork", func(val string) {
		evt.Message.Network = val
	})
	env.Define("GetChannel", func() string {
		return evt.Message.Channel
	})
	env.Define("SetChannel", func(val string) {
		evt.Message.Channel = val
	})
	env.Define("GetTimestamp", func() int64 {
		return evt.Message.Timestamp
	})
	env.Define("SetTimestamp", func(val int64) {
		evt.Message.Timestamp = val
	})
	env.Define("GetSender", func() string {
		return evt.Message.Sender
	})
	env.Define("SetSender", func(val string) {
		evt.Message.Sender = val
	})
	env.Define("GetCommand", func() string {
		return evt.Message.Command
	})
	env.Define("SetCommand", func(val string) {
		evt.Message.Command = val
	})
	env.Define("GetMessage", func() string {
		return evt.Message.Message
	})
	env.Define("SetMessage", func(val string) {
		evt.Message.Message = val
	})
	env.Define("IsOwnMsg", func() bool {
		return evt.Message.OwnMsg
	})
	env.Define("SetOwnMsg", func(val bool) {
		evt.Message.OwnMsg = val
	})
	env.Define("IsCancelled", func() bool {
		return evt.Cancelled
	})
	env.Define("SetCancelled", func(val bool) {
		evt.Cancelled = val
	})
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
			return
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
			return
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

	LoadIRC(env.NewModule("irc"), evt)
}

// LoadIRC loads irc command bindings into the given Anko VM environment
func LoadIRC(env *vm.Env, evt *mauircdi.Event) {
	env.Define("Nick", func(nick string) {
		evt.Network.Tunnel().SetNick(nick)
	})
	env.Define("Join", func(channels string, keys string) {
		evt.Network.Tunnel().Join(channels, keys)
	})
	env.Define("Part", func(channel string, reason string) {
		evt.Network.Tunnel().Part(channel, reason)
	})
	env.Define("Topic", func(channel string, topic string) {
		evt.Network.Tunnel().Topic(channel, topic)
	})
	env.Define("Privmsg", func(channel string, message string) {
		evt.Network.Tunnel().Privmsg(channel, message)
	})
}

// LoadUser loads user things into the given Anko VM environment
func LoadUser(env *vm.Env, evt *mauircdi.Event) {
	env.Define("GetEmail", evt.Network.GetOwner().GetEmail)
	env.Define("SendMessage", func(network, channel string, timestamp int64, sender, command, message string, ownmsg bool) {
		evt.Network.InsertAndSend(database.Message{
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
