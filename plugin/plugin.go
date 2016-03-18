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
func (s Script) Run(channel, sender, command, message string) (string, string, string, string) {
	L := lua.NewState()
	L.OpenLibs()

	L.SetGlobal("channel", lua.LString(channel))
	L.SetGlobal("sender", lua.LString(sender))
	L.SetGlobal("command", lua.LString(command))
	L.SetGlobal("message", lua.LString(message))

	defer L.Close()
	if err := L.DoString(s.String()); err != nil {
		panic(err)
	}

	return L.GetGlobal("channel").String(), L.GetGlobal("sender").String(),
		L.GetGlobal("command").String(), L.GetGlobal("message").String()
}
