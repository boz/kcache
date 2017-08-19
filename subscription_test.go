package kcache

import (
	"context"
	"testing"

	logutil "github.com/boz/go-logutil"
	"github.com/boz/kcache/filter"
	"github.com/stretchr/testify/assert"
)

func TestSubscription(t *testing.T) {
	testDoTestSubscription(t, "close", func(s subscription, _ chan struct{}) { s.Close() })
	testDoTestSubscription(t, "stopch", func(_ subscription, stopch chan struct{}) { close(stopch) })
}

func testDoTestSubscription(t *testing.T, name string, stopfn func(subscription, chan struct{})) {
	log := logutil.Default()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	readych := make(chan struct{})
	stopch := make(chan struct{})
	cache := newCache(ctx, log, stopch, filter.Null())

	sub := newSubscription(log, stopch, readych, cache)
	defer sub.Close()

	testAssertNotDone(t, name, sub)
	testAssertNotReady(t, name, sub)

	evt := testGenEvent(EventTypeCreate, "a", "b", "1")
	sub.send(evt)

	select {
	case ev, ok := <-sub.Events():
		assert.True(t, ok, name)
		assert.Equal(t, evt, ev, name)
	case <-testAsyncWaitch(ctx):
		assert.Fail(t, name)
	}

	stopfn(sub, stopch)

	testAssertDone(t, name, sub)
	testAssertNotReady(t, name, sub)

	sub.send(evt)

	select {
	case _, ok := <-sub.Events():
		assert.False(t, ok, name)
	case <-testAsyncWaitch(ctx):
		assert.Fail(t, name)
	}

}
