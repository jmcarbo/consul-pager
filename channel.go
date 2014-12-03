package consul_pager

import (
	"errors"
	//"fmt"
)

type ChannelI interface {
	Id() string
	SetId(string)
	Send(alert *Alert) error
}

type Channel struct {
	ChannelI
}

func (c *Channel) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var a map[string]SlackChannel
	if unmarshal(&a) == nil {
		//fmt.Printf("%#v\n", a)
		var b SlackChannel
		b = a["channeli"]
		(*c).ChannelI = &b
		return nil
	}
	return errors.New("Unable to recognize channel type")
}
