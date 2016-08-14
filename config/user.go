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
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"maunium.net/go/mauirc-common/messages"
	"maunium.net/go/mauirc-server/interfaces"
	"strings"
	"time"
)

type userImpl struct {
	Networks      netListImpl             `json:"networks"`
	Email         string                  `json:"email"`
	Password      string                  `json:"password"`
	AuthTokens    []authToken             `json:"authtokens,omitempty"`
	PasswordReset *authToken              `json:"passwordreset,omitempty"`
	EmailVerify   *authToken              `json:"emailverify,omitempty"`
	NewMessages   chan messages.Container `json:"-"`
	GlobalScripts []interfaces.Script     `json:"-"`
	Settings      interface{}             `json:"settings,omitempty"`
	HostConf      *configImpl             `json:"-"`
}

type authToken struct {
	Token string `json:"token"`
	Time  int64  `json:"expire"`
}

func (at authToken) HasExpired() bool {
	return at.Time > time.Now().Unix()
}

type netListImpl []*netImpl

func (nl netListImpl) ForEach(do func(net interfaces.Network)) {
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

func (user *userImpl) InitNetworks() {
	for _, network := range user.Networks {
		if network.Owner != nil {
			continue
		}
		network.ChannelInfo = make(map[string]*chanDataImpl)
		network.Owner = user
		network.Open()
		network.LoadScripts(user.HostConf.Path)
	}
}

func (user *userImpl) SendNetworkData(net interfaces.Network) {
	user.NewMessages <- messages.Container{Type: messages.MsgNetData, Object: net.GetNetData()}
	net.GetActiveChannels().ForEach(func(chd interfaces.ChannelData) {
		user.NewMessages <- messages.Container{Type: messages.MsgChanData, Object: chd}
	})
	user.NewMessages <- messages.Container{Type: messages.MsgChanList, Object: messages.ChanList{Network: net.GetName(), List: net.GetAllChannels()}}
}

// GetNetwork gets the network with the given name
func (user *userImpl) GetNetwork(name string) interfaces.Network {
	name = strings.ToLower(name)
	for _, network := range user.Networks {
		if network.Name == name {
			return network
		}
	}
	return nil
}

func (user *userImpl) DeleteNetwork(name string) bool {
	name = strings.ToLower(name)
	for i, network := range user.Networks {
		if network.Name == name {
			if i == 0 {
				user.Networks = user.Networks[1:]
			} else if i == len(user.Networks)-1 {
				user.Networks = user.Networks[:len(user.Networks)-1]
			} else {
				user.Networks = append(user.Networks[:i], user.Networks[i+1:]...)
			}
			return true
		}
	}
	return false
}

func (user *userImpl) AddNetwork(net interfaces.Network) bool {
	network, ok := net.(*netImpl)
	if ok {
		network.Name = strings.ToLower(network.Name)
		user.Networks = append(user.Networks, network)
		return true
	}
	return false
}

func (user *userImpl) CreateNetwork(name string, data []byte) (interfaces.Network, bool) {
	var net = &netImpl{}
	err := json.Unmarshal(data, net)
	if err != nil {
		return nil, false
	}
	net.Name = strings.ToLower(name)
	return net, true
}

func (user *userImpl) GetNetworks() interfaces.NetworkList {
	return user.Networks
}

func (user *userImpl) CheckAuthToken(token string) bool {
	var found = false

	var newAt = user.AuthTokens
	for i := 0; i < len(user.AuthTokens); i++ {
		if user.AuthTokens[i].HasExpired() {
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

func (user *userImpl) NewResetToken() (token string, timed time.Time) {
	token = generateAuthToken()
	timed = time.Now().Add(30 * time.Minute)
	user.PasswordReset = &authToken{Token: token, Time: timed.Unix()}
	return token, timed
}

func (user *userImpl) CheckResetToken(token string) bool {
	if user.PasswordReset == nil || user.PasswordReset.HasExpired() || user.PasswordReset.Token != token {
		return false
	}
	return true
}

func (user *userImpl) ClearResetToken() {
	user.PasswordReset = nil
}

func (user *userImpl) IsVerified() bool {
	return !user.HostConf.Mail.Enabled || user.EmailVerify == nil
}

func (user *userImpl) SetVerified() {
	user.EmailVerify = nil
}

func generateAuthToken() string {
	var authToken string
	b := make([]byte, 32)
	// Fill the byte array with cryptographically random bytes.
	n, err := rand.Read(b)
	if n == len(b) && err == nil {
		authToken = base64.RawURLEncoding.EncodeToString(b)
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

func (user *userImpl) GetMessageChan() chan messages.Container {
	return user.NewMessages
}

func (user *userImpl) GetGlobalScripts() []interfaces.Script {
	return user.GlobalScripts
}

func (user *userImpl) AddGlobalScript(s interfaces.Script) bool {
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
