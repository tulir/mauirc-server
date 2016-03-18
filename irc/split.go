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

// Package irc contains the IRC client
package irc

import (
	"strings"
)

func split(message string) []string {
	if strings.ContainsRune(message, '\n') {
		return strings.Split(message, "\n")
	} else if len(message) > 250 {
		return splitLen(message)
	} else {
		return nil
	}
}

func splitLen(message string) []string {
	lastIndex := -1
	for i := 0; i < 250; i++ {
		if message[i] == ' ' {
			lastIndex = i
		}
	}

	if lastIndex == -1 {
		for i := 0; i < 250; i++ {
			if message[i] == '-' || message[i] == '.' || message[i] == ',' {
				lastIndex = i
			}
		}
	} else {
		return []string{message[:lastIndex], message[lastIndex+1:]}
	}

	if lastIndex != -1 {
		return []string{message[:lastIndex+1], message[lastIndex+1:]}
	}

	return []string{message[:250], message[250:]}
}
