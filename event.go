package kcache

import "k8s.io/apimachinery/pkg/apis/meta/v1"

type EventType string

const (
	EventTypeCreate EventType = "create"
	EventTypeUpdate EventType = "update"
	EventTypeDelete EventType = "delete"
)

type Event interface {
	Type() EventType
	Resource() v1.Object
}

type event struct {
	eventType EventType
	resource  v1.Object
}

func NewEvent(et EventType, resource v1.Object) Event {
	return event{et, resource}
}

func (e event) Type() EventType {
	return e.eventType
}

func (e event) Resource() v1.Object {
	return e.resource
}
