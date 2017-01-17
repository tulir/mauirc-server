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
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"

	"maunium.net/go/maulogger"
)

var log = maulogger.CreateSublogger("Config/Mail", maulogger.LevelInfo)
var tmpl = template.New("root")

// LoadTemplates loads all templates from the template directory.
func (mail Config) LoadTemplates(path string) {
	path = filepath.Join(path, "templates")
	log.Debugln("Loading templates from", path)

	files, _ := ioutil.ReadDir(path)
	for _, f := range files {
		name := f.Name()
		nameParts := strings.Split(f.Name(), ".")
		if len(nameParts) > 1 {
			name = strings.Join(nameParts[:len(nameParts)-1], ".")
		}

		data, err := ioutil.ReadFile(filepath.Join(path, f.Name()))
		if err != nil {
			log.Errorf("Failed to read templates/%s: %s\n", f.Name(), err)
			continue
		}

		_, err = tmpl.New(name).Parse(string(data))
		if err != nil {
			log.Errorf("Failed to parse templates/%s: %s\n", f.Name(), err)
			continue
		}
		log.Debugln("Successfully parsed template %s", f.Name())
	}
}
