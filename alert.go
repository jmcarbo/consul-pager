package consul_pager

import (
	"fmt"
	//log "github.com/Sirupsen/logrus"
	"time"
)

type Alert struct {
	Id           string
	creationTime time.Time
	Subject      string
	Body         string
	Severity     string
	Sended       bool
}

func NewAlert(id string) *Alert {
	alert := &Alert{Id: id, creationTime: time.Now()}
	//log.Infof("Alert created %#v", alert)
	return alert
}

func (a *Alert) IsMature(duration string) bool {
	dur, _ := time.ParseDuration(duration)
	return time.Since(a.creationTime) >= dur
}

func (a *Alert) Message() string {
	return fmt.Sprintf("[%s] %s\n%s", a.Severity, a.Subject, a.Body)
}
