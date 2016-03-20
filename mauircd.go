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
package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	flag "github.com/ogier/pflag"
	"maunium.net/go/mauircd/config"
	"maunium.net/go/mauircd/database"
	"maunium.net/go/mauircd/web"
	"os"
	"os/signal"
	"syscall"
)

var nws = flag.StringP("config", "c", "/etc/mauircd/", "The path to mauIRCd configurations")

func main() {
	flag.Parse()

	err := config.Load(*nws)
	if err != nil {
		panic(err)
	}

	database.Load("root", flag.Arg(0), "127.0.0.1", 3306, "mauircd")
	web.Load("127.0.0.1", 29304)
	//irc.Create("pvlnet", "mauircd", "mauircd", "mauircd@maunium.net", "", "irc.fixme.fi", 6697, true)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nClosing mauIRCd")
		for _, user := range config.GetUsers() {
			for _, network := range user.Networks {
				network.Close()
			}
		}
		database.Close()
		os.Exit(0)
	}()
}
