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
	"strings"
)

type mailConfig struct {
	Enabled bool              `json:"enabled"`
	Mode    string            `json:"mode"`
	Config  map[string]string `json:"config"`
}

func (mail *mailConfig) Validate() error {
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
	}
	return nil
}

func (mail *mailConfig) Send() {

}
