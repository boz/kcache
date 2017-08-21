package kcache

import (
	"context"
	"testing"

	logutil "github.com/boz/go-logutil"
	"github.com/boz/kcache/filter"
	"github.com/boz/kcache/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCache_Sync(t *testing.T) {
	initial := []metav1.Object{
		testGenPod("default", "pod-1", "1"),
		testGenPod("default", "pod-2", "2"),
	}

	secondary := []metav1.Object{
		testGenPod("default", "pod-1", "3"),
		testGenPod("default", "pod-3", "4"),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stopch := make(chan struct{})
	defer close(stopch)

	log := logutil.Default()
	filter := filter.Null()

	cache := newCache(ctx, log, stopch, filter)

	assert.Len(t, cache.sync(initial), len(initial))

	events := cache.sync(secondary)

	require.Len(t, events, 3)

	found := make(map[string]bool)

	for _, evt := range events {
		name := evt.Resource().GetName()
		switch name {
		case "pod-1":
			if assert.Equal(t, EventTypeUpdate, evt.Type()) {
				found[name] = true
			}
		case "pod-2":
			if assert.Equal(t, EventTypeDelete, evt.Type()) {
				found[name] = true
			}
		case "pod-3":
			if assert.Equal(t, EventTypeCreate, evt.Type()) {
				found[name] = true
			}
		default:
			t.Errorf("unknown pod name: %v", name)
		}
	}
	require.Equal(t, 3, len(found))

	list, err := cache.List()
	require.NoError(t, err)
	require.Len(t, list, 2)

	found = make(map[string]bool)
	for _, obj := range list {
		name := obj.GetName()
		switch name {
		case "pod-1":
			found[name] = true
		case "pod-2":
			assert.Failf(t, "found unexpected pod in list", name)
		case "pod-3":
			found[name] = true
		}
	}

	require.Equal(t, 2, len(found))
}

func TestCache_update(t *testing.T) {
	initial := []metav1.Object{
		testGenPod("default", "pod-1", "1"),
		testGenPod("default", "pod-2", "2"),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stopch := make(chan struct{})
	defer close(stopch)

	log := logutil.Default()
	filter := filter.Null()

	cache := newCache(ctx, log, stopch, filter)

	// first sync returns zero events
	assert.NotEmpty(t, cache.sync(initial))

	{
		events := cache.update(testGenEvent(EventTypeUpdate, "default", "pod-1", "3"))
		require.Len(t, events, 1)
		assert.Equal(t, EventTypeUpdate, events[0].Type())
		assert.Equal(t, "pod-1", events[0].Resource().GetName())
	}

	{
		events := cache.update(testGenEvent(EventTypeDelete, "default", "pod-2", "4"))
		require.Len(t, events, 1)
		assert.Equal(t, EventTypeDelete, events[0].Type())
		assert.Equal(t, "pod-2", events[0].Resource().GetName())
	}

	{
		events := cache.update(testGenEvent(EventTypeCreate, "default", "pod-3", "5"))
		require.Len(t, events, 1)
		assert.Equal(t, EventTypeCreate, events[0].Type())
		assert.Equal(t, "pod-3", events[0].Resource().GetName())
	}

	list, err := cache.List()
	require.NoError(t, err)
	assert.Len(t, list, 2)

	found := make(map[string]bool)
	for _, obj := range list {
		name := obj.GetName()
		switch name {
		case "pod-1":
			found[name] = true
		case "pod-2":
			assert.Failf(t, "found unexpected pod in list", name)
		case "pod-3":
			found[name] = true
		}
	}
	require.Equal(t, 2, len(found))
}

func TestCache_refilter(t *testing.T) {
	initial := []metav1.Object{
		testGenPod("default", "pod-1", "1"),
		testGenPod("default", "pod-2", "2"),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stopch := make(chan struct{})
	defer close(stopch)

	log := logutil.Default()

	cache := newCache(ctx, log, stopch, filter.Null())

	// first sync returns zero events
	assert.NotEmpty(t, cache.sync(initial))

	filter := filter.FN(func(obj metav1.Object) bool {
		return obj.GetNamespace() == "default" &&
			obj.GetName() == "pod-1" &&
			obj.GetResourceVersion() < "5"
	})

	events := cache.refilter(initial, filter)
	require.Len(t, events, 1)
	evt := events[0]
	assert.Equal(t, EventTypeDelete, evt.Type())
	assert.Equal(t, "pod-2", evt.Resource().GetName())

	list, err := cache.List()
	require.NoError(t, err)
	require.Len(t, list, 1)
	obj := list[0]
	require.Equal(t, "pod-1", obj.GetName())

	assert.Empty(t, cache.update(NewEvent(EventTypeDelete, testGenPod("default", "pod-2", "3"))))
	assert.Empty(t, cache.update(NewEvent(EventTypeUpdate, testGenPod("default", "pod-2", "3"))))
	assert.Empty(t, cache.update(NewEvent(EventTypeCreate, testGenPod("default", "pod-2", "3"))))

	assert.Empty(t, cache.update(NewEvent(EventTypeDelete, testGenPod("default", "pod-3", "3"))))
	assert.Empty(t, cache.update(NewEvent(EventTypeUpdate, testGenPod("default", "pod-3", "3"))))
	assert.Empty(t, cache.update(NewEvent(EventTypeCreate, testGenPod("default", "pod-3", "3"))))

	evts := cache.update(NewEvent(EventTypeUpdate, testGenPod("default", "pod-1", "5")))
	assert.Len(t, evts, 1)
	assert.Equal(t, EventTypeDelete, evts[0].Type())
	assert.Equal(t, "default", evts[0].Resource().GetNamespace())
	assert.Equal(t, "pod-1", evts[0].Resource().GetName())

}

func TestCache_lifecycle_ctx(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	log := logutil.Default()

	cache := newCache(ctx, log, nil, filter.Null())

	assert.Len(t, cache.sync([]metav1.Object{testGenPod("a", "b", "1")}), 1)

	obj, err := cache.Get("a", "b")
	assert.NoError(t, err)
	require.NotNil(t, obj)
	assert.Equal(t, "a", obj.GetNamespace())
	assert.Equal(t, "b", obj.GetName())

	list, err := cache.List()
	assert.NoError(t, err)
	assert.Len(t, list, 1)

	cancel()

	testutil.AssertDone(t, "cache", cache)

	assert.Nil(t, cache.sync([]metav1.Object{testGenPod("a", "b", "1")}))
	assert.Nil(t, cache.update(testGenEvent(EventTypeCreate, "a", "b", "2")))
	assert.Nil(t, cache.refilter([]metav1.Object{testGenPod("a", "c", "3")}, filter.All()))

	list, err = cache.List()
	assert.Equal(t, ErrNotRunning, err)
	assert.Empty(t, list)

	obj, err = cache.Get("a", "b")
	assert.Equal(t, ErrNotRunning, err)
	assert.Nil(t, obj)
}

func TestCache_lifecycle_stopch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stopch := make(chan struct{})

	log := logutil.Default()

	cache := newCache(ctx, log, stopch, filter.Null())

	close(stopch)
	testutil.AssertDone(t, "cache", cache)

	assert.Nil(t, cache.sync([]metav1.Object{testGenPod("a", "b", "1")}))
	assert.Nil(t, cache.update(testGenEvent(EventTypeCreate, "a", "b", "2")))
	assert.Nil(t, cache.refilter([]metav1.Object{testGenPod("a", "c", "3")}, filter.All()))

	list, err := cache.List()
	assert.Equal(t, ErrNotRunning, err)
	assert.Empty(t, list)

	obj, err := cache.Get("a", "b")
	assert.Equal(t, ErrNotRunning, err)
	assert.Nil(t, obj)
}
