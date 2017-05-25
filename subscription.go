package kcache

import (
	lifecycle "github.com/boz/go-lifecycle"
	logutil "github.com/boz/go-logutil"
)

const (
	eventBufsiz = 100
)

type Subscription interface {
	Cache() CacheReader
	Ready() <-chan struct{}
	Events() <-chan Event
	Close()
}

type subscription interface {
	Subscription
	send(Event)
	done() <-chan struct{}
}

type _subscription struct {
	outch chan Event
	inch  chan Event

	readych <-chan struct{}

	cache CacheReader

	log logutil.Log
	lc  lifecycle.Lifecycle
}

func newSubscription(log logutil.Log, stopch <-chan struct{}, readych <-chan struct{}, cache CacheReader) subscription {
	log = log.WithComponent("subscription")

	lc := lifecycle.New()
	s := &_subscription{
		readych: readych,
		inch:    make(chan Event),
		outch:   make(chan Event, eventBufsiz),
		cache:   cache,
		log:     log,
		lc:      lc,
	}

	go s.lc.WatchChannel(stopch)

	go s.run()
	return s
}

func (s *_subscription) done() <-chan struct{} {
	return s.lc.Done()
}

func (s *_subscription) send(ev Event) {
	select {
	case s.inch <- ev:
	case <-s.lc.ShuttingDown():
	}
}

func (s *_subscription) run() {
	defer s.log.Un(s.log.Trace("run"))
	defer s.lc.ShutdownCompleted()
	defer s.lc.ShutdownInitiated()
	defer close(s.outch)

	for {
		select {
		case <-s.lc.ShutdownRequest():
			return
		case evt := <-s.inch:
			select {
			case s.outch <- evt:
			default:
				s.log.Warnf("event buffer overrun")
			}
		}
	}
}

func (s *_subscription) Close() {
	s.lc.Shutdown()
}

func (s *_subscription) Ready() <-chan struct{} {
	return s.readych
}

func (s *_subscription) Events() <-chan Event {
	return s.outch
}

func (s *_subscription) Cache() CacheReader {
	return s.cache
}
