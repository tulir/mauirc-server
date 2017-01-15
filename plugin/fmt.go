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

// Package plugin contains Lua plugin executing stuff
package plugin

import (
	pkg "fmt"
	"io"

	"github.com/mattn/anko/vm"
)

// Fmt is the importer for the fmt package with prints proxied
func Fmt(out io.Writer) func(env *vm.Env) *vm.Env {
	return func(env *vm.Env) *vm.Env {
		m := env.NewPackage("fmt")

		// If the plugin developer is intentionally trying to break things, allow it.
		m.Define("Fprint", pkg.Fprint)
		m.Define("Fprintf", pkg.Fprintf)
		m.Define("Fprintln", pkg.Fprintln)
		m.Define("Fscan", pkg.Fscan)
		m.Define("Fscanf", pkg.Fscanf)
		m.Define("Fscanln", pkg.Fscanln)

		// Proxied prints.
		m.Define("Print", func(a ...interface{}) (n int, err error) {
			return pkg.Fprint(out, a...)
		})
		m.Define("Printf", func(format string, a ...interface{}) (n int, err error) {
			return pkg.Fprintf(out, format, a...)
		})
		m.Define("Println", func(a ...interface{}) (n int, err error) {
			return pkg.Fprintln(out, a...)
		})

		// Direct user input not implemented.
		//m.Define("Scan", pkg.Scan)
		//m.Define("Scanf", pkg.Scanf)
		//m.Define("Scanln", pkg.Scanln)

		// String tools are just fine.
		m.Define("Sprint", pkg.Sprint)
		m.Define("Sprintf", pkg.Sprintf)
		m.Define("Sprintln", pkg.Sprintln)
		m.Define("Sscan", pkg.Sscan)
		m.Define("Sscanf", pkg.Sscanf)
		m.Define("Sscanln", pkg.Sscanln)
		m.Define("Errorf", pkg.Errorf)
		return m
	}
}
