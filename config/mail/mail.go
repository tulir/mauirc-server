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

// Package mail contains mail configs
package mail

import (
	"bytes"
	"fmt"
	"net/smtp"
	"strings"
)

// Config contains mail sending instructions.
type Config struct {
	Enabled bool              `json:"enabled"`
	Mode    string            `json:"mode"`
	Config  map[string]string `json:"config"`
}

// Validate the config.
func (mail Config) Validate() error {
	mail.Mode = strings.ToLower(mail.Mode)
	switch mail.Mode {
	case "smtp":
		_, ok := mail.Config["host"]
		if !ok {
			return fmt.Errorf("SMTP host not given")
		}
		_, ok = mail.Config["sender"]
		if !ok {
			return fmt.Errorf("SMTP sender not given")
		}
	case "sendmail":
		_, ok := mail.Config["binary"]
		if !ok {
			return fmt.Errorf("Sendmail binary not given")
		}
		return fmt.Errorf("Sendmail mailing not yet implemented")
	}
	return nil
}

// IsEnabled returns whether or not the mailing system is enabled.
func (mail Config) IsEnabled() bool {
	return mail.Enabled
}

// Send mail.
func (mail Config) Send(to, subject, template string, args map[string]interface{}) {
	switch mail.Mode {
	case "smtp":
		host := mail.Config["host"]
		sender := mail.Config["sender"]
		user, useAuth := mail.Config["username"]
		password, _ := mail.Config["password"]

		var auth smtp.Auth
		if useAuth {
			auth = smtp.PlainAuth("", user, password, host)
		}

		var buf bytes.Buffer
		buf.WriteString("From: ")
		buf.WriteString(sender)
		buf.WriteString("\n")
		buf.WriteString("To: ")
		buf.WriteString(to)
		buf.WriteString("\n")
		buf.WriteString("Subject: ")
		buf.WriteString(subject)
		buf.WriteString("\n\n")
		tmpl.ExecuteTemplate(&buf, template, args)

		smtp.SendMail(host, auth, sender, []string{to}, bytes.Replace([]byte{'\n'}, buf.Bytes(), []byte{'\n', '\r'}, -1))
	}
}
