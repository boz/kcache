package pod

import "github.com/boz/kcache"

func AdaptSubscription(publisher kcache.Publisher) Subscription {
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
