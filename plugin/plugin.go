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
	"github.com/yuin/gopher-lua"
)

// Script wraps a Lua script.
type Script string

// String turns the script into a string.
func (s Script) String() string {
	return string(s)
}

// Run the script with the given values.
func (s Script) Run(channel, sender, command, message string, cancelled bool) (string, string, string, string, bool) {
	L := lua.NewState()
	L.OpenLibs()

	L.SetGlobal("channel", lua.LString(channel))
	L.SetGlobal("sender", lua.LString(sender))
	L.SetGlobal("command", lua.LString(command))
	L.SetGlobal("message", lua.LString(message))
	L.SetGlobal("cancelled", lua.LBool(cancelled))

	defer L.Close()
	if err := L.DoString(s.String()); err != nil {
		panic(err)
	}

	return L.GetGlobal("channel").String(), L.GetGlobal("sender").String(),
		L.GetGlobal("command").String(), L.GetGlobal("message").String(),
		L.GetGlobal("cancelled").(bool)
}
