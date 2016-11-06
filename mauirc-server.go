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
package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"maunium.net/go/libmauirc"
	flag "maunium.net/go/mauflag"
	cfg "maunium.net/go/mauirc-server/config"
	"maunium.net/go/mauirc-server/database"
	"maunium.net/go/mauirc-server/ident"
	"maunium.net/go/mauirc-server/interfaces"
	"maunium.net/go/mauirc-server/web"
	log "maunium.net/go/maulogger"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

var confPath = flag.Make().LongKey("config").ShortKey("c").Default("/etc/mauirc/").Usage("The path to mauIRC server configurations").String()
var logPath = flag.Make().LongKey("logs").ShortKey("l").Default("/var/log/mauirc/").Usage("The path to mauIRC server logs").String()
var debug = flag.Make().LongKey("debug").ShortKey("d").Default("false").Usage("Use to enable debug prints").Bool()
var config interfaces.Configuration
var version = "2.0.0"

func init() {
	libmauirc.Version = "mauIRC Server " + version
}

func main() {
	flag.Parse()

	os.MkdirAll(*logPath, 0755)
	log.DefaultLogger.FileFormat = func(date string, i int) string {
		return filepath.Join(*logPath, fmt.Sprintf("%[1]s-%02[2]d.log", date, i))
	}
	log.OpenFile()
	if *debug {
		log.DefaultLogger.PrintLevel = 0
	}

	log.Infoln("Initializing mauIRC Server", version)

	log.Debugln("Loading config from", *confPath)
	config = cfg.NewConfig(*confPath)
	err := config.Load()
	if err != nil {
		log.Fatalln("Failed to load config:", err)
		log.Close()
		os.Exit(1)
	}

	if config.GetMail().IsEnabled() {
		config.GetMail().LoadTemplates(*confPath)
	}

	if config.GetIDENTConfig().Enabled {
		log.Debugln("Enabling the IDENTd")
		err = ident.Load(config.GetIDENTConfig())
		if err != nil {
			log.Fatalln("Failed to enable IDENTd:", err)
			log.Close()
			os.Exit(2)
		}
		go ident.Listen()
	}

	log.Debugln("Loading database with SQL string", config.GetSQLString())
	err = database.Load(config.GetSQLString())
	if err != nil {
		log.Fatalln("Failed to load database:", err)
		log.Close()
		os.Exit(3)
	}

	log.Infoln("mauIRC server initialized. Connecting to IRC networks")
	config.Connect()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		log.Debugln("Now listening to interruptions and SIGTERMs")
		<-c
		log.Infoln("Closing mauIRC server", version)
		config.GetUsers().ForEach(func(user interfaces.User) {
			log.Debugln("Closing connections and saving scripts of", user.GetNameFromEmail())
			user.GetNetworks().ForEach(func(net interfaces.Network) {
				net.Disconnect()
				net.SaveScripts(config.GetPath())
				net.Save()
			})
			user.SaveGlobalScripts(config.GetPath())
		})
		time.Sleep(1 * time.Second)
		log.Debugln("Closing the database connection")
		database.Close()
		log.Debugln("Saving config")
		config.Save()
		log.Debugln("Closing log")
		log.Close()
		os.Exit(0)
	}()

	web.Load(config)
}
