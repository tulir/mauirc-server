package db

import (
	"database/sql"
	"fmt"
	"time"
)

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
		"user VARCHAR(255) NOT NULL," +
		"network VARCHAR(255) NOT NULL," +
		"channel VARCHAR(255) NOT NULL," +
		"timestamp BIGINT NOT NULL," +
		"sender VARCHAR(255) NOT NULL," +
		"action VARCHAR(255) NOT NULL," +
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

// Insert a message into the database
func Insert(user, network, channel, sender, action, message string) error {
	_, err := db.Exec("INSERT INTO messages (user, network, channel, timestamp, sender, action, message) VALUES (?, ?, ?, ?, ?, ?, ?);",
		user, network, channel, time.Now().Unix(), sender, action, message)
	return err
}
