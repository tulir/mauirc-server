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

// Package ident contains an RFC 1413 implementation
package ident

import (
	"bufio"
	"fmt"
	mauircdi "maunium.net/go/mauircd/interfaces"
	"net"
	"strconv"
	"strings"
)

// Ports is the port->name mapping
var Ports = make(map[int]string)
var ln net.Listener

// Load the IDENTd
func Load(config mauircdi.IdentConf) error {
	var ipport = fmt.Sprintf("%s:%d", config.IP, config.Port)
	var err error
	ln, err = net.Listen("tcp", ipport)
	if err != nil {
		return err
	}
	return nil
}

// Listen to incoming connections
func Listen() {
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("[IDENT] Failed connection from", conn.RemoteAddr().String())
			continue
		}
		//fmt.Println("[IDENT] IDENT connection from", conn.RemoteAddr().String())
		go handleConn(conn)
	}
}

func handleConn(socket net.Conn) {
	defer socket.Close()

	br := bufio.NewReaderSize(socket, 512)
	msg, err := br.ReadString('\n')
	if err != nil {
		fmt.Printf("[IDENT] Failed to read from %s: %s\n", socket.RemoteAddr().String(), err)
		return
	} else if len(msg) > 20 {
		return
	}

	reqParts := strings.Split(strings.TrimSpace(msg), ", ")
	if len(reqParts) != 2 {
		return
	}

	localPort, err := strconv.Atoi(reqParts[0])
	if err != nil {
		return
	}

	remotePort, err := strconv.Atoi(reqParts[1])
	if err != nil {
		return
	}

	name, ok := Ports[localPort]
	if !ok {
		fmt.Fprintf(socket, "%d, %d : ERROR : NO-USER\r\n", localPort, remotePort)
	} else {
		fmt.Fprintf(socket, "%d, %d : USERID : UNIX : %s\r\n", localPort, remotePort, name)
	}
}
