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

// Package messages contains mauIRC client <-> server messages
package messages

import (
	"encoding/json"
	"strconv"
)

// Message types
const (
	MsgRaw        = "raw"
	MsgInvite     = "invite"
	MsgNickChange = "nickchange"
	MsgNetData    = "netdata"
	MsgChanData   = "chandata"
	MsgWhois      = "whois"
	MsgClear      = "clear"
	MsgDelete     = "delete"
	MsgChanList   = "chanlist"
	MsgMessage    = "message"
	MsgKick       = "kick"
	MsgMode       = "mode"
	MsgClose      = "close"
	MsgOpen       = "open"
)

// Container is a basic wrapper for a type string and the actual message object
type Container struct {
	Type   string      `json:"type"`
	Object interface{} `json:"object"`
}

// Message wraps an IRC message
type Message struct {
	ID        int64    `json:"id,omitempty"`
	Network   string   `json:"network"`
	Channel   string   `json:"channel"`
	Timestamp int64    `json:"timestamp,omitempty"`
	Sender    string   `json:"sender,omitempty"`
	Command   string   `json:"command"`
	Message   string   `json:"message"`
	OwnMsg    bool     `json:"ownmsg,omitempty"`
	Preview   *Preview `json:"preview,omitempty"`
}

// ParseMessage parses a Message object from a generic object
func ParseMessage(obj interface{}) (msg Message) {
	mp, ok := obj.(map[string]interface{})
	if !ok {
		return
	}

	id, _ := mp["id"].(json.Number)
	msg.ID, _ = strconv.ParseInt(string(id), 10, 64)
	timestamp, _ := mp["timestamp"].(json.Number)
	msg.Timestamp, _ = strconv.ParseInt(string(timestamp), 10, 64)
	msg.Network, _ = mp["network"].(string)
	msg.Channel, _ = mp["channel"].(string)
	msg.Sender, _ = mp["sender"].(string)
	msg.Command, _ = mp["command"].(string)
	msg.Message, _ = mp["message"].(string)
	msg.OwnMsg, _ = mp["ownmsg"].(bool)
	pw, ok := mp["preview"]
	if ok {
		msg.Preview = ParsePreview(pw)
	}
	return
}

// RawMessage is a raw IRC message
type RawMessage struct {
	Network string `json:"network"`
	Message string `json:"message"`
}

// ParseRawMessage parses a RawMessage object from a generic object
func ParseRawMessage(obj interface{}) (msg RawMessage) {
	mp, ok := obj.(map[string]interface{})
	if !ok {
		return
	}

	msg.Network, _ = mp["network"].(string)
	msg.Message, _ = mp["message"].(string)
	return
}

// NickChange is the message for IRC nick changes
type NickChange struct {
	Network string `json:"network"`
	Nick    string `json:"nick"`
}

// ParseNickChange parses a NickChange object from a generic object
func ParseNickChange(obj interface{}) (msg NickChange) {
	mp, ok := obj.(map[string]interface{})
	if !ok {
		return
	}

	msg.Network, _ = mp["network"].(string)
	msg.Nick, _ = mp["nick"].(string)
	return
}

// NetData contains basic network data
type NetData struct {
	Name      string `json:"name"`
	User      string `json:"user"`
	Realname  string `json:"realname"`
	Nick      string `json:"nick"`
	IP        string `json:"ip"`
	Port      uint16 `json:"port"`
	SSL       bool   `json:"ssl"`
	Connected bool   `json:"connected"`
}

// ParseNetData parses a NetData object from a generic object
func ParseNetData(obj interface{}) (msg NetData) {
	mp, ok := obj.(map[string]interface{})
	if !ok {
		return
	}

	msg.Name, _ = mp["name"].(string)
	msg.User, _ = mp["user"].(string)
	msg.Realname, _ = mp["realname"].(string)
	msg.Nick, _ = mp["nick"].(string)
	msg.IP, _ = mp["ip"].(string)
	port, _ := mp["port"].(json.Number)
	portuint64, _ := strconv.ParseUint(string(port), 10, 16)
	msg.Port = uint16(portuint64)
	msg.SSL, _ = mp["ssl"].(bool)
	msg.Connected, _ = mp["connected"].(bool)
	return
}

// ChanList contains a channel list and network name
type ChanList struct {
	Network string   `json:"network"`
	List    []string `json:"list"`
}

// ParseChanList parses a ChanList object from a generic object
func ParseChanList(obj interface{}) (msg ChanList) {
	mp, ok := obj.(map[string]interface{})
	if !ok {
		return
	}

	msg.Network, _ = mp["network"].(string)
	list, _ := mp["list"].([]interface{})
	msg.List = make([]string, len(list))
	for i, lo := range list {
		msg.List[i], _ = lo.(string)
	}
	return
}

// Invite an user to a channel
type Invite struct {
	Network string `json:"network"`
	Channel string `json:"channel"`
	Sender  string `json:"sender"`
}

// ParseInvite parses a Invite object from a generic object
func ParseInvite(obj interface{}) (msg Invite) {
	mp, ok := obj.(map[string]interface{})
	if !ok {
		return
	}

	msg.Network, _ = mp["network"].(string)
	msg.Channel, _ = mp["channel"].(string)
	msg.Sender, _ = mp["sender"].(string)
	return
}

// ClearHistory tells the client to clear the specific channel
type ClearHistory struct {
	Network string `json:"network"`
	Channel string `json:"channel"`
}

// ParseClearHistory parses a ClearHistory object from a generic object
func ParseClearHistory(obj interface{}) (msg ClearHistory) {
	mp, ok := obj.(map[string]interface{})
	if !ok {
		return
	}

	msg.Network, _ = mp["network"].(string)
	msg.Channel, _ = mp["channel"].(string)
	return
}

// WhoisData contains WHOIS information
type WhoisData struct {
	Channels   map[string]string `json:"channels"`
	Nick       string            `json:"nick"`
	User       string            `json:"user"`
	Host       string            `json:"host"`
	RealName   string            `json:"realname"`
	Away       string            `json:"away"`
	Server     string            `json:"server"`
	ServerInfo string            `json:"server-info"`
	IdleTime   int64             `json:"idle"`
	Idle       string            `json:"-"`
	SecureConn bool              `json:"secure-connection"`
	Operator   bool              `json:"operator"`
}

// ParseWhoisData parses a WhoisData object from a generic object
func ParseWhoisData(obj interface{}) (msg WhoisData) {
	mp, ok := obj.(map[string]interface{})
	if !ok {
		return
	}

	chlist, _ := mp["channels"].(map[string]interface{})
	msg.Channels = make(map[string]string)
	for ch, role := range chlist {
		msg.Channels[ch], _ = role.(string)
	}

	msg.Nick, _ = mp["nick"].(string)
	msg.User, _ = mp["user"].(string)
	msg.Host, _ = mp["host"].(string)
	msg.RealName, _ = mp["realname"].(string)
	msg.Away, _ = mp["away"].(string)
	msg.Server, _ = mp["server"].(string)
	msg.ServerInfo, _ = mp["server-info"].(string)
	idle, _ := mp["idle"].(json.Number)
	msg.IdleTime, _ = strconv.ParseInt(string(idle), 10, 64)
	msg.SecureConn, _ = mp["secure-connection"].(bool)
	msg.Operator, _ = mp["operator"].(bool)
	return
}

// DeleteMessage contains information about which message to delete
type DeleteMessage int64

// ParseDeleteMessage parses a DeleteMessage object from a generic object
func ParseDeleteMessage(obj interface{}) (msg DeleteMessage) {
	switch v := obj.(type) {
	case json.Number:
		i, _ := strconv.ParseInt(string(v), 10, 64)
		return DeleteMessage(i)
	case int:
		return DeleteMessage(v)
	case int64:
		return DeleteMessage(v)
	default:
		return 0
	}
}

// OpenCloseChannel contains information about which channel to close
type OpenCloseChannel struct {
	Network string `json:"network"`
	Channel string `json:"channel"`
}

// ParseOpenCloseChannel parses a OpenCloseChannel object from a generic object
func ParseOpenCloseChannel(obj interface{}) (msg OpenCloseChannel) {
	mp, ok := obj.(map[string]interface{})
	if !ok {
		return
	}

	msg.Network, _ = mp["network"].(string)
	msg.Channel, _ = mp["channel"].(string)
	return
}

// Kick contains information about who to kick
type Kick struct {
	Network string `json:"network"`
	Channel string `json:"channel"`
	User    string `json:"user"`
	Message string `json:"message"`
}

// ParseKick parses a Kick object from a generic object
func ParseKick(obj interface{}) (msg Kick) {
	mp, ok := obj.(map[string]interface{})
	if !ok {
		return
	}

	msg.Network, _ = mp["network"].(string)
	msg.Channel, _ = mp["channel"].(string)
	msg.User, _ = mp["users"].(string)
	msg.Message, _ = mp["message"].(string)
	return
}

// Mode contains information about who to kick
type Mode struct {
	Network string `json:"network"`
	Channel string `json:"channel"`
	Message string `json:"message"`
	Args    string `json:"args"`
}

// ParseMode parses a Mode object from a generic object
func ParseMode(obj interface{}) (msg Mode) {
	mp, ok := obj.(map[string]interface{})
	if !ok {
		return
	}

	msg.Network, _ = mp["network"].(string)
	msg.Channel, _ = mp["channel"].(string)
	msg.Message, _ = mp["message"].(string)
	msg.Args, _ = mp["args"].(string)
	return
}

// ModelistEntry contains a mode and a target
type ModelistEntry struct {
	Mode   rune   `json:"mode"`
	Target string `json:"target"`
}

// ParseModelistEntry parses a ModelistEntry object from a generic object
func ParseModelistEntry(obj interface{}) (msg ModelistEntry) {
	mp, ok := obj.(map[string]interface{})
	if !ok {
		return
	}

	mode, _ := mp["mode"].(string)
	if len(mode) > 0 {
		msg.Mode = rune(mode[0])
	}

	msg.Target, _ = mp["target"].(string)
	return
}

// ChanData contains channel information
type ChanData struct {
	Network    string          `json:"network"`
	Name       string          `json:"name"`
	Userlist   []string        `json:"userlist"`
	Topic      string          `json:"topic"`
	TopicSetBy string          `json:"topicsetby"`
	TopicSetAt int64           `json:"topicsetat"`
	Modelist   []ModelistEntry `json:"modes"`
}

// ParseChanData parses a ChanData object from a generic object
func ParseChanData(obj interface{}) (msg ChanData) {
	mp, ok := obj.(map[string]interface{})
	if !ok {
		return
	}

	msg.Network, _ = mp["network"].(string)
	msg.Name, _ = mp["name"].(string)

	list, _ := mp["userlist"].([]interface{})
	msg.Userlist = make([]string, len(list))
	for i, lo := range list {
		msg.Userlist[i], _ = lo.(string)
	}

	msg.Topic, _ = mp["topic"].(string)
	msg.TopicSetBy, _ = mp["topicsetby"].(string)
	topicsetat, _ := mp["topicsetat"].(json.Number)
	msg.TopicSetAt, _ = strconv.ParseInt(string(topicsetat), 10, 64)

	ml, _ := mp["modes"].([]interface{})
	msg.Modelist = make([]ModelistEntry, len(ml))
	for i, mod := range ml {
		msg.Modelist[i] = ParseModelistEntry(mod)
	}

	return
}
