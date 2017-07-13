package ingress

import (
	"github.com/boz/kcache"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

type CacheReader interface {
	Get(ns string, name string) (*v1beta1.Ingress, error)
	List() ([]*v1beta1.Ingress, error)
}

type CacheController interface {
	Cache() CacheReader
	Ready() <-chan struct{}
}

type Event interface {
	Type() kcache.EventType
	Resource() *v1beta1.Ingress
}

type Subscription interface {
	CacheController
	Events() <-chan Event
	Close()
	Done() <-chan struct{}
}
