package kcache

import (
	"context"

	lifecycle "github.com/boz/go-lifecycle"
	logutil "github.com/boz/go-logutil"
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

func newSubscription(ctx context.Context, log logutil.Log, readych <-chan struct{}, cache CacheReader) subscription {
	log = log.WithComponent("subscription")

	lc := lifecycle.New()
	s := &_subscription{
		readych: readych,
		inch:    make(chan Event),
		outch:   make(chan Event),
		cache:   cache,
		log:     log,
		lc:      lc,
	}

	go s.lc.WatchContext(ctx)

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

	var buf []Event

	for {
		var outch chan Event
		var evt Event

		if len(buf) > 0 {
			outch = s.outch
			evt = buf[0]
		}

		select {
		case <-s.lc.ShutdownRequest():
			return
		case outch <- evt:
			buf = buf[1:]
		case evt := <-s.inch:
			s.log.Debugf("event: %v", evt)
			buf = append(buf, evt)
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
