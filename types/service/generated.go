// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/cheekybits/genny

package service

import (
	"fmt"

	"github.com/boz/kcache"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
) // This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/cheekybits/genny

var (
	ErrInvalidType = fmt.Errorf("invalid type")
	adapter        = _adapter{}
)

type CacheReader interface {
	Get(ns string, name string) (*v1.Service, error)
	List() ([]*v1.Service, error)
}

type CacheController interface {
	Cache() CacheReader
	Ready() <-chan struct{}
}

type Event interface {
	Type() kcache.EventType
	Resource() *v1.Service
}

type Subscription interface {
	CacheController
	Events() <-chan Event
	Close()
	Done() <-chan struct{}
}

type _adapter struct{}

func (_adapter) adaptObject(obj metav1.Object) (*v1.Service, error) {
	if obj, ok := obj.(*v1.Service); ok {
		return obj, nil
	}
	return nil, ErrInvalidType
}

func (a _adapter) adaptList(objs []metav1.Object) ([]*v1.Service, error) {
	var ret []*v1.Service
	for _, orig := range objs {
		adapted, err := a.adaptObject(orig)
		if err != nil {
			continue
		}
		ret = append(ret, adapted)
	}
	return ret, nil
}

func NewCache(parent kcache.CacheReader) CacheReader {
	return &cache{parent}
}

type cache struct {
	parent kcache.CacheReader
}

func (c *cache) Get(ns string, name string) (*v1.Service, error) {
	obj, err := c.parent.Get(ns, name)
	switch {
	case err != nil:
		return nil, err
	case obj == nil:
		return nil, nil
	default:
		return adapter.adaptObject(obj)
	}
}

func (c *cache) List() ([]*v1.Service, error) {
	objs, err := c.parent.List()
	if err != nil {
		return nil, err
	}
	return adapter.adaptList(objs)
}

type event struct {
	etype    kcache.EventType
	resource *v1.Service
}

func wrapEvent(evt kcache.Event) (Event, error) {
	obj, err := adapter.adaptObject(evt.Resource())
	if err != nil {
		return nil, err
	}
	return event{evt.Type(), obj}, nil
}

func (e event) Type() kcache.EventType {
	return e.etype
}

func (e event) Resource() *v1.Service {
	return e.resource
}

func SubscribeTo(publisher kcache.Publisher) Subscription {
	parent := publisher.Subscribe()
	return newSubscription(parent)
}

type subscription struct {
	parent kcache.Subscription
	cache  CacheReader
	outch  chan Event
}

func newSubscription(parent kcache.Subscription) Subscription {
	s := &subscription{
		parent: parent,
		cache:  NewCache(parent.Cache()),
		outch:  make(chan Event, kcache.EventBufsiz),
	}
	go s.run()
	return s
}

func (s *subscription) run() {
	defer close(s.outch)
	for pevt := range s.parent.Events() {
		evt, err := wrapEvent(pevt)
		if err != nil {
			continue
		}
		select {
		case s.outch <- evt:
		default:
		}
	}
}

func (s *subscription) Cache() CacheReader {
	return s.cache
}

func (s *subscription) Ready() <-chan struct{} {
	return s.parent.Ready()
}

func (s *subscription) Events() <-chan Event {
	return s.outch
}

func (s *subscription) Close() {
	s.parent.Close()
}

func (s *subscription) Done() <-chan struct{} {
	return s.parent.Done()
}
