package kcache

import (
	"context"
	"testing"

	logutil "github.com/boz/go-logutil"
	"github.com/boz/kcache/filter"
	"github.com/boz/kcache/nsname"
	"github.com/boz/kcache/testutil"
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

	testutil.AssertNotReady(t, "publisher", publisher)
	testutil.AssertNotReady(t, "sub", sub)
	testutil.AssertNotReady(t, "sub_wf", sub_wf)
	testutil.AssertNotReady(t, "sub_ff", sub_ff)
	testutil.AssertNotReady(t, "clone", clone)
	testutil.AssertNotReady(t, "clone_wf", clone_wf)
	testutil.AssertNotReady(t, "clone_ff", clone_ff)

	close(readych)

	testutil.AssertReady(t, "publisher", publisher)
	testutil.AssertReady(t, "sub", sub)
	testutil.AssertReady(t, "sub_wf", sub_wf)
	testutil.AssertNotReady(t, "sub_ff", sub_ff)
	testutil.AssertReady(t, "clone", clone)
	testutil.AssertReady(t, "clone_wf", clone_wf)
	testutil.AssertNotReady(t, "clone_ff", clone_ff)

	publisher.Close()

	testutil.AssertDone(t, "parent", parent)
	testutil.AssertDone(t, "publisher", publisher)
	testutil.AssertDone(t, "sub", sub)
	testutil.AssertDone(t, "sub_wf", sub_wf)
	testutil.AssertDone(t, "sub_ff", sub_ff)
	testutil.AssertDone(t, "clone", clone)
	testutil.AssertDone(t, "clone_wf", clone_wf)
	testutil.AssertDone(t, "clone_ff", clone_ff)
}

func TestPublisher_Subscribe(t *testing.T) {
	log := logutil.Default()
	parent, cache, readych := testNewSubscription(t, log, filter.Null())
	publisher := newPublisher(log, parent)
	defer parent.Close()

	doTestPublisherSubscribe(t, parent, cache, publisher, readych)
}

func TestFilterPublisher_Subscribe(t *testing.T) {
	log := logutil.Default()
	parent, cache, readych := testNewSubscription(t, log, filter.Null())
	publisher := newPublisher(log, parent)
	defer parent.Close()

	fpublisher, err := publisher.CloneWithFilter(filter.Null())
	require.NoError(t, err)

	doTestPublisherSubscribe(t, parent, cache, fpublisher, readych)
}

func doTestPublisherSubscribe(t *testing.T,
	parent subscription, cache cache, publisher Controller, readych chan struct{}) {

	close(readych)

	sub, err := publisher.Subscribe()
	require.NoError(t, err)

	testPublisherSubscriber(t, parent, cache, sub)

	publisher.Close()
	testutil.AssertDone(t, "publisher", publisher)
}

func TestPublisher_SubscribeWithFilter(t *testing.T) {
	log := logutil.Default()
	parent, cache, readych := testNewSubscription(t, log, filter.Null())
	publisher := newPublisher(log, parent)
	defer parent.Close()

	doTestPublisherSubscribeWithFilter(t, parent, cache, publisher, readych)
}

func TestFilterPublisher_SubscribeWithFilter(t *testing.T) {
	log := logutil.Default()
	parent, cache, readych := testNewSubscription(t, log, filter.Null())
	publisher := newPublisher(log, parent)
	defer parent.Close()

	fpublisher, err := publisher.CloneWithFilter(filter.Null())
	require.NoError(t, err)

	doTestPublisherSubscribeWithFilter(t, parent, cache, fpublisher, readych)
}

func doTestPublisherSubscribeWithFilter(t *testing.T,
	parent subscription, cache cache, publisher Controller, readych chan struct{}) {

	close(readych)

	f := filter.NSName(nsname.New("a", "c"))
	sub, err := publisher.SubscribeWithFilter(f)
	require.NoError(t, err)

	testutil.AssertReady(t, "sub", sub)

	testPublisherFilteredSubscriber(t, parent, cache, sub)

	publisher.Close()
	testutil.AssertDone(t, "publisher", publisher)
}

func TestPublisher_SubscribeForFilter(t *testing.T) {
	log := logutil.Default()
	parent, cache, readych := testNewSubscription(t, log, filter.Null())
	publisher := newPublisher(log, parent)
	defer parent.Close()
	doTestPublisherSubscribeForFilter(t, parent, cache, publisher, readych)
}

func TestFilterPublisher_SubscribeForFilter(t *testing.T) {
	log := logutil.Default()
	parent, cache, readych := testNewSubscription(t, log, filter.Null())
	publisher := newPublisher(log, parent)
	defer parent.Close()

	fpublisher, err := publisher.CloneWithFilter(filter.Null())
	require.NoError(t, err)

	doTestPublisherSubscribeForFilter(t, parent, cache, fpublisher, readych)
}

func doTestPublisherSubscribeForFilter(t *testing.T,
	parent subscription, cache cache, publisher Controller, readych chan struct{}) {

	close(readych)

	sub, err := publisher.SubscribeForFilter()
	require.NoError(t, err)

	testutil.AssertNotReady(t, "before refilter", sub)

	f := filter.NSName(nsname.New("a", "c"))
	sub.Refilter(f)

	testutil.AssertReady(t, "after refilter", sub)

	testPublisherFilteredSubscriber(t, parent, cache, sub)

	publisher.Close()
	testutil.AssertDone(t, "publisher", publisher)
}

func TestPublisher_Clone(t *testing.T) {
	log := logutil.Default()
	parent, cache, readych := testNewSubscription(t, log, filter.Null())
	publisher := newPublisher(log, parent)
	defer parent.Close()

	doTestPublisherClone(t, parent, cache, publisher, readych)
}

func TestFilterPublisher_Clone(t *testing.T) {
	log := logutil.Default()
	parent, cache, readych := testNewSubscription(t, log, filter.Null())
	publisher := newPublisher(log, parent)
	defer parent.Close()

	fpublisher, err := publisher.CloneWithFilter(filter.Null())
	require.NoError(t, err)

	doTestPublisherClone(t, parent, cache, fpublisher, readych)
}

func doTestPublisherClone(t *testing.T,
	parent subscription, cache cache, publisher Controller, readych chan struct{}) {

	close(readych)

	ppublisher, err := publisher.Clone()
	require.NoError(t, err)

	sub, err := ppublisher.Subscribe()
	require.NoError(t, err)

	testPublisherSubscriber(t, parent, cache, sub)

	publisher.Close()
	testutil.AssertDone(t, "publisher", publisher)
}

func TestPublisher_CloneWithFilter(t *testing.T) {
	log := logutil.Default()
	parent, cache, readych := testNewSubscription(t, log, filter.Null())
	publisher := newPublisher(log, parent)
	defer parent.Close()

	doTestPublisherCloneWithFilter(t, parent, cache, publisher, readych)
}

func TestFilterPublisher_CloneWithFilter(t *testing.T) {
	log := logutil.Default()
	parent, cache, readych := testNewSubscription(t, log, filter.Null())
	publisher := newPublisher(log, parent)
	defer parent.Close()

	fpublisher, err := publisher.CloneWithFilter(filter.Null())
	require.NoError(t, err)

	doTestPublisherCloneWithFilter(t, parent, cache, fpublisher, readych)
}

func doTestPublisherCloneWithFilter(t *testing.T,
	parent subscription, cache cache, publisher Controller, readych chan struct{}) {

	close(readych)

	f := filter.NSName(nsname.New("a", "c"))
	ppublisher, err := publisher.CloneWithFilter(f)
	require.NoError(t, err)

	sub, err := ppublisher.Subscribe()
	require.NoError(t, err)

	testPublisherFilteredSubscriber(t, parent, cache, sub)

	publisher.Close()
	testutil.AssertDone(t, "publisher", publisher)
}

func TestPublisher_CloneForFilter(t *testing.T) {
	log := logutil.Default()
	parent, cache, readych := testNewSubscription(t, log, filter.Null())
	publisher := newPublisher(log, parent)
	defer parent.Close()

	doTestPublisherCloneForFilter(t, parent, cache, publisher, readych)
}

func TestFilterPublisher_CloneForFilter(t *testing.T) {
	log := logutil.Default()
	parent, cache, readych := testNewSubscription(t, log, filter.Null())
	publisher := newPublisher(log, parent)
	defer parent.Close()

	fpublisher, err := publisher.CloneWithFilter(filter.Null())
	require.NoError(t, err)

	doTestPublisherCloneForFilter(t, parent, cache, fpublisher, readych)
}

func doTestPublisherCloneForFilter(t *testing.T,
	parent subscription, cache cache, publisher Controller, readych chan struct{}) {

	close(readych)

	ppublisher, err := publisher.CloneForFilter()
	require.NoError(t, err)

	sub, err := ppublisher.Subscribe()
	require.NoError(t, err)

	testutil.AssertNotReady(t, "ppublisher", ppublisher)
	testutil.AssertNotReady(t, "sub", sub)

	f := filter.NSName(nsname.New("a", "c"))
	ppublisher.Refilter(f)

	testutil.AssertReady(t, "ppublisher", ppublisher)
	testutil.AssertReady(t, "sub", sub)

	testPublisherFilteredSubscriber(t, parent, cache, sub)

	publisher.Close()
	testutil.AssertDone(t, "publisher", publisher)
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
	case <-testutil.AsyncWaitch(ctx):
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
	case <-testutil.AsyncWaitch(ctx):
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
	case <-testutil.AsyncWaitch(ctx):
		assert.Fail(t, "no event")
	}

	list, err := sub.Cache().List()
	assert.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "a", list[0].GetNamespace())
	assert.Equal(t, "c", list[0].GetName())
}
