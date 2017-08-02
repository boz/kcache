package kcache

import (
	"context"
	"errors"

	lifecycle "github.com/boz/go-lifecycle"
	logutil "github.com/boz/go-logutil"
	"github.com/boz/kcache/client"
	"github.com/boz/kcache/filter"
)

var (
	ErrNotRunning = errors.New("Not running")
)

type Publisher interface {
	Subscribe() (Subscription, error)
	SubscribeWithFilter(filter.Filter) (FilterSubscription, error)
	Clone() (Controller, error)
	CloneWithFilter(filter.Filter) (FilterController, error)
}

type CacheController interface {
	Cache() CacheReader
	Ready() <-chan struct{}
}

type Controller interface {
	CacheController
	Publisher
	Done() <-chan struct{}
	Close()
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

	subscription subscription
	publisher    Publisher

	log logutil.Log
	lc  lifecycle.Lifecycle
	ctx context.Context
}

func (c *controller) Ready() <-chan struct{} {
	return c.readych
}

func (c *controller) Close() {
	c.lc.Shutdown()
}

func (c *controller) Done() <-chan struct{} {
	return c.lc.Done()
}

func (c *controller) Cache() CacheReader {
	return c.cache
}

func (c *controller) Subscribe() (Subscription, error) {
	return c.publisher.Subscribe()
}

func (c *controller) SubscribeWithFilter(f filter.Filter) (FilterSubscription, error) {
	return c.publisher.SubscribeWithFilter(f)
}

func (c *controller) Clone() (Controller, error) {
	return c.publisher.Clone()
}

func (c *controller) CloneWithFilter(f filter.Filter) (FilterController, error) {
	return c.publisher.CloneWithFilter(f)
}

func (c *controller) run() {
	defer c.lc.ShutdownCompleted()
	defer c.lc.ShutdownInitiated()
	initialized := false

	for {
		select {
		case <-c.lc.ShutdownRequest():
			return

		case result := <-c.lister.Result():

			if result.err != nil {
				c.log.Err(result.err, "lister error")
				return
			}

			version, err := listResourceVersion(result.list)
			if err != nil {
				c.log.Err(result.err, "error fetching resource version")
				return
			}

			list, err := extractList(result.list)
			if err != nil {
				c.log.Err(result.err, "extractList()")
				return
			}

			events := c.cache.sync(list)

			c.log.Debugf("list complete: version: %v, items: %v, events: %v", version, len(list), len(events))

			if !initialized {
				initialized = true
				close(c.readych)
			}

			c.distributeEvents(events)

			c.watcher.reset(version)

		case evt := <-c.watcher.events():

			events := c.cache.update(evt)

			c.distributeEvents(events)
		}
	}
}

func (c *controller) distributeEvents(events []Event) {
	for _, evt := range events {
		c.subscription.send(evt)
	}
}
