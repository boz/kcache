/*
* AUTO GENERATED - DO NOT EDIT BY HAND
 */

package pod

import (
	"context"
	"testing"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"

	logutil "github.com/boz/go-logutil"
	"github.com/boz/kcache"
	"github.com/boz/kcache/client/mocks"
	"github.com/boz/kcache/filter"
	"github.com/boz/kcache/nsname"
	"github.com/boz/kcache/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestController(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := logutil.Default()

	eventch := make(chan watch.Event, 10)
	listch := make(chan time.Time, 1)

	mwatch := &mocks.WatchInterface{}
	mwatch.On("ResultChan").Return(eventch)
	mwatch.On("Stop").Return()

	obj_a := testGenObject("ns", "a", "1")
	obj_b := testGenObject("ns", "b", "2")
	obj_c := testGenObject("ns", "c", "3")
	obj_a_2 := testGenObject("ns", "a", "4")

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

	controller, err := BuildController(ctx, log, client)
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
				assert.Equal(t, kcache.EventTypeUpdate, ev.Type(), name)
			case obj_c.GetName():
				ctimes++
				assert.Equal(t, kcache.EventTypeCreate, ev.Type(), name)
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
				assert.Equal(t, kcache.EventTypeUpdate, ev.Type(), name)
			case obj_c.GetName():
				ctimes++
				assert.Equal(t, kcache.EventTypeCreate, ev.Type(), name)
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

			assert.Equal(t, kcache.EventTypeUpdate, evt.Type(), name)
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

func TestMonitor(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := logutil.Default()

	eventch := make(chan watch.Event, 10)

	mwatch := &mocks.WatchInterface{}
	mwatch.On("ResultChan").Return(eventch)
	mwatch.On("Stop").Return()

	obj_a := testGenObject("ns", "a", "1")
	obj_b := testGenObject("ns", "b", "2")
	obj_c := testGenObject("ns", "a", "3")
	obj_d := testGenObject("ns", "b", "4")

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
		},
	}

	client := &mocks.Client{}

	client.On("Watch", mock.Anything, mock.AnythingOfType("v1.ListOptions")).Return(mwatch, nil)
	client.On("List", mock.Anything, mock.AnythingOfType("v1.ListOptions")).
		Return(list, nil)

	controller, err := BuildController(ctx, log, client)
	require.NoError(t, err)
	defer controller.Close()

	icalled := make(chan bool)
	ccalled := make(chan bool)
	ucalled := make(chan bool)
	dcalled := make(chan bool)

	u_icalled := make(chan bool)
	u_ccalled := make(chan bool)
	u_ucalled := make(chan bool)
	u_dcalled := make(chan bool)

	h := BuildHandler().OnInitialize(func(objs []*v1.Pod) {
		if assert.Len(t, objs, 1) {
			assert.Equal(t, obj_a.GetNamespace(), objs[0].GetNamespace())
			assert.Equal(t, obj_a.GetName(), objs[0].GetName())
		}
		close(icalled)
	}).OnCreate(func(obj *v1.Pod) {
		assert.Equal(t, obj_b.GetNamespace(), obj.GetNamespace())
		assert.Equal(t, obj_b.GetName(), obj.GetName())
		close(ccalled)
	}).OnUpdate(func(obj *v1.Pod) {
		assert.Equal(t, obj_c.GetNamespace(), obj.GetNamespace())
		assert.Equal(t, obj_c.GetName(), obj.GetName())
		close(ucalled)
	}).OnDelete(func(obj *v1.Pod) {
		assert.Equal(t, obj_d.GetNamespace(), obj.GetNamespace())
		assert.Equal(t, obj_d.GetName(), obj.GetName())
		close(dcalled)
	}).Create()

	uh := BuildUnitaryHandler().OnInitialize(func(obj *v1.Pod) {
		assert.Equal(t, obj_a.GetNamespace(), obj.GetNamespace())
		assert.Equal(t, obj_a.GetName(), obj.GetName())
		close(u_icalled)
	}).OnCreate(func(obj *v1.Pod) {
		assert.Equal(t, obj_b.GetNamespace(), obj.GetNamespace())
		assert.Equal(t, obj_b.GetName(), obj.GetName())
		close(u_ccalled)
	}).OnUpdate(func(obj *v1.Pod) {
		assert.Equal(t, obj_c.GetNamespace(), obj.GetNamespace())
		assert.Equal(t, obj_c.GetName(), obj.GetName())
		close(u_ucalled)
	}).OnDelete(func(obj *v1.Pod) {
		assert.Equal(t, obj_d.GetNamespace(), obj.GetNamespace())
		assert.Equal(t, obj_d.GetName(), obj.GetName())
		close(u_dcalled)
	}).Create()

	m, err := NewMonitor(controller, h)
	assert.NoError(t, err)

	um, err := NewMonitor(controller, ToUnitary(log, uh))
	assert.NoError(t, err)

	select {
	case <-icalled:
	case <-testutil.AsyncWaitch(ctx):
		assert.Fail(t, "initialize not called")
	}

	select {
	case <-u_icalled:
	case <-testutil.AsyncWaitch(ctx):
		assert.Fail(t, "unitary initialize not called")
	}

	eventch <- watch.Event{
		Type:   watch.Added,
		Object: obj_b,
	}

	eventch <- watch.Event{
		Type:   watch.Modified,
		Object: obj_c,
	}

	eventch <- watch.Event{
		Type:   watch.Deleted,
		Object: obj_d,
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

	select {
	case <-u_ccalled:
	case <-testutil.AsyncWaitch(ctx):
		assert.Fail(t, "unitary create not called")
	}

	select {
	case <-u_ucalled:
	case <-testutil.AsyncWaitch(ctx):
		assert.Fail(t, "unitary update not called")
	}

	select {
	case <-u_dcalled:
	case <-testutil.AsyncWaitch(ctx):
		assert.Fail(t, "unitary delete not called")
	}

	m.Close()
	testutil.AssertDone(t, "monitor", m)

	um.Close()
	testutil.AssertDone(t, "monitor", um)

	controller.Close()
	testutil.AssertDone(t, "controller", controller)

}

func testGenObject(ns, name, vsn string) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       ns,
			Name:            name,
			ResourceVersion: vsn,
		},
	}
}
