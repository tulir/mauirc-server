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
	"maunium.net/go/mauircd/plugin"
	"os"
	"path/filepath"
	"strings"
)

// LoadScripts loads the scripts of this network
func (net *Network) LoadScripts(path string) error {
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
func (net *Network) SaveScripts(path string) error {
	path = filepath.Join(path, net.Owner.Email, net.Name)

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err = os.MkdirAll(path, 0755)
		if err != nil {
			return err
		}
	}

	for _, script := range net.Scripts {
		err := ioutil.WriteFile(filepath.Join(path, script.Name+".lua"), []byte(script.TheScript), 0644)
		if err != nil {
			fmt.Printf("Failed to save script \"%s\" for network %s owned by %s\n", script.Name+".lua", net.Name, net.Owner.Email)
		}
	}
	return nil
}
