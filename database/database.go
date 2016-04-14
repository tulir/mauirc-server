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

// Package database contains the database systems
package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"maunium.net/go/mauircd/util/preview"
)

// Message wraps an IRC message
type Message struct {
	ID        int64            `json:"id"`
	Network   string           `json:"network"`
	Channel   string           `json:"channel"`
	Timestamp int64            `json:"timestamp"`
	Sender    string           `json:"sender"`
	Command   string           `json:"command"`
	Message   string           `json:"message"`
	OwnMsg    bool             `json:"ownmsg"`
	Preview   *preview.Preview `json:"preview"`
}

var db *sql.DB

// Load the database
func Load(sqlStr string) error {
	var err error
	db, err = sql.Open("mysql", sqlStr)
	if err != nil {
		return err
	} else if db == nil {
		return errors.New("Failed to open SQL connection!")
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS messages (" +
		"id BIGINT PRIMARY KEY AUTO_INCREMENT," +
		"email VARCHAR(255) NOT NULL," +
		"network VARCHAR(255) NOT NULL," +
		"channel VARCHAR(255) NOT NULL," +
		"timestamp BIGINT NOT NULL," +
		"sender VARCHAR(255) NOT NULL," +
		"command VARCHAR(255) NOT NULL," +
		"message TEXT NOT NULL," +
		"ownmessage TINYINT(1) NOT NULL," +
		"preview TEXT" +
		") DEFAULT CHARSET=utf8;")
	return err
}

// Close the database connection
func Close() {
	db.Close()
}

// GetHistory gets the last n messages
func GetHistory(email string, n int) ([]Message, error) {
	results, err := db.Query("SELECT id, network, channel, timestamp, sender, command, message, ownmessage, preview FROM messages WHERE email=? ORDER BY id DESC LIMIT ?", email, n)
	if err != nil {
		return nil, err
	}
	return scanMessages(results)
}

// GetNetworkHistory gets the last n messages on the given network
func GetNetworkHistory(email, network string, n int) ([]Message, error) {
	results, err := db.Query("SELECT id, network, channel, timestamp, sender, command, message, ownmessage, preview FROM messages WHERE email=? AND network=? ORDER BY id DESC LIMIT ?", email, network, n)
	if err != nil {
		return nil, err
	}
	return scanMessages(results)
}

// GetChannelHistory gets the last n messages on the given channel
func GetChannelHistory(email, network, channel string, n int) ([]Message, error) {
	results, err := db.Query("SELECT id, network, channel, timestamp, sender, command, message, ownmessage, preview FROM messages WHERE email=? AND network=? AND channel=? ORDER BY id DESC LIMIT ?", email, network, channel, n)
	if err != nil {
		return nil, err
	}
	return scanMessages(results)
}

func scanMessages(results *sql.Rows) ([]Message, error) {
	var messages []Message
	for results.Next() {
		if results.Err() != nil {
			return messages, results.Err()
		}

		var network, channel, sender, command, message, previewStr string
		var ownmessage bool
		var timestamp, id int64

		results.Scan(&id, &network, &channel, &timestamp, &sender, &command, &message, &ownmessage, &previewStr)

		var pw = &preview.Preview{}
		if len(previewStr) > 0 {
			json.Unmarshal([]byte(previewStr), pw)
		} else {
			pw = nil
		}

		messages = append(messages, Message{
			ID:        id,
			Network:   network,
			Channel:   channel,
			Timestamp: timestamp,
			Sender:    sender,
			Command:   command,
			Message:   message,
			OwnMsg:    ownmessage,
			Preview:   pw,
		})
	}
	return messages, nil
}

// DeleteMessage deletes the message with the given ID
func DeleteMessage(email string, id int64) error {
	_, err := db.Exec("DELETE FROM messages WHERE email=? AND id=?;", email, id)
	return err
}

// ClearChannel clears all the messages in the given channel.
func ClearChannel(email, network, channel string) error {
	_, err := db.Exec("DELETE FROM messages WHERE email=? AND network=? AND channel=?;", email, network, channel)
	return err
}

// ClearNetwork clears all the messages in the channels that are in the given network.
func ClearNetwork(email, network string) error {
	_, err := db.Exec("DELETE FROM messages WHERE email=? AND network=?;", email, network)
	return err
}

// ClearUser clears all messages owned by the given user.
func ClearUser(email string) error {
	_, err := db.Exec("DELETE FROM messages WHERE email=?;", email)
	return err
}

// Insert a message into the database
func Insert(email string, msg Message) int64 {
	var preview = ""
	if msg.Preview != nil {
		data, err := json.Marshal(msg.Preview)
		if err == nil {
			preview = string(data)
		}
	}
	db.Exec("INSERT INTO messages (email, network, channel, timestamp, sender, command, message, ownmessage, preview) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);",
		email, msg.Network, msg.Channel, msg.Timestamp, msg.Sender, msg.Command, msg.Message, msg.OwnMsg, preview)

	result := db.QueryRow("SELECT id FROM messages WHERE email=? AND network=? AND channel=? AND timestamp=? AND sender=? AND command=? AND message=? AND ownmessage=?;",
		email, msg.Network, msg.Channel, msg.Timestamp, msg.Sender, msg.Command, msg.Message, msg.OwnMsg)
	var id int64
	result.Scan(&id)
	return id
}
