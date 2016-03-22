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
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strings"
)

var config *Configuration

// Load the config at the given path
func Load(path string) error {
	config = &Configuration{}
	data, err := ioutil.ReadFile(filepath.Join(path, "config.json"))
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, config)
	if err != nil {
		return err
	}

	config.Path = path
	for _, user := range config.Users {
		for _, network := range user.Networks {
			network.Open(user)
		}
	}

	return nil
}

// Save the configuration file
func Save() error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(config.Path, "config.json"), data, 0644)
	return err
}

// GetUser gets the user with the given email
func GetUser(email string) *User {
	email = strings.ToLower(email)
	for _, user := range config.Users {
		if user.Email == email {
			return user
		}
	}
	return nil
}

// GetUsers returns all users
func GetUsers() []*User {
	return config.Users
}

// GetSQLConfig gets the SQL configuration
func GetSQLConfig() SQLConfig {
	return config.SQL
}
