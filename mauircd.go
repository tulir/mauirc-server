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
	"maunium.net/go/libmauirc"
	flag "maunium.net/go/mauflag"
	cfg "maunium.net/go/mauircd/config"
	"maunium.net/go/mauircd/database"
	"maunium.net/go/mauircd/ident"
	"maunium.net/go/mauircd/interfaces"
	"maunium.net/go/mauircd/web"
	log "maunium.net/go/maulogger"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

var confPath = flag.Make().LongKey("config").ShortKey("c").Default("/etc/mauircd/").Usage("The path to mauIRCd configurations").String()
var logPath = flag.Make().LongKey("logs").ShortKey("l").Default("/var/log/mauircd/").Usage("The path to mauIRCd logs").String()
var debug = flag.Make().LongKey("debug").ShortKey("d").Default("false").Usage("Use to enable debug prints").Bool()
var config mauircdi.Configuration
var version = "1.1.0"

func init() {
	libmauirc.Version = "mauIRCd " + version
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

	log.Infoln("Initializing mauIRCd", version)

	log.Debugln("Loading config from", *confPath)
	config = cfg.NewConfig(*confPath)
	err := config.Load()
	if err != nil {
		log.Fatalln("Failed to load config:", err)
		log.Close()
		os.Exit(1)
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

	log.Infoln("mauIRCd initialized. Connecting to IRC networks")
	config.Connect()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		log.Debugln("Now listening to interruptions and SIGTERMs")
		<-c
		go func() {
			time.Sleep(time.Second * 15)
			log.Fatalln("mauIRCd not closed within 15 seconds. Terminating.")
			log.Close()
			os.Exit(5)
		}()
		log.Infoln("Closing mauIRCd", version)
		config.GetUsers().ForEach(func(user mauircdi.User) {
			log.Debugln("Closing connections and saving scripts of", user.GetNameFromEmail())
			user.GetNetworks().ForEach(func(net mauircdi.Network) {
				net.Close()
				net.SaveScripts(config.GetPath())
				net.Save()
			})
		})
		log.Debugln("Waiting for all connections to close properly")
		time.Sleep(2 * time.Second)
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
