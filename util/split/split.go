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

// Package split helps with splitting messages properly
package split

import (
	"strings"
)

// All splits a message by newlines and if the message is longer than 250
// characters, split it into smaller pieces using SplitLen
func All(message string) []string {
	splitted := []string{message}
	if strings.ContainsRune(message, '\n') {
		splitted = strings.Split(message, "\n")
	} else if len(message) > 250 {
		for len(splitted[len(splitted)-1]) > 250 {
			if len(splitted) < 2 {
				a, b := ByLength(splitted[0])
				if len(b) != 0 {
					splitted = []string{a, b}
				} else {
					splitted = []string{a}
				}
			} else {
				a, b := ByLength(splitted[len(splitted)-1])
				splitted[len(splitted)-1] = a
				if len(b) != 0 {
					splitted = append(splitted, b)
				}
			}
		}
	}
	return splitted
}

// ByLength splits a message into pieces that are less than 250 characters long.
// If the message contains a space character before the character limit,
func ByLength(message string) (string, string) {
	if len(message) < 250 {
		return message, ""
	}
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
		return message[:lastIndex], message[lastIndex+1:]
	}

	if lastIndex != -1 {
		return message[:lastIndex+1], message[lastIndex+1:]
	}

	return message[:250], message[250:]
}
