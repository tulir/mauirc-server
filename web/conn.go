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

// Package web contains the HTTP server
package web

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"maunium.net/go/mauircd/config"
	"maunium.net/go/mauircd/database"
	"net/http"
	"time"
)

type sendform struct {
	Network string `json:"network"`
	Channel string `json:"channel"`
	Command string `json:"command"`
	Message string `json:"message"`
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 30 * time.Second
	pingPeriod     = 25 * time.Second
	maxMessageSize = 1024
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
}

type connection struct {
	ws   *websocket.Conn
	user *config.User
}

func (c *connection) readPump() {
	defer func() {
		c.ws.Close()
	}()
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				fmt.Println("Unexpected close:", err)
			}
			break
		}

		var sf sendform
		err = json.Unmarshal(message, &sf)
		if err != nil || len(sf.Network) == 0 || len(sf.Channel) == 0 || len(sf.Command) == 0 || len(sf.Message) == 0 {
			continue
		}

		c.user.GetNetwork(sf.Network).SendMessage(sf.Channel, sf.Command, sf.Message)
	}
}

func (c *connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

func (c *connection) writeJSON(payload interface{}) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteJSON(payload)
}

type aggregate struct {
	val    database.Message
	source *config.Network
}

func (c *connection) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()
	agg := make(chan aggregate)
	for i := 0; i < len(c.user.Networks); i++ {
		var ii = i
		go func() {
			ch := c.user.Networks[ii]
			for val := range ch.NewMessages {
				agg <- aggregate{val, ch}
			}
		}()
	}

	for {
		select {
		case new, ok := <-agg:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
			} else if err := c.writeJSON(new.val); err != nil {
				fmt.Println("Disconnected:", err)
			} else {
				continue
			}
			new.source.NewMessages <- new.val
			return
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

/*func (c *connection) waitAuth() bool {
	_, message, err := c.ws.ReadMessage()
	if err != nil {
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
			fmt.Println("Unexpected close:", err)
		}
		return false
	}

	var af authform
	err = json.Unmarshal(message, &af)
	if err != nil || len(af.Email) == 0 || len(af.Password) == 0 {
		c.write(websocket.TextMessage, []byte("{\"auth\": true, \"success\": false}"))
		c.write(websocket.CloseMessage, []byte{})
		return false
	}

	user := config.GetUser(af.Email)
	if user.CheckPassword(af.Password) {
		c.user = user
		c.write(websocket.TextMessage, []byte("{\"auth\": true, \"success\": true}"))
		return true
	}

	c.write(websocket.TextMessage, []byte("{\"auth\": true, \"success\": false}"))
	c.write(websocket.CloseMessage, []byte{})
	return false
}*/

func serveWs(w http.ResponseWriter, r *http.Request) {
	success, user := checkAuth(w, r)
	if !success {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println("Auth fail")
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Failed to connect:", err)
		return
	}
	c := &connection{ws: ws, user: user}

	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	/*if !c.waitAuth() {
		return
	}*/

	go c.writePump()
	c.readPump()
}
