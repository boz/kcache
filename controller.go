package kcache

import (
	"context"

	lifecycle "github.com/boz/go-lifecycle"
	logutil "github.com/boz/go-logutil"
	"github.com/boz/kcache/client"
)

type Publisher interface {
	Subscribe() Subscription
}

type CacheController interface {
	Cache() CacheReader
	Ready() <-chan struct{}
}

type Controller interface {
	CacheController
	Publisher
	Done() <-chan struct{}
	Stop()
}

func NewController(ctx context.Context, log logutil.Log, client client.Client) (Controller, error) {
	return NewBuilder().
		Context(ctx).
		Log(log).
		Client(client).
		Create()
}

type controller struct {

	// closed when initialization complete
	readych chan struct{}

	watcher watcher
	lister  lister
	cache   cache

	subscribech   chan chan<- Subscription
	unsubscribech chan subscription
	subscriptions map[subscription]struct{}

	log logutil.Log
	lc  lifecycle.Lifecycle
	ctx context.Context
}

func (c *controller) Ready() <-chan struct{} {
	return c.readych
}

func (c *controller) Stop() {
	c.lc.Shutdown()
}

func (c *controller) Done() <-chan struct{} {
	return c.lc.Done()
}

func (c *controller) Cache() CacheReader {
	return c.cache
}

func (c *controller) Subscribe() Subscription {
	resultch := make(chan Subscription, 1)
	select {
	case <-c.lc.ShuttingDown():
		return nil
	case c.subscribech <- resultch:
		return <-resultch
	}
}

func (c *controller) run() {
	defer c.log.Un(c.log.Trace("run"))
	defer c.lc.ShutdownCompleted()
	defer c.lc.ShutdownInitiated()
	initialized := false

	for {
		select {
		case <-c.lc.ShutdownRequest():
			return

		case result := <-c.lister.Result():

			if result.err != nil {
				c.log.Err(result.err, "lister.Result()")
				return
			}

			version, err := listResourceVersion(result.list)
			if err != nil {
				c.log.Err(result.err, "lister.Result()")
				return
			}

			events := c.cache.sync(result.list)

			if !initialized {
				initialized = true
				close(c.readych)
			}

			c.distributeEvents(events)

			c.watcher.reset(version)

		case evt := <-c.watcher.events():

			events := c.cache.update(evt)

			c.distributeEvents(events)

		case resultch := <-c.subscribech:
			c.doSubscribe(resultch)
		case sub := <-c.unsubscribech:
			delete(c.subscriptions, sub)
		}
	}
}

func (c *controller) distributeEvents(events []Event) {
	for _, evt := range events {
		for sub := range c.subscriptions {
			sub.send(evt)
		}
	}
}

func (c *controller) doSubscribe(ch chan<- Subscription) {
	defer c.log.Un(c.log.Trace("doSubscribe"))

	sub := newSubscription(c.ctx, c.log, c.readych, c.cache)
	c.subscriptions[sub] = struct{}{}

	go func() {
		select {
		case <-c.lc.ShuttingDown():
			sub.Close()
		case <-sub.done():
			c.unsubscribech <- sub
		}
	}()
	ch <- sub
}
