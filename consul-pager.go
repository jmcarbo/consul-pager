package consul_pager

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/armon/consul-api"
	apixtra "github.com/jmcarbo/consul-apixtra"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"time"
)

//Connect establishes a connection to local running consul agent.
func Connect(address, datacenter, token string) *consulapi.Client {
	config := consulapi.DefaultConfig()
	if address != "" {
		config.Address = address
	}
	if datacenter != "" {
		config.Datacenter = datacenter
	}
	if token != "" {
		config.Token = token
	}
	client, err := consulapi.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	return client
}

type ConsulPager struct {
	alerts          map[string]*Alert
	client          *consulapi.Client
	kvlock          *apixtra.Lock
	Channels        map[string]Channel
	DefaultMaturity string
	Events          map[string]*consulapi.UserEvent
}

func NewConsulPager(client *consulapi.Client) *ConsulPager {
	cp := &ConsulPager{client: client}
	cp.kvlock = apixtra.NewLock(client, "Pager")
	if cp.kvlock == nil {
		return nil
	}
	cp.alerts = make(map[string]*Alert, 10)
	cp.Channels = make(map[string]Channel, 10)
	cp.Events = make(map[string]*consulapi.UserEvent, 10)
	cp.DefaultMaturity = "3s"
	return cp
}

func LoadPagerFromYAML(fileName string, client *consulapi.Client) *ConsulPager {
	vals, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	var cp ConsulPager
	err = yaml.Unmarshal(vals, &cp)
	cp.kvlock = apixtra.NewLock(client, "Pager")
	if cp.kvlock == nil {
		return nil
	}
	cp.alerts = make(map[string]*Alert, 10)
	cp.Events = make(map[string]*consulapi.UserEvent, 10)
	cp.client = client
	if cp.DefaultMaturity == "" {
		cp.DefaultMaturity = "3s"
	}
	for k, v := range cp.Channels {
		v.SetId(k)
	}
	return &cp
}

func (cp *ConsulPager) IsStopped() bool {
	return cp.kvlock.IsUnlocked()
}

func (cp *ConsulPager) IsLeader() bool {
	err := cp.kvlock.IsLocked()
	if err != nil {
		log.Error(err)
		return false
	}
	return true
}

func (cp *ConsulPager) HasAlert(id string) bool {
	_, ok := cp.alerts[id]
	return ok
}

func (cp *ConsulPager) AddChannel(channel Channel) error {
	id := channel.Id()
	cp.Channels[id] = channel
	return nil
}

func (cp *ConsulPager) AlertIsMature(id string, duration string) bool {
	alert, ok := cp.alerts[id]
	if ok {
		return alert.IsMature(duration)
	} else {
		return false
	}
}

func (cp *ConsulPager) AlertIsSended(id string) bool {
	alert, ok := cp.alerts[id]
	if ok {
		return alert.Sended
	} else {
		return false
	}
}

func (cp *ConsulPager) SendAlert(id string) error {
	alert, ok := cp.alerts[id]
	if ok {
		for _, v := range cp.Channels {
			//log.Infof("Sending alert %s to channel %s", alert.Id, v.Id())
			v.Send(alert)
		}
		alert.Sended = true
		return nil
	} else {
		return errors.New("Can't find alert")
	}
}

func (cp *ConsulPager) SetAlertBody(id string, body string) error {
	alert, ok := cp.alerts[id]
	if ok {
		alert.Body = body
	}
	return nil
}
func (cp *ConsulPager) SetAlertSeverity(id string, severity string) error {
	alert, ok := cp.alerts[id]
	if ok {
		alert.Severity = severity
	}
	return nil
}
func (cp *ConsulPager) Run() error {
	err := cp.kvlock.Lock(nil)
	if err != nil {
		log.Error(err)
		return err
	}

	// Notify events
	go func() {
		var modi uint64
		modi = 0
		dur, _ := time.ParseDuration("3s")
		for {
			uevents, qm, err := cp.client.Event().List("", &consulapi.QueryOptions{WaitTime: dur, WaitIndex: modi})
			if err != nil {
				return
			}

			for _, a := range uevents {
				_, ok := cp.Events[a.ID]
				if ok {
				} else {
					log.Infof("%#v", a)
					cp.Events[a.ID] = a
				}
			}

			modi = qm.LastIndex
			if !cp.kvlock.IsLeader() {
				return
			}
		}
	}()

	// Notify alerts
	go func() {
		var modi uint64
		modi = 0
		dur, _ := time.ParseDuration("3s")
		for {
			hcks, qm, err := cp.client.Health().State("any", &consulapi.QueryOptions{WaitTime: dur, WaitIndex: modi})
			if err != nil {
				return
			}
			for _, a := range hcks {
				switch a.Status {
				case "passing":
					delete(cp.alerts, a.Name)
					//log.Infof("%#v", a)
				case "warning":
					if cp.HasAlert(a.Name) {
						if cp.AlertIsMature(a.Name, cp.DefaultMaturity) {
							if !cp.AlertIsSended(a.Name) {
								cp.SetAlertBody(a.Name, a.Notes)
								cp.SetAlertSeverity(a.Name, "warning")
								cp.SendAlert(a.Name)
							}
						}
					} else {
						alert := NewAlert(a.Name)
						cp.alerts[a.Name] = alert
					}
					//log.Warnf("%#v", a)
				case "critical":
					if cp.HasAlert(a.Name) {
						if cp.AlertIsMature(a.Name, cp.DefaultMaturity) {
							if !cp.AlertIsSended(a.Name) {
								cp.SetAlertBody(a.Name, a.Notes)
								cp.SetAlertSeverity(a.Name, "critical")
								cp.SendAlert(a.Name)
							}
						}

					} else {
						alert := NewAlert(a.Name)
						alert.Subject = a.Name
						cp.alerts[a.Name] = alert
					}
					//log.Errorf("%#v", a)
				}
			}
			modi = qm.LastIndex
			if !cp.kvlock.IsLeader() {
				return
			}
		}
	}()
	return nil
}

func (cp *ConsulPager) Stop() error {
	err := cp.kvlock.Unlock()
	if err != nil {
		return err
	}
	return nil
}

func (cp *ConsulPager) Destroy() error {
	err := cp.Stop()
	if err != nil {
		return err
	}
	err = cp.kvlock.Destroy()
	if err != nil {
		return err
	}
	return nil
}

func (cp *ConsulPager) SaveToYAML(fileName string) error {
	d, err := yaml.Marshal(cp)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	//fmt.Printf("--- t dump:\n%s\n\n", string(d))
	ioutil.WriteFile(fileName, d, 0644)
	return nil
}
