package kcache

import (
	"context"
	"testing"
	"time"

	logutil "github.com/boz/go-logutil"
	"github.com/boz/kcache/filter"
	"github.com/stretchr/testify/assert"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	testAsyncWaitDuration = time.Millisecond
)

type testReadyable interface {
	Ready() <-chan struct{}
}

type testDoneable interface {
	Done() <-chan struct{}
}

func testTimerch(ctx context.Context, duration time.Duration) <-chan time.Time {
	t := time.NewTimer(duration)
	go func() {
		<-ctx.Done()
		t.Stop()
		select {
		case <-t.C:
		default:
		}
	}()
	return t.C
}

func testAsyncWaitch(ctx context.Context) <-chan time.Time {
	return testTimerch(ctx, testAsyncWaitDuration)
}

func testGenPod(ns, name, vsn string) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       ns,
			Name:            name,
			ResourceVersion: vsn,
		},
	}
}

func testGenEvent(et EventType, ns, name, vsn string) Event {
	return NewEvent(et, testGenPod(ns, name, vsn))
}

func testNewSubscription(t *testing.T, log logutil.Log, f filter.Filter) (subscription, cache, chan struct{}) {

	ctx, cancel := context.WithCancel(context.Background())
	readych := make(chan struct{})
	cache := newCache(ctx, log, nil, f)

	sub := newSubscription(log, nil, readych, cache)

	go func() {
		<-sub.Done()
		cancel()
	}()

	return sub, cache, readych

}

func testAssertReady(t *testing.T, name string, obj testReadyable) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	select {
	case <-obj.Ready():
	case <-testAsyncWaitch(ctx):
		assert.Fail(t, "expected to be ready but wasn't: %v", name)
	}
}

func testAssertNotReady(t *testing.T, name string, obj testReadyable) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	select {
	case <-obj.Ready():
		assert.Fail(t, "expected to not be ready but was: %v", name)
	case <-testAsyncWaitch(ctx):
	}
}

func testAssertDone(t *testing.T, name string, obj testDoneable) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	select {
	case <-obj.Done():
	case <-testAsyncWaitch(ctx):
		assert.Fail(t, "expected to be done but wasn't: %v", name)
	}
}
func testAssertNotDone(t *testing.T, name string, obj testDoneable) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	select {
	case <-obj.Done():
		assert.Fail(t, "expected to be not done but wasn: %v", name)
	case <-testAsyncWaitch(ctx):
	}
}
