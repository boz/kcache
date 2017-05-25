package kcache

import (
	"context"

	logutil "github.com/boz/go-logutil"
)

type filterSubscription struct {
	parent Subscription

	outch   chan Event
	readych chan struct{}
	stopch  chan struct{}
	cache   cache

	log logutil.Log
}

func NewFilterSubscription(log logutil.Log, parent Subscription, filter Filter) Subscription {

	ctx := context.Background()

	stopch := make(chan struct{})

	s := &filterSubscription{
		parent:  parent,
		outch:   make(chan Event, eventBufsiz),
		readych: make(chan struct{}),
		stopch:  stopch,
		cache:   newCache(ctx, log, stopch, filter),
		log:     log,
	}

	go s.run()

	return s
}

func (s *filterSubscription) run() {
	defer s.log.Un(s.log.Trace("run"))
	defer close(s.outch)
	defer close(s.stopch)

	preadych := s.parent.Ready()

	for {
		select {
		case <-preadych:

			list, err := s.parent.Cache().List()
			if err != nil {
				s.log.Err(err, "parent.Cache().List()")
				s.parent.Close()
				return
			}

			s.cache.sync(list)
			close(s.readych)
			preadych = nil

		case evt, ok := <-s.parent.Events():
			if !ok {
				return
			}

			for _, evt := range s.cache.update(evt) {
				select {
				case s.outch <- evt:
				default:
					s.log.Warnf("event buffer overrun")
				}
			}
		}
	}
}

func (s *filterSubscription) Cache() CacheReader {
	return s.cache
}
func (s *filterSubscription) Ready() <-chan struct{} {
	return s.readych
}
func (s *filterSubscription) Events() <-chan Event {
	return s.outch
}
func (s *filterSubscription) Close() {
	s.parent.Close()
}
