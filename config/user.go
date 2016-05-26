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
	"crypto/rand"
	"encoding/base64"
	"golang.org/x/crypto/bcrypt"
	"maunium.net/go/mauircd/interfaces"
	"strings"
	"time"
)

type userImpl struct {
	Networks      netListImpl            `json:"networks"`
	Email         string                 `json:"email"`
	Password      string                 `json:"password"`
	AuthTokens    []authToken            `json:"authtokens,omitempty"`
	NewMessages   chan mauircdi.Message  `json:"-"`
	GlobalScripts []mauircdi.Script      `json:"-"`
	Settings      interface{}            `json:"settings"`
	HostConf      mauircdi.Configuration `json:"-"`
}

type authToken struct {
	Token string `json:"token"`
	Time  int64  `json:"expire"`
}

type netListImpl []*netImpl

func (nl netListImpl) ForEach(do func(net mauircdi.Network)) {
	for _, net := range nl {
		do(net)
	}
}

func (user *userImpl) Save() {
	for _, net := range user.Networks {
		net.Save()
	}
}

// CheckPassword checks if the password is correct
func (user *userImpl) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil
}

// SetPassword sets the users password
func (user *userImpl) SetPassword(newPassword string) error {
	password, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(password)
	return nil
}

// GetNetwork gets the network with the given name
func (user *userImpl) GetNetwork(name string) mauircdi.Network {
	name = strings.ToLower(name)
	for _, network := range user.Networks {
		if network.Name == name {
			return network
		}
	}
	return nil
}

func (user *userImpl) GetNetworks() mauircdi.NetworkList {
	return user.Networks
}

func (user *userImpl) CheckAuthToken(token string) bool {
	var found = false

	var newAt = user.AuthTokens
	for i := 0; i < len(user.AuthTokens); i++ {
		if user.AuthTokens[i].Time < time.Now().Unix() {
			newAt[i] = newAt[len(newAt)-1]
			newAt = newAt[:len(newAt)-1]
		} else if user.AuthTokens[i].Token == token {
			found = true
		}
	}
	user.AuthTokens = newAt

	return found
}

func (user *userImpl) NewAuthToken() string {
	at := generateAuthToken()
	user.AuthTokens = append(user.AuthTokens, authToken{Token: at, Time: time.Now().Add(30 * 24 * time.Hour).Unix()})
	return at
}

func generateAuthToken() string {
	var authToken string
	b := make([]byte, 32)
	// Fill the byte array with cryptographically random bytes.
	n, err := rand.Read(b)
	if n == len(b) && err == nil {
		authToken = base64.RawStdEncoding.EncodeToString(b)
		return authToken
	}

	return ""
}

func (user *userImpl) GetEmail() string {
	return user.Email
}

func (user *userImpl) GetNameFromEmail() string {
	parts := strings.Split(user.GetEmail(), "@")
	return parts[0]
}

func (user *userImpl) GetMessageChan() chan mauircdi.Message {
	return user.NewMessages
}

func (user *userImpl) GetGlobalScripts() []mauircdi.Script {
	return user.GlobalScripts
}

func (user *userImpl) AddGlobalScript(s mauircdi.Script) bool {
	for i := 0; i < len(user.GlobalScripts); i++ {
		if user.GlobalScripts[i].GetName() == s.GetName() {
			user.GlobalScripts[i] = s
			return false
		}
	}
	user.GlobalScripts = append(user.GlobalScripts, s)
	return true
}

func (user *userImpl) RemoveGlobalScript(name string) bool {
	for i := 0; i < len(user.GlobalScripts); i++ {
		if user.GlobalScripts[i].GetName() == name {
			if i == 0 {
				user.GlobalScripts = user.GlobalScripts[1:]
			} else if i == len(user.GlobalScripts)-1 {
				user.GlobalScripts = user.GlobalScripts[:len(user.GlobalScripts)-1]
			} else {
				user.GlobalScripts = append(user.GlobalScripts[:i], user.GlobalScripts[i+1:]...)
			}
			return true
		}
	}
	return false
}

func (user *userImpl) GetSettings() interface{} {
	return user.Settings
}

func (user *userImpl) SetSettings(val interface{}) {
	user.Settings = val
}
