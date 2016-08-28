// mauIRC-server - The IRC bouncer/backend system for mauIRC clients.
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
	"io/ioutil"
	"maunium.net/go/mauirc-server/interfaces"
	"maunium.net/go/mauirc-server/plugin"
	"os"
	"path/filepath"
	"strings"
)

// LoadScripts loads the scripts of this network
func (net *netImpl) LoadScripts(path string) error {
	path = filepath.Join(path, net.Owner.Email, net.Name)

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err = os.MkdirAll(path, 0700)
		if err != nil {
			return err
		}
	}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	for _, f := range files {
		data, err := ioutil.ReadFile(filepath.Join(path, f.Name()))
		if err != nil {
			log.Errorf("Failed to read script \"%s\" for network %s owned by %s\n", f.Name(), net.Name, net.Owner.Email)
		}
		net.Scripts = append(net.Scripts, plugin.Script{TheScript: string(data), Name: strings.Split(f.Name(), ".")[0]})
	}
	return nil
}

// SaveScripts saves the scripts of this network
func (net *netImpl) SaveScripts(path string) error {
	path = filepath.Join(path, net.Owner.Email, net.Name)

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err = os.MkdirAll(path, 0700)
		if err != nil {
			return err
		}
	}

	for _, script := range net.Scripts {
		err := ioutil.WriteFile(filepath.Join(path, script.GetName()+".ank"), []byte(script.GetScript()), 0644)
		if err != nil {
			log.Errorf("Failed to save script \"%s\" for network %s owned by %s\n", script.GetName()+".ank", net.Name, net.Owner.Email)
		}
	}
	return nil
}

// LoadGlobalScripts loads the global scripts of this user
func (user *userImpl) LoadGlobalScripts(path string) error {
	path = filepath.Join(path, user.Email, "global")

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err = os.MkdirAll(path, 0700)
		if err != nil {
			return err
		}
	}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	for _, f := range files {
		data, err := ioutil.ReadFile(filepath.Join(path, f.Name()))
		if err != nil {
			log.Errorf("Failed to read global script \"%s\" owned by %s\n", f.Name(), user.Email)
		}
		user.GlobalScripts = append(user.GlobalScripts, plugin.Script{TheScript: string(data), Name: strings.Split(f.Name(), ".")[0]})
	}
	return nil
}

// SaveGlobalScripts saves the global scripts of this user
func (user *userImpl) SaveGlobalScripts(path string) error {
	path = filepath.Join(path, user.Email, "global")

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err = os.MkdirAll(path, 0700)
		if err != nil {
			return err
		}
	}

	for _, script := range user.GlobalScripts {
		err := ioutil.WriteFile(filepath.Join(path, script.GetName()+".ank"), []byte(script.GetScript()), 0644)
		if err != nil {
			log.Errorf("Failed to save global script \"%s\" owned by %s\n", script.GetName()+".ank", user.Email)
		}
	}
	return nil
}

// RunScripts runs all the scripts of this network and all global scripts on the given message
func (net *netImpl) RunScripts(evt *interfaces.Event, receiving bool) {
	netChanged := false
	for _, s := range net.Scripts {
		netChanged = net.RunScript(s, evt, receiving)
		if netChanged {
			return
		}
	}

	for _, s := range net.Owner.GlobalScripts {
		netChanged = net.RunScript(s, evt, receiving)
		if netChanged {
			return
		}
	}
}

// RunScript runs a single script and sends it to another network if needed.
func (net *netImpl) RunScript(s interfaces.Script, evt *interfaces.Event, receiving bool) bool {
	s.Run(evt)
	if evt.Message.Network != net.Name {
		if net.SwitchMessageNetwork(evt.Message, receiving) {
			return true
		}
		evt.Message.Network = net.Name
	}
	return false
}

// DeleteScript deletes an entry from a script array
func DeleteScript(scripts []interfaces.Script, i int) []interfaces.Script {
	if i == 0 {
		return scripts[1:]
	} else if i == len(scripts)-1 {
		return scripts[:len(scripts)-1]
	} else {
		return append(scripts[:i], scripts[i+1:]...)
	}
}
