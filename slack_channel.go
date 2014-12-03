package consul_pager

import (
	log "github.com/Sirupsen/logrus"
	"github.com/jmcarbo/slackcli/slack"
)

type SlackChannel struct {
	id        string
	Subdomain string
	Token     string
	Channel   string
	User      string
}

func NewSlackChannel(subdomain, token, channel, user string) *SlackChannel {
	sc := &SlackChannel{Subdomain: subdomain, Token: token, Channel: channel, User: user}
	return sc
}

func (sc *SlackChannel) Id() string {
	return sc.id
}

func (sc *SlackChannel) SetId(id string) {
	sc.id = id
}

func (sc *SlackChannel) Send(alert *Alert) error {
	log.Infof("Sending alert %s to  %s", alert.Message(), sc.id)
	payload := slack.BuildPayload(sc.Channel, sc.User, "", ":ghost:", alert.Message())
	slack.Post(sc.Subdomain, sc.Token, payload)
	return nil
}
