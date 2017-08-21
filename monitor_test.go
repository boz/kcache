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

func TestMonitor(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := logutil.Default()
	parent, cache, readych := testNewSubscription(t, log, filter.Null())
	publisher := newPublisher(log, parent)
	defer parent.Close()

	icalled := make(chan bool)
	ccalled := make(chan bool)
	ucalled := make(chan bool)
	dcalled := make(chan bool)

	h := BuildHandler().OnInitialize(func(objs []metav1.Object) {
		if assert.Len(t, objs, 1) {
			assert.Equal(t, "a", objs[0].GetNamespace())
			assert.Equal(t, "b", objs[0].GetName())
		}
		close(icalled)
	}).OnCreate(func(obj metav1.Object) {
		assert.Equal(t, "b", obj.GetNamespace())
		assert.Equal(t, "b", obj.GetName())
		close(ccalled)
	}).OnUpdate(func(obj metav1.Object) {
		assert.Equal(t, "b", obj.GetNamespace())
		assert.Equal(t, "b", obj.GetName())
		close(ucalled)
	}).OnDelete(func(obj metav1.Object) {
		assert.Equal(t, "b", obj.GetNamespace())
		assert.Equal(t, "b", obj.GetName())
		close(dcalled)
	}).Create()

	cache.sync([]metav1.Object{testGenPod("a", "b", "1")})

	m, err := NewMonitor(publisher, h)
	assert.NoError(t, err)

	close(readych)

	parent.send(testGenEvent(EventTypeCreate, "b", "b", "1"))
	parent.send(testGenEvent(EventTypeUpdate, "b", "b", "2"))
	parent.send(testGenEvent(EventTypeDelete, "b", "b", "3"))

	select {
	case <-icalled:
	case <-testutil.AsyncWaitch(ctx):
		assert.Fail(t, "initialize not called")
	}

	select {
	case <-ccalled:
	case <-testutil.AsyncWaitch(ctx):
		assert.Fail(t, "create not called")
	}

	select {
	case <-ucalled:
	case <-testutil.AsyncWaitch(ctx):
		assert.Fail(t, "update not called")
	}

	select {
	case <-dcalled:
	case <-testutil.AsyncWaitch(ctx):
		assert.Fail(t, "delete not called")
	}

	m.Close()
	testutil.AssertDone(t, "monitor", m)
}

func TestMonitor_lifecycle_close_early(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := logutil.Default()
	parent, _, readych := testNewSubscription(t, log, filter.Null())
	publisher := newPublisher(log, parent)

	calledch := make(chan bool)

	h := BuildHandler().OnInitialize(func(_ []metav1.Object) {
		close(calledch)
	}).Create()

	m, err := NewMonitor(publisher, h)
	require.NoError(t, err)

	publisher.Close()

	testutil.AssertDone(t, "publisher", m)

	close(readych)

	select {
	case <-calledch:
		assert.Fail(t, "initialize called")
	case <-testutil.AsyncWaitch(ctx):
	}

}
