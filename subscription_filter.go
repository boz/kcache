package kcache

import (
	"context"

	logutil "github.com/boz/go-logutil"
	"github.com/boz/kcache/filter"
)

type FilterSubscription interface {
	Subscription
	Refilter(filter.Filter)
}

type filterSubscription struct {
	parent Subscription

	refilterch chan filter.Filter

	outch   chan Event
	readych chan struct{}
	stopch  chan struct{}

	filter filter.Filter
	cache  cache

	log logutil.Log
}

func newFilterSubscription(log logutil.Log, parent Subscription, f filter.Filter) FilterSubscription {

	ctx := context.Background()

	stopch := make(chan struct{})

	s := &filterSubscription{
		parent:     parent,
		refilterch: make(chan filter.Filter),
		outch:      make(chan Event, EventBufsiz),
		readych:    make(chan struct{}),
		stopch:     stopch,
		filter:     f,
		cache:      newCache(ctx, log, stopch, f),
		log:        log,
	}

	go s.run()

	return s
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
func (s *filterSubscription) Done() <-chan struct{} {
	return s.parent.Done()
}

func (s *filterSubscription) Refilter(filter filter.Filter) {
	select {
	case s.refilterch <- filter:
	case <-s.Done():
	}
}

func (s *filterSubscription) run() {
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
			} else {
				s.cache.sync(list)
				close(s.readych)
			}

			preadych = nil

		case f := <-s.refilterch:

			if filter.FiltersEqual(s.filter, f) {
				continue
			}

			list, err := s.parent.Cache().List()
			if err != nil {
				s.log.Err(err, "parent.Cache().List()")
				s.parent.Close()
				continue
			}

			events := s.cache.refilter(list, f)
			s.filter = f
			s.distributeEvents(events)

		case evt, ok := <-s.parent.Events():
			if !ok {
				return
			}

			s.distributeEvents(s.cache.update(evt))
		}
	}
}

func (s *filterSubscription) distributeEvents(events []Event) {
	for _, evt := range events {
		select {
		case s.outch <- evt:
		default:
			s.log.Warnf("event buffer overrun")
		}
	}
}
