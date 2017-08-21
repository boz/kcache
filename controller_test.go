package kcache

import (
	"context"
	"testing"
	"time"

	"github.com/boz/kcache/client/mocks"
	"github.com/boz/kcache/filter"
	"github.com/boz/kcache/nsname"
	"github.com/boz/kcache/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func TestController(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	eventch := make(chan watch.Event, 10)
	listch := make(chan time.Time, 1)

	mwatch := &mocks.WatchInterface{}
	mwatch.On("ResultChan").Return(eventch)
	mwatch.On("Stop").Return()

	obj_a := testGenPod("ns", "a", "1")
	obj_b := testGenPod("ns", "b", "2")
	obj_c := testGenPod("ns", "c", "3")
	obj_a_2 := testGenPod("ns", "a", "4")

	fltr := filter.NSName(nsname.New(obj_a.GetNamespace(), obj_a.GetName()))

	list := &v1.PodList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodList",
			APIVersion: "1",
		},
		ListMeta: metav1.ListMeta{
			ResourceVersion: "1",
		},
		Items: []v1.Pod{
			*obj_a,
			*obj_b,
		},
	}

	client := &mocks.Client{}

	client.On("Watch", mock.Anything, mock.AnythingOfType("v1.ListOptions")).Return(mwatch, nil)
	client.On("List", mock.Anything, mock.AnythingOfType("v1.ListOptions")).
		WaitUntil(listch).
		Return(list, nil)

	controller, err := NewBuilder().
		Context(ctx).
		Client(client).
		Create()

	require.NoError(t, err)

	sub, err := controller.Subscribe()
	require.NoError(t, err)
	sub_wf, err := controller.SubscribeWithFilter(fltr)
	require.NoError(t, err)
	sub_ff, err := controller.SubscribeForFilter()
	require.NoError(t, err)

	clone, err := controller.Clone()
	require.NoError(t, err)
	clone_wf, err := controller.CloneWithFilter(fltr)
	require.NoError(t, err)
	clone_ff, err := controller.CloneForFilter()
	require.NoError(t, err)

	csub, err := clone.Subscribe()
	require.NoError(t, err)
	csub_wf, err := clone_wf.Subscribe()
	require.NoError(t, err)
	csub_ff, err := clone_ff.Subscribe()
	require.NoError(t, err)

	testutil.AssertNotReady(t, "controller", controller)

	testutil.AssertNotReady(t, "sub", sub)
	testutil.AssertNotReady(t, "sub_wf", sub_wf)
	testutil.AssertNotReady(t, "sub_ff", sub_ff)

	testutil.AssertNotReady(t, "clone", clone)
	testutil.AssertNotReady(t, "clone_wf", clone_wf)
	testutil.AssertNotReady(t, "clone_ff", clone_ff)

	testutil.AssertNotReady(t, "csub", csub)
	testutil.AssertNotReady(t, "csub_wf", csub_wf)
	testutil.AssertNotReady(t, "csub_ff", csub_ff)

	listch <- time.Now()

	testutil.AssertReady(t, "controller", controller)

	testutil.AssertReady(t, "sub", sub)
	testutil.AssertReady(t, "sub_wf", sub_wf)
	testutil.AssertNotReady(t, "sub_ff", sub_ff)

	testutil.AssertReady(t, "clone", clone)
	testutil.AssertReady(t, "clone_wf", clone_wf)
	testutil.AssertNotReady(t, "clone_ff", clone_ff)

	testutil.AssertReady(t, "csub", csub)
	testutil.AssertReady(t, "csub_wf", csub_wf)
	testutil.AssertNotReady(t, "csub_ff", csub_ff)

	fullcache := func(name string, c CacheController) {

		slist, err := c.Cache().List()
		assert.NoError(t, err, name)
		assert.Len(t, slist, 2, name)

		obj, err := c.Cache().Get(obj_a.GetNamespace(), obj_a.GetName())
		if assert.NoError(t, err, name) && assert.NotNil(t, obj, name) {
			assert.Equal(t, obj_a.GetNamespace(), obj.GetNamespace(), name)
			assert.Equal(t, obj_a.GetName(), obj.GetName(), name)
		}

		obj, err = c.Cache().Get(obj_b.GetNamespace(), obj_b.GetName())
		if assert.NoError(t, err, name) && assert.NotNil(t, obj, name) {
			assert.Equal(t, obj_b.GetNamespace(), obj.GetNamespace(), name)
			assert.Equal(t, obj_b.GetName(), obj.GetName(), name)
		}

	}

	halfcache := func(name string, c CacheController) {

		slist, err := c.Cache().List()
		assert.NoError(t, err, name)

		if assert.Len(t, slist, 1) {
			assert.Equal(t, obj_a.GetNamespace(), slist[0].GetNamespace(), name)
			assert.Equal(t, obj_a.GetName(), slist[0].GetName(), name)
		}

		obj, err := c.Cache().Get(obj_a.GetNamespace(), obj_a.GetName())
		if assert.NoError(t, err, name) && assert.NotNil(t, obj, name) {
			assert.Equal(t, obj_a.GetNamespace(), obj.GetNamespace(), name)
			assert.Equal(t, obj_a.GetName(), obj.GetName(), name)
		}

		obj, err = c.Cache().Get(obj_b.GetNamespace(), obj_b.GetName())
		assert.NoError(t, err, name)
		assert.Nil(t, obj, name)

	}

	fullcache("controller", controller)
	fullcache("sub", sub)
	fullcache("clone", clone)

	halfcache("sub_wf", sub_wf)
	halfcache("clone_wf", clone_wf)

	fullcache("csub", csub)
	halfcache("csub_wf", csub_wf)

	sub_ff.Refilter(fltr)
	clone_ff.Refilter(fltr)

	testutil.AssertReady(t, "sub_ff", sub_ff)
	testutil.AssertReady(t, "clone_ff", clone_ff)
	testutil.AssertReady(t, "csub_ff", csub_ff)

	halfcache("sub_ff", sub_ff)
	halfcache("clone_ff", clone_ff)
	halfcache("csub_ff", sub_ff)

	eventch <- watch.Event{
		Type:   watch.Added,
		Object: obj_c,
	}
	eventch <- watch.Event{
		Type:   watch.Modified,
		Object: obj_a_2,
	}

	fullevt := func(name string, sub Subscription) {
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		atimes := 0
		ctimes := 0

		select {
		case ev, ok := <-sub.Events():
			if !assert.True(t, ok, name) {
				return
			}
			switch ev.Resource().GetName() {
			case obj_a.GetName():
				atimes++
				assert.Equal(t, EventTypeUpdate, ev.Type(), name)
			case obj_c.GetName():
				ctimes++
				assert.Equal(t, EventTypeCreate, ev.Type(), name)
			default:
				assert.Fail(t, "unknown event %v: %#v", ev)
			}
		case <-testutil.AsyncWaitch(ctx):
			assert.Fail(t, "no first event", name)
			return
		}

		select {
		case ev, ok := <-sub.Events():
			if !assert.True(t, ok, name) {
				return
			}
			switch ev.Resource().GetName() {
			case obj_a.GetName():
				atimes++
				assert.Equal(t, EventTypeUpdate, ev.Type(), name)
			case obj_c.GetName():
				ctimes++
				assert.Equal(t, EventTypeCreate, ev.Type(), name)
			default:
				assert.Fail(t, "unknown event %v: %#v", ev)
			}
		case <-testutil.AsyncWaitch(ctx):
			assert.Fail(t, "no second event", name)
			return
		}

		assert.Equal(t, 1, atimes, name)
		assert.Equal(t, 1, ctimes, name)
	}

	halfevt := func(name string, sub Subscription) {
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		select {
		case evt, ok := <-sub.Events():

			if !assert.True(t, ok, name) {
				return
			}

			assert.Equal(t, EventTypeUpdate, evt.Type(), name)
			assert.Equal(t, obj_a.GetNamespace(), evt.Resource().GetNamespace())
			assert.Equal(t, obj_a.GetName(), evt.Resource().GetName())

		case <-testutil.AsyncWaitch(ctx):
			assert.Fail(t, "no events", name)
			return
		}

		select {
		case <-sub.Events():
			assert.Fail(t, "too many events", name)
		case <-testutil.AsyncWaitch(ctx):
		}

	}

	fullevt("sub", sub)
	halfevt("sub_wf", sub_wf)
	halfevt("sub_ff", sub_ff)

	fullevt("csub", csub)
	halfevt("csub_wf", csub_wf)
	halfevt("csub_ff", csub_ff)

	controller.Close()

	testutil.AssertDone(t, "controller", controller)
	testutil.AssertDone(t, "sub", sub)
	testutil.AssertDone(t, "sub_wf", sub_wf)
	testutil.AssertDone(t, "sub_ff", sub_ff)
	testutil.AssertDone(t, "clone", clone)
	testutil.AssertDone(t, "clone_wf", clone_wf)
	testutil.AssertDone(t, "clone_ff", clone_ff)
	testutil.AssertDone(t, "csub", csub)
	testutil.AssertDone(t, "csub_wf", csub_wf)
	testutil.AssertDone(t, "csub_ff", csub_ff)

}
