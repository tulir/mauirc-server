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
	"github.com/thoj/go-ircevent"
	"maunium.net/go/mauircd/database"
	"maunium.net/go/mauircd/interfaces"
	"maunium.net/go/mauircd/util/preview"
	"maunium.net/go/mauircd/util/split"
	"maunium.net/go/mauircd/util/userlist"
	"strings"
	"time"
)

type netImpl struct {
	Name     string   `json:"name"`
	Nick     string   `json:"nick"`
	User     string   `json:"user"`
	Realname string   `json:"realname"`
	IP       string   `json:"ip"`
	Port     int      `json:"port"`
	Password string   `json:"password"`
	SSL      bool     `json:"ssl"`
	Chs      []string `json:"channels"`

	Owner       *userImpl             `json:"-"`
	IRC         *irc.Connection       `json:"-"`
	Scripts     []mauircdi.Script     `json:"-"`
	ChannelInfo cdlImpl               `json:"-"`
	ChannelList []string              `json:"-"`
	WhoisData   map[string]*whoisData `json:"-"`
}

type whoisData struct {
	Channels   map[string]rune `json:"channels"`
	Nick       string          `json:"nick"`
	User       string          `json:"user"`
	Host       string          `json:"host"`
	RealName   string          `json:"realname"`
	Away       string          `json:"away"`
	Server     string          `json:"server"`
	ServerInfo string          `json:"server-info"`
	IdleTime   int64           `json:"idle"`
	SecureConn bool            `json:"secure-connection"`
	Operator   bool            `json:"operator"`
}

func (net *netImpl) Save() {
	net.Chs = []string{}
	for ch := range net.ChannelInfo {
		net.Chs = append(net.Chs, ch)
	}
}

// Open an IRC connection
func (net *netImpl) Open() {
	i := irc.IRC(net.Nick, net.User)

	i.UseTLS = net.SSL
	i.QuitMessage = "mauIRCd shutting down..."
	if len(net.Password) > 0 {
		i.Password = net.Password
	}

	for _, ch := range net.Chs {
		net.ChannelInfo.Put(&chanDataImpl{Network: net.Name, Name: ch})
	}

	net.IRC = i

	i.AddCallback("PRIVMSG", net.privmsg)
	i.AddCallback("NOTICE", net.privmsg)
	i.AddCallback("CPRIVMSG", net.privmsg)
	i.AddCallback("CNOTICE", net.privmsg)
	i.AddCallback("CTCP_ACTION", net.action)
	i.AddCallback("JOIN", net.join)
	i.AddCallback("PART", net.part)
	i.AddCallback("KICK", net.kick)
	i.AddCallback("MODE", net.mode)
	i.AddCallback("TOPIC", net.topic)
	i.AddCallback("NICK", net.nick)
	i.AddCallback("QUIT", net.quit)
	i.AddCallback("DISCONNECTED", net.disconnected)
	i.AddCallback("001", net.connected)
	i.AddCallback("353", net.userlist)
	i.AddCallback("366", net.userlistend)
	i.AddCallback("322", net.chanlist)
	i.AddCallback("323", net.chanlistend)
	i.AddCallback("332", net.topicresp)
	i.AddCallback("333", net.topicset)
	i.AddCallback("482", net.noperms)
	i.AddCallback("301", net.isAway)
	i.AddCallback("311", net.whoisUser)
	i.AddCallback("312", net.whoisServer)
	i.AddCallback("313", net.whoisOperator)
	i.AddCallback("317", net.whoisIdle)
	i.AddCallback("318", net.whoisEnd)
	i.AddCallback("319", net.whoisChannels)

	if err := net.Connect(); err != nil {
		fmt.Printf("Failed to connect to %s:%d: %s", net.IP, net.Port, err)
	}
}

func (net *netImpl) Connect() error {
	err := net.IRC.Connect(fmt.Sprintf("%s:%d", net.IP, net.Port))
	return err
}

func (net *netImpl) Disconnect() {
	net.IRC.Disconnect()
}

func (net *netImpl) IsConnected() bool {
	return net.IRC.Connected()
}

// ReceiveMessage stores the message and sends it to the client
func (net *netImpl) ReceiveMessage(channel, sender, command, message string) {
	msg := database.Message{Network: net.Name, Channel: channel, Timestamp: time.Now().Unix(), Sender: sender, Command: command, Message: message}

	if msg.Sender == net.Nick || (command == "nick" && message == net.Nick) {
		msg.OwnMsg = true
	} else {
		msg.OwnMsg = false
	}

	if msg.Channel == "AUTH" || msg.Channel == "*" {
		return
	} else if msg.Channel == net.Nick {
		if len(msg.Sender) > 0 {
			net.GetActiveChannels().Put(&chanDataImpl{Network: net.Name, Name: msg.Sender})
		}
		msg.Channel = msg.Sender
	}

	var evt = &mauircdi.Event{Message: msg, Network: net, Cancelled: false}
	net.RunScripts(evt, true)
	if evt.Cancelled {
		return
	}
	msg = evt.Message

	if len(msg.Channel) == 0 || len(msg.Command) == 0 {
		return
	}

	net.InsertAndSend(msg)
}

// SendMessage sends the given message to the given channel
func (net *netImpl) SendMessage(channel, command, message string) {
	msg := database.Message{Network: net.Name, Channel: channel, Timestamp: time.Now().Unix(), Sender: net.Nick, Command: command, Message: message}

	if msg.Sender == net.Nick {
		msg.OwnMsg = true
	} else {
		msg.OwnMsg = false
	}

	var evt = &mauircdi.Event{Message: msg, Network: net, Cancelled: false}
	net.RunScripts(evt, true)
	if evt.Cancelled {
		return
	}
	msg = evt.Message

	if splitted := split.All(msg.Message); len(splitted) > 1 {
		for _, piece := range splitted {
			net.SendMessage(msg.Channel, msg.Command, piece)
		}
		return
	}

	if net.sendToIRC(msg) {
		net.InsertAndSend(msg)
	}
}

func (net *netImpl) sendToIRC(msg database.Message) bool {
	if !strings.HasPrefix(msg.Channel, "*") {
		switch msg.Command {
		case "privmsg":
			net.IRC.Privmsg(msg.Channel, msg.Message)
			return true
		case "action":
			net.IRC.Action(msg.Channel, msg.Message)
			return true
		case "topic":
			net.IRC.SendRawf("TOPIC %s :%s", msg.Channel, msg.Message)
		case "join":
			net.IRC.Join(msg.Channel)
		case "part":
			net.IRC.Part(msg.Channel)
		case "nick":
			net.IRC.Nick(msg.Message)
		case "whois":
			net.IRC.Whois(msg.Channel)
		}
	}
	return false
}

// SwitchNetwork sends the given message to another network
func (net *netImpl) SwitchMessageNetwork(msg database.Message, receiving bool) bool {
	newNet := net.Owner.GetNetwork(msg.Network)
	if newNet != nil {
		if receiving {
			newNet.ReceiveMessage(msg.Channel, msg.Sender, msg.Command, msg.Message)
		} else {
			newNet.SendMessage(msg.Channel, msg.Command, msg.Message)
		}
		return true
	}
	return false
}

// InsertAndSend inserts the given message into the database and sends it to the client
func (net *netImpl) InsertAndSend(msg database.Message) {
	msg.Preview, _ = preview.GetPreview(msg.Message)
	msg.ID = database.Insert(net.Owner.Email, msg)
	net.Owner.NewMessages <- mauircdi.Message{Type: "message", Object: msg}
}

// Close the IRC connection.
func (net *netImpl) Close() {
	if net.IRC.Connected() {
		net.IRC.Quit()
	}
}

func (net *netImpl) GetOwner() mauircdi.User {
	return net.Owner
}

func (net *netImpl) GetName() string {
	return net.Name
}

func (net *netImpl) GetNick() string {
	return net.Nick
}

func (net *netImpl) GetActiveChannels() mauircdi.ChannelDataList {
	return net.ChannelInfo
}

func (net *netImpl) GetAllChannels() []string {
	return net.ChannelList
}

func (net *netImpl) SendRaw(msg string, args ...interface{}) {
	net.IRC.SendRawf(msg, args...)
}

func (net *netImpl) GetScripts() []mauircdi.Script {
	return net.Scripts
}

func (net *netImpl) AddScript(s mauircdi.Script) bool {
	for i := 0; i < len(net.Scripts); i++ {
		if net.Scripts[i].GetName() == s.GetName() {
			net.Scripts[i] = s
			return false
		}
	}
	net.Scripts = append(net.Scripts, s)
	return true
}

func (net *netImpl) RemoveScript(name string) bool {
	for i := 0; i < len(net.Scripts); i++ {
		if net.Scripts[i].GetName() == name {
			if i == 0 {
				net.Scripts = net.Scripts[1:]
			} else if i == len(net.Scripts)-1 {
				net.Scripts = net.Scripts[:len(net.Scripts)-1]
			} else {
				net.Scripts = append(net.Scripts[:i], net.Scripts[i+1:]...)
			}
			return true
		}
	}
	return false
}

func (net *netImpl) GetWhoisData(name string) *whoisData {
	data, ok := net.WhoisData[name]
	if !ok {
		net.WhoisData[name] = &whoisData{Nick: name}
		return net.WhoisData[name]
	}
	return data
}

func (net *netImpl) GetWhoisDataIfExists(name string) *whoisData {
	data, ok := net.WhoisData[name]
	if !ok {
		return nil
	}
	return data
}

func (net *netImpl) RemoveWhoisData(name string) {
	net.WhoisData[name] = nil
	delete(net.WhoisData, name)
}

type chanDataImpl struct {
	Network           string            `json:"network"`
	Name              string            `json:"name"`
	UserList          userlist.List     `json:"userlist"`
	Topic             string            `json:"topic"`
	TopicSetBy        string            `json:"topicsetby"`
	TopicSetAt        int64             `json:"topicsetat"`
	ModeList          mauircdi.ModeList `json:"modes"`
	ReceivingUserList bool              `json:"-"`
}

func (cd *chanDataImpl) GetUsers() []string {
	return cd.UserList
}

func (cd *chanDataImpl) GetName() string {
	return cd.Name
}

func (cd *chanDataImpl) GetTopic() string {
	return cd.Topic
}

func (cd *chanDataImpl) GetNetwork() string {
	return cd.Network
}

func (cd *chanDataImpl) Modes() mauircdi.ModeList {
	return cd.ModeList
}

type cdlImpl map[string]*chanDataImpl

func (cdl cdlImpl) Get(channel string) (mauircdi.ChannelData, bool) {
	val, ok := cdl[channel]
	return val, ok
}

func (cdl cdlImpl) Put(data mauircdi.ChannelData) {
	dat, ok := data.(*chanDataImpl)
	if ok {
		cdl[data.GetName()] = dat
	}
}

func (cdl cdlImpl) Remove(channel string) {
	delete(cdl, channel)
}

func (cdl cdlImpl) Has(channel string) bool {
	_, ok := cdl[channel]
	return ok
}

func (cdl cdlImpl) ForEach(do func(mauircdi.ChannelData)) {
	for _, val := range cdl {
		do(val)
	}
}
