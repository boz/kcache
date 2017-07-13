package pod

import (
	"github.com/boz/kcache"
	"k8s.io/client-go/pkg/api/v1"
)

type CacheReader interface {
	Get(ns string, name string) (*v1.Pod, error)
	List() ([]*v1.Pod, error)
}

type CacheController interface {
	Cache() CacheReader
	Ready() <-chan struct{}
}

type Event interface {
	Type() kcache.EventType
	Resource() *v1.Pod
}

type Subscription interface {
	CacheController
	Events() <-chan Event
	Close()
	Done() <-chan struct{}
}
