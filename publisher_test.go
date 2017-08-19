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

func TestPublisher_lifecycle(t *testing.T) {
	log := logutil.Default()
	parent, _, readych := testNewSubscription(t, log, filter.Null())
	publisher := newPublisher(log, parent)
	defer parent.Close()

	sub, err := publisher.Subscribe()
	require.NoError(t, err)
	require.NotNil(t, sub)

	sub_wf, err := publisher.SubscribeWithFilter(filter.Null())
	require.NoError(t, err)
	require.NotNil(t, sub_wf)

	sub_ff, err := publisher.SubscribeForFilter()
	require.NoError(t, err)
	require.NotNil(t, sub_ff)

	clone, err := publisher.Clone()
	require.NoError(t, err)
	require.NotNil(t, clone)

	clone_wf, err := publisher.CloneWithFilter(filter.Null())
	require.NoError(t, err)
	require.NotNil(t, clone_wf)

	clone_ff, err := publisher.CloneForFilter()
	require.NoError(t, err)
	require.NotNil(t, clone_ff)

	testAssertNotReady(t, "publisher", publisher)
	testAssertNotReady(t, "sub", sub)
	testAssertNotReady(t, "sub_wf", sub_wf)
	testAssertNotReady(t, "sub_ff", sub_ff)
	testAssertNotReady(t, "clone", clone)
	testAssertNotReady(t, "clone_wf", clone_wf)
	testAssertNotReady(t, "clone_ff", clone_ff)

	close(readych)

	testAssertReady(t, "publisher", publisher)
	testAssertReady(t, "sub", sub)
	testAssertReady(t, "sub_wf", sub_wf)
	testAssertNotReady(t, "sub_ff", sub_ff)
	testAssertReady(t, "clone", clone)
	testAssertReady(t, "clone_wf", clone_wf)
	testAssertNotReady(t, "clone_ff", clone_ff)

	publisher.Close()

	testAssertDone(t, "parent", parent)
	testAssertDone(t, "publisher", publisher)
	testAssertDone(t, "sub", sub)
	testAssertDone(t, "sub_wf", sub_wf)
	testAssertDone(t, "sub_ff", sub_ff)
	testAssertDone(t, "clone", clone)
	testAssertDone(t, "clone_wf", clone_wf)
	testAssertDone(t, "clone_ff", clone_ff)
}

func TestPublisher_Subscribe(t *testing.T) {
	log := logutil.Default()
	parent, cache, readych := testNewSubscription(t, log, filter.Null())
	publisher := newPublisher(log, parent)
	defer parent.Close()

	close(readych)

	sub, err := publisher.Subscribe()
	require.NoError(t, err)

	testPublisherSubscriber(t, parent, cache, sub)
}

func TestPublisher_SubscribeWithFilter(t *testing.T) {

	log := logutil.Default()
	parent, cache, readych := testNewSubscription(t, log, filter.Null())
	publisher := newPublisher(log, parent)
	defer parent.Close()

	close(readych)

	f := filter.NSName(nsname.New("a", "c"))
	sub, err := publisher.SubscribeWithFilter(f)
	require.NoError(t, err)

	testAssertReady(t, "sub", sub)

	testPublisherFilteredSubscriber(t, parent, cache, sub)
}

func TestPublisher_SubscribeForFilter(t *testing.T) {

	log := logutil.Default()
	parent, cache, readych := testNewSubscription(t, log, filter.Null())
	publisher := newPublisher(log, parent)
	defer parent.Close()

	close(readych)

	sub, err := publisher.SubscribeForFilter()
	require.NoError(t, err)

	testAssertNotReady(t, "before refilter", sub)

	f := filter.NSName(nsname.New("a", "c"))
	sub.Refilter(f)

	testAssertReady(t, "after refilter", sub)

	testPublisherFilteredSubscriber(t, parent, cache, sub)

}

func TestPublisher_Clone(t *testing.T) {
	log := logutil.Default()
	parent, cache, readych := testNewSubscription(t, log, filter.Null())
	publisher := newPublisher(log, parent)
	defer parent.Close()

	close(readych)

	ppublisher, err := publisher.Clone()
	require.NoError(t, err)

	sub, err := ppublisher.Subscribe()
	require.NoError(t, err)

	testPublisherSubscriber(t, parent, cache, sub)
}

func TestPublisher_CloneWithFilter(t *testing.T) {
	log := logutil.Default()
	parent, cache, readych := testNewSubscription(t, log, filter.Null())
	publisher := newPublisher(log, parent)
	defer parent.Close()

	close(readych)

	f := filter.NSName(nsname.New("a", "c"))
	ppublisher, err := publisher.CloneWithFilter(f)
	require.NoError(t, err)

	sub, err := ppublisher.Subscribe()
	require.NoError(t, err)

	testPublisherFilteredSubscriber(t, parent, cache, sub)
}

func TestPublisher_CloneForFilter(t *testing.T) {
	log := logutil.Default()
	parent, cache, readych := testNewSubscription(t, log, filter.Null())
	publisher := newPublisher(log, parent)
	defer parent.Close()

	close(readych)

	ppublisher, err := publisher.CloneForFilter()
	require.NoError(t, err)

	sub, err := ppublisher.Subscribe()
	require.NoError(t, err)

	testAssertNotReady(t, "ppublisher", ppublisher)
	testAssertNotReady(t, "sub", sub)

	f := filter.NSName(nsname.New("a", "c"))
	ppublisher.Refilter(f)

	testAssertReady(t, "ppublisher", ppublisher)
	testAssertReady(t, "sub", sub)

	testPublisherFilteredSubscriber(t, parent, cache, sub)
}

func testPublisherSubscriber(t *testing.T, parent subscription, cache cache, sub Subscription) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	evt := testGenEvent(EventTypeCreate, "a", "b", "1")
	cache.update(evt)
	parent.send(evt)

	select {
	case ev, ok := <-sub.Events():
		assert.True(t, ok)
		assert.NotNil(t, ev)
		assert.Equal(t, EventTypeCreate, ev.Type())
		assert.Equal(t, "a", ev.Resource().GetNamespace())
		assert.Equal(t, "b", ev.Resource().GetName())
	case <-testAsyncWaitch(ctx):
		assert.Fail(t, "no event")
	}

	list, err := sub.Cache().List()
	assert.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "a", list[0].GetNamespace())
	assert.Equal(t, "b", list[0].GetName())
}

func testPublisherFilteredSubscriber(t *testing.T, parent subscription, cache cache, sub Subscription) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	evt := testGenEvent(EventTypeCreate, "a", "b", "1")
	cache.update(evt)
	parent.send(evt)

	select {
	case <-sub.Events():
		assert.Fail(t, "filtered event")
	case <-testAsyncWaitch(ctx):
	}

	evt = testGenEvent(EventTypeCreate, "a", "c", "1")
	cache.update(evt)
	parent.send(evt)

	select {
	case ev, ok := <-sub.Events():
		assert.True(t, ok)
		assert.NotNil(t, ev)
		assert.Equal(t, EventTypeCreate, ev.Type())
		assert.Equal(t, "a", ev.Resource().GetNamespace())
		assert.Equal(t, "c", ev.Resource().GetName())
	case <-testAsyncWaitch(ctx):
		assert.Fail(t, "no event")
	}

	list, err := sub.Cache().List()
	assert.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "a", list[0].GetNamespace())
	assert.Equal(t, "c", list[0].GetName())
}
