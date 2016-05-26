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
	"os"
)

func (conf *configImpl) AddIdent(name, ip string, port int) error {
	file, err := os.OpenFile(conf.Ident.File, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(file, conf.Ident.Format, name, ip, port)
	return err
}

func (conf *configImpl) ClearIdent() error {
	os.Remove(conf.Ident.File)
	f, err := os.Create(conf.Ident.File)
	if err != nil {
		return err
	}
	return f.Chmod(0644)
}
