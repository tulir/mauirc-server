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
	"github.com/mattn/anko/vm"
	"maunium.net/go/mauircd/interfaces"
)

// Script wraps a Lua script.
type Script struct {
	TheScript string
	Name      string
}

// GetName returns the name of the script
func (s Script) GetName() string {
	return s.Name
}

// GetScript returns the script data
func (s Script) GetScript() string {
	return s.TheScript
}

// Run the script with the given values.
func (s Script) Run(evt *mauircdi.Event) {
	var env = vm.NewEnv()

	var event = env.NewModule("event")
	LoadEvent(event, evt)
	var network = env.NewModule("network")
	LoadNetwork(network, evt)
	var user = env.NewModule("user")
	LoadUser(user, evt)

	val, err := env.Execute(s.GetScript())
	if err != nil {
		fmt.Println(val, err)
	}

	evt.Message.Network = getString(event, "network", evt.Message.Network)
	evt.Message.Channel = getString(event, "channel", evt.Message.Channel)
	evt.Message.Timestamp = getInt(event, "timestamp", evt.Message.Timestamp)
	evt.Message.Sender = getString(event, "sender", evt.Message.Sender)
	evt.Message.Command = getString(event, "command", evt.Message.Command)
	evt.Message.Message = getString(event, "message", evt.Message.Message)
	evt.Message.OwnMsg = getBool(event, "ownevt.Message", evt.Message.OwnMsg)
	evt.Cancelled = getBool(event, "cancelled", evt.Cancelled)
}

func getString(env *vm.Env, path string, def string) string {
	val, err := env.Get(path)
	if err != nil {
		return def
	}
	return val.String()
}

func getBool(env *vm.Env, path string, def bool) bool {
	val, err := env.Get(path)
	if err != nil {
		return def
	}
	return val.Bool()
}

func getInt(env *vm.Env, path string, def int64) int64 {
	val, err := env.Get(path)
	if err != nil {
		return def
	}
	return val.Int()
}
