package irc

import (
	"fmt"

	goirc "github.com/thoj/go-ircevent"
	"maunium.net/go/mauircd/database"
	"maunium.net/go/mauircd/plugin"
)

// Network is a mauircd network connection
type Network struct {
	IRC     *goirc.Connection
	Owner   string
	Name    string
	Scripts []plugin.Script
}

// Create an IRC connection
func Create(name, nick, user, password, ip string, port int, ssl bool) *Network {
	i := goirc.IRC(nick, user)

	i.UseTLS = ssl
	if len(password) > 0 {
		i.Password = password
	}
	i.Connect(fmt.Sprintf("%s:%d", ip, port))
	mauirc := &Network{IRC: i, Owner: user, Name: name}

	i.AddCallback("PRIVMSG", mauirc.privmsg)
	i.AddCallback("CTCP_ACTION", mauirc.action)

	return mauirc
}

func (net *Network) message(channel, sender, command, message string) {
	for _, s := range net.Scripts {
		channel, sender, command, message = s.Run(channel, sender, command, message)
	}

	database.Insert(net.Owner, net.Name, channel, sender, command, message)
}

func (net *Network) privmsg(evt *goirc.Event) {
	net.message(evt.Arguments[0], evt.Nick, "privmsg", evt.Message())
}

func (net *Network) action(evt *goirc.Event) {
	net.message(evt.Arguments[0], evt.Nick, "action", evt.Message())
}
