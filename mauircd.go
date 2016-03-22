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
	"time"
)

var nws = flag.StringP("config", "c", "/etc/mauircd/", "The path to mauIRCd configurations")

func main() {
	flag.Parse()

	err := config.Load(*nws)
	if err != nil {
		panic(err)
	}

	database.Load(fmt.Sprintf("%[1]s:%[2]s@tcp(%[3]s:%[4]d)/%[5]s",
		config.GetConfig().SQL.Username,
		config.GetConfig().SQL.Password,
		config.GetConfig().SQL.IP,
		config.GetConfig().SQL.Port,
		config.GetConfig().SQL.Database,
	))

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
		time.Sleep(2 * time.Second)
		database.Close()
		config.Save()
		os.Exit(0)
	}()
	web.Load(config.GetConfig().Address, config.GetConfig().IP, config.GetConfig().Port)
}
