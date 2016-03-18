package irc

import (
	"fmt"

	goirc "github.com/thoj/go-ircevent"
	"maunium.net/go/mauircd/database"
)

// MauIRCon is a mauircd connection
type MauIRCon struct {
	IRC     *goirc.Connection
	User    string
	Network string
}

// Create an IRC connection
func Create(name, nick, user, password, ip string, port int, ssl bool) *MauIRCon {
	i := goirc.IRC(nick, user)

	i.UseTLS = ssl
	if len(password) > 0 {
		i.Password = password
	}
	i.Connect(fmt.Sprintf("%s:%d", ip, port))
	mauirc := &MauIRCon{IRC: i, User: user, Network: name}

	i.AddCallback("PRIVMSG", mauirc.privmsg)
	i.AddCallback("CTCP_ACTION", mauirc.action)

	return mauirc
}

func (i *MauIRCon) privmsg(event *goirc.Event) {
    database.
}

func (i *MauIRCon) action(event *goirc.Event) {

}
