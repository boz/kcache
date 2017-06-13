package kcache

import (
	lifecycle "github.com/boz/go-lifecycle"
	logutil "github.com/boz/go-logutil"
)

type FilterController interface {
	Controller
	Refilter(Filter)
}

type publisherSubscription struct {
	parent Subscription

	subscribech   chan chan<- Subscription
	unsubscribech chan subscription
	subscriptions map[subscription]struct{}

	lc  lifecycle.Lifecycle
	log logutil.Log
}

func NewPublisher(log logutil.Log, parent Subscription) Controller {
	s := &publisherSubscription{
		parent:        parent,
		subscribech:   make(chan chan<- Subscription),
		unsubscribech: make(chan subscription),
		subscriptions: make(map[subscription]struct{}),
		lc:            lifecycle.New(),
		log:           log.WithComponent("publisher"),
	}

	go s.run()

	return s
}

func (s *publisherSubscription) Ready() <-chan struct{} {
	return s.parent.Ready()
}

func (s *publisherSubscription) Cache() CacheReader {
	return s.parent.Cache()
}

func (s *publisherSubscription) Close() {
	s.parent.Close()
}

func (s *publisherSubscription) Done() <-chan struct{} {
	return s.lc.Done()
}

func (s *publisherSubscription) Subscribe() Subscription {
	resultch := make(chan Subscription, 1)
	select {
	case <-s.lc.ShuttingDown():
		return nil
	case s.subscribech <- resultch:
		return <-resultch
	}
}

func (s *publisherSubscription) run() {
	defer s.log.Un(s.log.Trace("run"))
	defer s.lc.ShutdownCompleted()
	defer s.lc.ShutdownInitiated()

	for {
		select {
		case evt, ok := <-s.parent.Events():
			if !ok {
				return
			}
			s.distributeEvent(evt)
		case resultch := <-s.subscribech:
			resultch <- s.createSubscription()
		case sub := <-s.unsubscribech:
			delete(s.subscriptions, sub)
		}
	}
}

func (s *publisherSubscription) distributeEvent(evt Event) {
	for sub := range s.subscriptions {
		sub.send(evt)
	}
}

func (s *publisherSubscription) createSubscription() Subscription {
	defer s.log.Un(s.log.Trace("doSubscribe"))

	sub := newSubscription(s.log, s.lc.ShuttingDown(), s.parent.Ready(), s.parent.Cache())

	s.subscriptions[sub] = struct{}{}

	go func() {

		select {
		case <-sub.Done():
		case <-s.lc.ShuttingDown():
			sub.Close()
			return
		}

		select {
		case s.unsubscribech <- sub:
		case <-s.lc.ShuttingDown():
		}

	}()

	return sub
}

func NewFilterPublisher(log logutil.Log, subscription FilterSubscription) FilterController {
	return &filterController{subscription, NewPublisher(log, subscription)}
}

type filterController struct {
	subscription FilterSubscription
	parent       Controller
}

func (c *filterController) Cache() CacheReader {
	return c.parent.Cache()
}

func (c *filterController) Ready() <-chan struct{} {
	return c.parent.Ready()
}

func (c *filterController) Subscribe() Subscription {
	return c.parent.Subscribe()
}

func (c *filterController) Done() <-chan struct{} {
	return c.parent.Done()
}

func (c *filterController) Close() {
	c.parent.Close()
}

func (c *filterController) Refilter(filter Filter) {
	c.subscription.Refilter(filter)
}
