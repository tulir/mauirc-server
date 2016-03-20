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
	"fmt"
	"time"
)

// Message wraps an IRC message
type Message struct {
	Network   string `json:"network"`
	Channel   string `json:"channel"`
	Timestamp int64  `json:"timestamp"`
	Sender    string `json:"sender"`
	Command   string `json:"command"`
	Message   string `json:"message"`
}

var db *sql.DB

// Load the database
func Load(username, password, ip string, port int, database string) error {
	var err error
	db, err = sql.Open("mysql", fmt.Sprintf("%[1]s:%[2]s@tcp(%[3]s:%[4]d)/%[5]s", username, password, ip, port, database))
	if err != nil {
		return err
	} else if db == nil {
		return fmt.Errorf("Failed to open SQL connection!")
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
		");")
	if err != nil {
		return err
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS users (" +
		"email VARCHAR(255) PRIMARY KEY," +
		"password BINARY(60) NOT NULL," +
		"lastfetch BIGINT NOT NULL," +
		"user VARCHAR(255) NOT NULL," +
		"nick VARCHAR(255) NOT NULL," +
		"realname VARCHAR(255) NOT NULL" +
		");")
	if err != nil {
		return err
	}
	return nil
}

// Close the database connection
func Close() {
	db.Close()
}

// GetUnread gets all the unread messages of the given user
func GetUnread(email string) ([]Message, error) {
	result := db.QueryRow("SELECT lastfetch FROM users WHERE email=?", email)
	var lastfetch int64
	result.Scan(&lastfetch)

	results, err := db.Query("SELECT network, channel, timestamp, sender, command, message FROM messages WHERE email=? AND timestamp>?", email, lastfetch)
	if err != nil {
		return nil, err
	}
	messages, err := scanMessages(results)
	if err != nil {
		return messages, err
	}

	db.Exec("UPDATE users SET lastfetch=? WHERE email=?", time.Now().Unix(), email)

	return messages, nil
}

// GetHistory gets the last n messages
func GetHistory(email string, n int) ([]Message, error) {
	results, err := db.Query("SELECT network, channel, timestamp, sender, command, message FROM messages WHERE email=? ORDER BY id DESC LIMIT ?", email, n)
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
		var timestamp int64

		results.Scan(&network, &channel, &timestamp, &sender, &command, &message)

		messages = append(messages, Message{
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

// Insert a message into the database
func Insert(email string, msg Message) error {
	_, err := db.Exec("INSERT INTO messages (email, network, channel, timestamp, sender, command, message) VALUES (?, ?, ?, ?, ?, ?, ?);",
		email, msg.Network, msg.Channel, msg.Timestamp, msg.Sender, msg.Command, msg.Message)
	return err
}
