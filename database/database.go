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
	"errors"
)

// Message wraps an IRC message
type Message struct {
	ID        int64  `json:"id"`
	Network   string `json:"network"`
	Channel   string `json:"channel"`
	Timestamp int64  `json:"timestamp"`
	Sender    string `json:"sender"`
	Command   string `json:"command"`
	Message   string `json:"message"`
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
		"message TEXT NOT NULL" +
		") DEFAULT CHARSET=utf8;")
	return err
}

// Close the database connection
func Close() {
	db.Close()
}

// GetHistory gets the last n messages
func GetHistory(email string, n int) ([]Message, error) {
	results, err := db.Query("SELECT id, network, channel, timestamp, sender, command, message FROM messages WHERE email=? ORDER BY id DESC LIMIT ?", email, n)
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

		var network, channel, sender, command, message string
		var timestamp, id int64

		results.Scan(&id, &network, &channel, &timestamp, &sender, &command, &message)

		messages = append(messages, Message{
			ID:        id,
			Network:   network,
			Channel:   channel,
			Timestamp: timestamp,
			Sender:    sender,
			Command:   command,
			Message:   message,
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
func Insert(email string, msg Message) (int64, error) {
	result := db.QueryRow("INSERT INTO messages (email, network, channel, timestamp, sender, command, message) VALUES (?, ?, ?, ?, ?, ?, ?); SELECT LAST_INSERT_ID();",
		email, msg.Network, msg.Channel, msg.Timestamp, msg.Sender, msg.Command, msg.Message)
	var id int64
	err := result.Scan(id)
	return id, err
}
