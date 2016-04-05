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
	"github.com/yuin/gopher-lua"
	"maunium.net/go/mauircd/database"
)

// Script wraps a Lua script.
type Script struct {
	TheScript string
	Name      string
}

func getString(L *lua.LState, event *lua.LTable, name string) string {
	return lua.LVAsString(L.GetField(event, name))
}

func getBool(L *lua.LState, event *lua.LTable, name string) bool {
	return lua.LVAsBool(L.GetField(event, name))
}

func getInt64(L *lua.LState, event *lua.LTable, name string) int64 {
	return int64(float64(lua.LVAsNumber(L.GetField(event, name))))
}

// Run the script with the given values.
func (s Script) Run(msg database.Message, cancelled bool) (database.Message, bool) {
	L := lua.NewState()
	L.OpenLibs()

	event := L.NewTypeMetatable("event")
	L.SetField(event, "network", lua.LString(msg.Network))
	L.SetField(event, "channel", lua.LString(msg.Channel))
	L.SetField(event, "timestamp", lua.LString(msg.Timestamp))
	L.SetField(event, "sender", lua.LString(msg.Sender))
	L.SetField(event, "command", lua.LString(msg.Command))
	L.SetField(event, "message", lua.LString(msg.Message))
	L.SetField(event, "ownmsg", lua.LBool(msg.OwnMsg))
	L.SetField(event, "cancelled", lua.LBool(cancelled))
	L.SetGlobal("event", event)

	defer L.Close()
	if err := L.DoString(s.TheScript); err != nil {
		fmt.Println(err)
	}

	msg.Network = getString(L, event, "network")
	msg.Channel = getString(L, event, "channel")
	msg.Sender = getString(L, event, "sender")
	msg.Command = getString(L, event, "command")
	msg.Message = getString(L, event, "message")
	msg.OwnMsg = getBool(L, event, "ownmsg")
	msg.Timestamp = getInt64(L, event, "timestamp")
	return msg, getBool(L, event, "cancelled")
}
