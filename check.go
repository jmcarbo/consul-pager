package consul_pager

import (
	"errors"
	"github.com/armon/consul-api"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Check struct {
	Name, Interval, Script string
}

func LoadChecksFromYAML(fileName string, client *consulapi.Client) error {
	vals, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	var checks []Check
	err = yaml.Unmarshal(vals, &checks)
	if err != nil {
		return err
	}
	for _, c := range checks {
		if c.Name == "" || c.Interval == "" || c.Script == "" {
			errors.New("Needs name, interval and script")
		} else {
			client := Connect()
			client.Agent().CheckRegister(&consulapi.AgentCheckRegistration{c.Name, c.Name, "",
				consulapi.AgentServiceCheck{Interval: c.Interval, Script: c.Script}})
		}
	}
	return nil
}
