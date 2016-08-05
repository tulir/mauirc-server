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

// Package socket contains the WebSocket handlers
package socket

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"maunium.net/go/mauirc-server/interfaces"
	"maunium.net/go/mauirc-server/web/auth"
	"maunium.net/go/mauirc-server/web/errors"
	"maunium.net/go/mauirc-server/web/util"
	"maunium.net/go/maulogger"
	"net/http"
	"time"
)

var log = maulogger.CreateSublogger("Web/Socket", maulogger.LevelInfo)

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
	user interfaces.User
}

func (c *connection) readPump() {
	defer func() {
		c.ws.Close()
	}()
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Warnln("Unexpected close:", err)
			}
			break
		}

		var data = make(map[string]string)
		err = json.Unmarshal(message, &data)
		if err != nil {
			continue
		}
		c.user.HandleCommand(data)
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

func (c *connection) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()

	for {
		select {
		case new, ok := <-c.user.GetMessageChan():
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				c.user.GetMessageChan() <- new
				return
			}

			err := c.writeJSON(new)
			if err != nil {
				log.Debugln("Disconnected:", err)
				c.user.GetMessageChan() <- new
				return
			}
		case <-ticker.C:
			err := c.write(websocket.PingMessage, []byte{})
			if err != nil {
				return
			}
		}
	}
}

// Serve the WebSocket upgrader
func Serve(w http.ResponseWriter, r *http.Request) {
	success, user := auth.Check(w, r)
	if !success {
		errors.Write(w, errors.NotAuthenticated)
		log.Debugln(util.GetIP(r), "tried to connect to socket without authentication")
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Warnln("Failed to connect:", err)
		return
	}
	c := &connection{ws: ws, user: user}

	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	go c.writePump()

	c.user.GetNetworks().ForEach(func(net interfaces.Network) {
		c.user.SendNetworkData(net)
	})

	c.readPump()
}
