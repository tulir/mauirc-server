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
	"github.com/Jeffail/gabs"
	"maunium.net/go/mauircd/database"
)

// HandleCommand handles mauIRC commands from clients
func (user User) HandleCommand(data *gabs.Container) {
	typ, ok := data.Path("type").Data().(string)
	if !ok {
		return
	}

	switch typ {
	case "message":
		user.cmdMessage(data)
	case "userlist":
		user.cmdUserlist(data)
	case "clearbuffer":
		user.cmdClearbuffer(data)
	case "deletemessage":
		// TODO: Delete Message
	case "importscript":
		// TODO: Import script
	}
}

// TODO: Implement deletemessage and importscript
/*case "deletemessage":
      if len(args) > 0 {
          id, err := strconv.Atoi(args[0])
          if err != nil {
              net.ReceiveMessage("*mauirc", "mauIRCd", "privmsg", "Couldn't parse int from "+args[0])
              return
          }
          database.DeleteMessage(net.Owner.Email, int64(id))
      }
  case "importscript":
      if len(args) > 1 {
          args[0] = strings.ToLower(args[0])
          data, err := download(args[1])
          if err != nil {
              fmt.Println(err)
              net.ReceiveMessage("*mauirc", "mauIRCd", "privmsg", "Failed to download script from http://pastebin.com/raw/"+args[1])
              return
          }
          for i := 0; i < len(net.Scripts); i++ {
              if net.Scripts[i].Name == args[0] {
                  net.Scripts[i].TheScript = data
                  net.ReceiveMessage("*mauirc", "mauIRCd", "privmsg", "Successfully updated script with name "+args[0])
                  return
              }
          }
          net.Scripts = append(net.Scripts, plugin.Script{TheScript: data, Name: args[0]})
          net.ReceiveMessage("*mauirc", "mauIRCd", "privmsg", "Successfully loaded script with name "+args[0])
      }
  }*/

type cmdResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (user User) cmdClearbuffer(data *gabs.Container) {
	network, ok := data.Path("network").Data().(string)
	if !ok {
		return
	}

	net := user.GetNetwork(network)
	if net == nil {
		user.NewMessages <- MauMessage{Type: "command-response", Object: cmdResponse{Success: false, Message: "No such network: " + network}}
		return
	}

	channel, ok := data.Path("channel").Data().(string)
	if !ok {
		return
	}
	err := database.ClearChannel(user.Email, net.Name, channel)
	if err != nil {
		user.NewMessages <- MauMessage{Type: "command-response", Object: cmdResponse{Success: false, Message: "Failed to clear buffer of " + net.Name}}
		return
	}
	user.NewMessages <- MauMessage{Type: "command-response", Object: cmdResponse{Success: false, Message: "Successfully cleared buffer of " + channel + " on " + net.Name}}
}

func (user User) cmdUserlist(data *gabs.Container) {
	network, ok := data.Path("network").Data().(string)
	if !ok {
		return
	}

	net := user.GetNetwork(network)
	if net == nil {
		return
	}

	channel, ok := data.Path("channel").Data().(string)
	if !ok {
		return
	}
	info := net.ChannelInfo[channel]
	if info == nil {
		return
	}
	user.NewMessages <- MauMessage{Type: "userlist", Object: info.UserList}
}

func (user User) cmdMessage(data *gabs.Container) {
	network, ok := data.Path("network").Data().(string)
	if !ok {
		return
	}

	net := user.GetNetwork(network)
	if net == nil {
		return
	}

	channel, okChan := data.Path("network").Data().(string)
	command, okCmd := data.Path("network").Data().(string)
	message, okMsg := data.Path("network").Data().(string)
	if !okChan || !okCmd || !okMsg {
		return
	}

	net.SendMessage(channel, command, message)
}
