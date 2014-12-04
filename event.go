package consul_pager

import ()

type Event struct {
	Id string
}

func NewEvent(id string) *Event {
	return &Event{Id: id}
}
