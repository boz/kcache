package kcache

import (
	"context"
	"testing"

	logutil "github.com/boz/go-logutil"
	"github.com/boz/kcache/filter"
	"github.com/boz/kcache/nsname"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilterSubscriptionReady_immediate(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := logutil.Default()
	parent, cache, readych := testNewSubscription(t, log, filter.Null())
	sub := newFilterSubscription(log, parent, filter.Null(), false)
	defer parent.Close()

	testDoFilterSubscriptionReady(t, "immediate", parent, sub, cache)

	close(readych)

	testAssertReady(t, "immediate", sub)

	list, err := sub.Cache().List()
	assert.NoError(t, err)
	assert.NotEmpty(t, list)

	evt := testGenEvent(EventTypeCreate, "a", "c", "1")
	parent.send(evt)

	select {
	case <-sub.Events():
	case <-testAsyncWaitch(ctx):
		assert.Fail(t, "deferred")
	}

	testDoTestFilterSubscriptionShutdown(t, "immediate", parent, sub)

}

func TestFilterSubscriptionReady_deferred(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := logutil.Default()
	parent, cache, readych := testNewSubscription(t, log, filter.Null())
	sub := newFilterSubscription(log, parent, filter.Null(), true)
	defer parent.Close()

	testDoFilterSubscriptionReady(t, "deferred", parent, sub, cache)

	close(readych)

	testAssertNotReady(t, "deferred", sub)

	list, err := sub.Cache().List()
	assert.NoError(t, err)
	assert.Empty(t, list)

	evt := testGenEvent(EventTypeCreate, "a", "c", "1")
	parent.send(evt)

	select {
	case <-sub.Events():
		assert.Fail(t, "deferred")
	case <-testAsyncWaitch(ctx):
	}

	testDoTestFilterSubscriptionShutdown(t, "deferred", parent, sub)

}

func TestFilterSubscriptionRefilter_immediate_refilter_before_ready(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := logutil.Default()
	parent, cache, readych := testNewSubscription(t, log, filter.Null())
	sub := newFilterSubscription(log, parent, filter.Null(), false)
	defer parent.Close()

	cache.update(testGenEvent(EventTypeCreate, "a", "b", "1"))
	cache.update(testGenEvent(EventTypeCreate, "a", "c", "2"))

	f := filter.NSName(nsname.New("a", "c"))

	sub.Refilter(f)

	testAssertNotReady(t, "ready", sub)

	parent.send(testGenEvent(EventTypeCreate, "a", "d", "3"))

	select {
	case <-sub.Events():
		assert.Fail(t, "events before ready")
	case <-testAsyncWaitch(ctx):
	}

	close(readych)

	testAssertReady(t, "ready", sub)

	select {
	case <-sub.Events():
		assert.Fail(t, "events after ready")
	case <-testAsyncWaitch(ctx):
	}

	list, err := sub.Cache().List()
	assert.NoError(t, err)
	assert.Len(t, list, 1)

	assert.Equal(t, "a", list[0].GetNamespace())
	assert.Equal(t, "c", list[0].GetName())

	parent.send(testGenEvent(EventTypeCreate, "a", "c", "4"))
	select {
	case evt, ok := <-sub.Events():
		require.True(t, ok)
		require.NotNil(t, evt)
		assert.Equal(t, EventTypeUpdate, evt.Type())
		assert.Equal(t, "a", evt.Resource().GetNamespace())
		assert.Equal(t, "c", evt.Resource().GetName())
	case <-testAsyncWaitch(ctx):
		assert.Fail(t, "events after ready")
	}

	parent.send(testGenEvent(EventTypeCreate, "b", "c", "4"))
	select {
	case <-sub.Events():
		assert.Fail(t, "filtered event")
	case <-testAsyncWaitch(ctx):
	}

}

func TestFilterSubscriptionRefilter_immediate_refilter_after_ready(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := logutil.Default()
	parent, cache, readych := testNewSubscription(t, log, filter.Null())
	sub := newFilterSubscription(log, parent, filter.Null(), false)
	defer parent.Close()

	cache.update(testGenEvent(EventTypeCreate, "a", "b", "1"))
	cache.update(testGenEvent(EventTypeCreate, "a", "c", "2"))

	f := filter.NSName(nsname.New("a", "c"))

	close(readych)

	testAssertReady(t, "ready", sub)

	sub.Refilter(f)

	select {
	case evt, ok := <-sub.Events():
		require.True(t, ok)
		require.NotNil(t, evt)
		assert.Equal(t, EventTypeDelete, evt.Type())
		assert.Equal(t, "a", evt.Resource().GetNamespace())
		assert.Equal(t, "b", evt.Resource().GetName())
	case <-testAsyncWaitch(ctx):
		assert.Fail(t, "events after refilter")
	}

	list, err := sub.Cache().List()
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "a", list[0].GetNamespace())
	assert.Equal(t, "c", list[0].GetName())

	parent.send(testGenEvent(EventTypeCreate, "a", "c", "4"))
	select {
	case evt, ok := <-sub.Events():
		require.True(t, ok)
		require.NotNil(t, evt)
		assert.Equal(t, EventTypeUpdate, evt.Type())
		assert.Equal(t, "a", evt.Resource().GetNamespace())
		assert.Equal(t, "c", evt.Resource().GetName())
	case <-testAsyncWaitch(ctx):
		assert.Fail(t, "events after ready")
	}

	parent.send(testGenEvent(EventTypeCreate, "b", "c", "4"))
	select {
	case <-sub.Events():
		assert.Fail(t, "filtered event")
	case <-testAsyncWaitch(ctx):
	}

}

func TestFilterSubscriptionRefilter_deferred_refilter_before_ready(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := logutil.Default()
	parent, cache, readych := testNewSubscription(t, log, filter.Null())
	sub := newFilterSubscription(log, parent, filter.Null(), true)
	defer parent.Close()

	cache.update(testGenEvent(EventTypeCreate, "a", "b", "1"))
	cache.update(testGenEvent(EventTypeCreate, "a", "c", "2"))

	f := filter.NSName(nsname.New("a", "c"))

	sub.Refilter(f)

	testAssertNotReady(t, "ready before refilter", sub)

	close(readych)

	testAssertReady(t, "sub after refilter", sub)

	select {
	case <-sub.Events():
		assert.Fail(t, "events after refilter")
	case <-testAsyncWaitch(ctx):
	}

	list, err := sub.Cache().List()
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "a", list[0].GetNamespace())
	assert.Equal(t, "c", list[0].GetName())

	parent.send(testGenEvent(EventTypeCreate, "a", "c", "4"))
	select {
	case evt, ok := <-sub.Events():
		require.True(t, ok)
		require.NotNil(t, evt)
		assert.Equal(t, EventTypeUpdate, evt.Type())
		assert.Equal(t, "a", evt.Resource().GetNamespace())
		assert.Equal(t, "c", evt.Resource().GetName())
	case <-testAsyncWaitch(ctx):
		assert.Fail(t, "events after ready")
	}

	parent.send(testGenEvent(EventTypeCreate, "b", "c", "4"))
	select {
	case <-sub.Events():
		assert.Fail(t, "filtered event")
	case <-testAsyncWaitch(ctx):
	}
}

func TestFilterSubscriptionRefilter_deferred_refilter_after_ready(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := logutil.Default()
	parent, cache, readych := testNewSubscription(t, log, filter.Null())
	sub := newFilterSubscription(log, parent, filter.Null(), true)
	defer parent.Close()

	cache.update(testGenEvent(EventTypeCreate, "a", "b", "1"))
	cache.update(testGenEvent(EventTypeCreate, "a", "c", "2"))

	f := filter.NSName(nsname.New("a", "c"))

	close(readych)

	testAssertNotReady(t, "sub  before refilter", sub)

	sub.Refilter(f)

	testAssertReady(t, "sub after refilter", sub)

	select {
	case <-sub.Events():
		assert.Fail(t, "events after refilter")
	case <-testAsyncWaitch(ctx):
	}

	list, err := sub.Cache().List()
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "a", list[0].GetNamespace())
	assert.Equal(t, "c", list[0].GetName())

	parent.send(testGenEvent(EventTypeCreate, "a", "c", "4"))
	select {
	case evt, ok := <-sub.Events():
		require.True(t, ok)
		require.NotNil(t, evt)
		assert.Equal(t, EventTypeUpdate, evt.Type())
		assert.Equal(t, "a", evt.Resource().GetNamespace())
		assert.Equal(t, "c", evt.Resource().GetName())
	case <-testAsyncWaitch(ctx):
		assert.Fail(t, "events after ready")
	}

	parent.send(testGenEvent(EventTypeCreate, "b", "c", "4"))
	select {
	case <-sub.Events():
		assert.Fail(t, "filtered event")
	case <-testAsyncWaitch(ctx):
	}
}

func testDoFilterSubscriptionReady(t *testing.T, name string, parent subscription, sub FilterSubscription, c cache) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	testAssertNotDone(t, name, sub)
	testAssertNotReady(t, name, sub)

	evt := testGenEvent(EventTypeCreate, "a", "b", "1")
	c.update(evt)
	parent.send(evt)

	testAssertNotReady(t, name, sub)

	list, err := sub.Cache().List()
	assert.NoError(t, err, name)
	assert.Empty(t, list, name)

	select {
	case <-sub.Events():
		assert.Fail(t, name)
	case <-testAsyncWaitch(ctx):
	}

}

func testDoTestFilterSubscriptionShutdown(t *testing.T, name string, parent subscription, sub FilterSubscription) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	parent.Close()

	testAssertDone(t, name, sub)

	select {
	case _, ok := <-sub.Events():
		assert.False(t, ok, name)
	case <-testAsyncWaitch(ctx):
		assert.Fail(t, name)
	}

}
