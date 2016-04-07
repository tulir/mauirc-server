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
	"io/ioutil"
	"maunium.net/go/mauircd/database"
	"maunium.net/go/mauircd/interfaces"
	"maunium.net/go/mauircd/plugin"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// LoadScripts loads the scripts of this network
func (net *netImpl) LoadScripts(path string) error {
	path = filepath.Join(path, net.Owner.Email, net.Name)

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err = os.MkdirAll(path, 0755)
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
			fmt.Printf("Failed to read script \"%s\" for network %s owned by %s\n", f.Name(), net.Name, net.Owner.Email)
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
		err = os.MkdirAll(path, 0755)
		if err != nil {
			return err
		}
	}

	for _, script := range net.Scripts {
		err := ioutil.WriteFile(filepath.Join(path, script.GetName()+".lua"), []byte(script.GetScript()), 0644)
		if err != nil {
			fmt.Printf("Failed to save script \"%s\" for network %s owned by %s\n", script.GetName()+".lua", net.Name, net.Owner.Email)
		}
	}
	return nil
}

// RunScripts runs all the scripts of this network and all global scripts on the given message
func (net *netImpl) RunScripts(msg database.Message, cancelled, receiving bool) (database.Message, bool) {
	netChanged := false
	for _, s := range net.Scripts {
		msg, cancelled, netChanged = net.RunScript(msg, s, cancelled, receiving)
		if netChanged {
			return msg, true
		}
	}

	for _, s := range net.Owner.GlobalScripts {
		msg, cancelled, netChanged = net.RunScript(msg, s, cancelled, receiving)
		if netChanged {
			return msg, true
		}
	}
	return msg, cancelled
}

// RunScript runs a single script and sends it to another network if needed.
func (net *netImpl) RunScript(msg database.Message, s mauircdi.Script, cancelled, receiving bool) (database.Message, bool, bool) {
	msg, cancelled = s.Run(msg, cancelled)
	if msg.Network != net.Name {
		if net.SwitchMessageNetwork(msg, receiving) {
			return msg, cancelled, true
		}
		msg.Network = net.Name
	}
	return msg, cancelled, false
}

func download(pasteID string) (string, error) {
	response, err := http.Get("http://pastebin.com/raw/" + pasteID)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	response.Body.Close()

	if len(data) == 0 {
		return "", fmt.Errorf("No data received!")
	}

	return string(data), nil
}
