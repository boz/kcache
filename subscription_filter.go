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

	deferReady bool
	refilterch chan filter.Filter

	outch   chan Event
	readych chan struct{}
	stopch  chan struct{}

	filter filter.Filter
	cache  cache

	log logutil.Log
}

func newFilterSubscription(log logutil.Log, parent Subscription, f filter.Filter, deferReady bool) FilterSubscription {

	ctx := context.Background()

	stopch := make(chan struct{})

	s := &filterSubscription{
		parent:     parent,
		refilterch: make(chan filter.Filter),
		outch:      make(chan Event, EventBufsiz),
		readych:    make(chan struct{}),
		stopch:     stopch,
		deferReady: deferReady,
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

	pending := false
	ready := false

	for {
		select {
		case <-preadych:

			preadych = nil

			if s.deferReady && !pending {
				continue
			}

			list, err := s.parent.Cache().List()

			if err != nil {
				s.log.Err(err, "parent.Cache().List()")
				s.parent.Close()
				continue
			}

			s.cache.sync(list)

			close(s.readych)
			ready = true

		case f := <-s.refilterch:

			isNew := !filter.FiltersEqual(s.filter, f)

			switch {

			case preadych != nil && !isNew:
				pending = true
				continue

			case preadych != nil && isNew:
				s.cache.refilter(nil, f)
				s.filter = f
				pending = true
				continue

			case ready && !isNew:
				continue

			case !ready && !isNew:
				close(s.readych)
				ready = true
				continue

			}

			// pready == nil && isNew

			list, err := s.parent.Cache().List()
			if err != nil {
				s.log.Err(err, "parent.Cache().List()")
				s.parent.Close()
				continue
			}

			events := s.cache.refilter(list, f)
			s.filter = f

			if !ready {
				close(s.readych)
				ready = true
				continue
			}

			s.distributeEvents(events)

		case evt, ok := <-s.parent.Events():

			switch {
			case !ok:
				return
			case !ready:
				continue
			}

			events := s.cache.update(evt)

			s.distributeEvents(events)

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
