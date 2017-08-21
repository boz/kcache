package testutil

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var testAsyncWaitDuration time.Duration

func init() {
	dstr := os.Getenv("KCACHE_TEST_ASYNC_DURATION")
	if dstr == "" {
		dstr = "10ms"
	}
	d, err := time.ParseDuration(dstr)
	if err != nil {
		panic("invalid KCACHE_TEST_ASYNC_DURATION: " + err.Error())
	}
	testAsyncWaitDuration = d
}

type readyable interface {
	Ready() <-chan struct{}
}

type doneable interface {
	Done() <-chan struct{}
}

func Timerch(ctx context.Context, duration time.Duration) <-chan time.Time {
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

func AsyncWaitch(ctx context.Context) <-chan time.Time {
	return Timerch(ctx, testAsyncWaitDuration)
}

func AssertReady(t *testing.T, name string, obj readyable) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	select {
	case <-obj.Ready():
	case <-AsyncWaitch(ctx):
		assert.Fail(t, "expected to be ready but wasn't: %v", name)
	}
}

func AssertNotReady(t *testing.T, name string, obj readyable) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	select {
	case <-obj.Ready():
		assert.Fail(t, "expected to not be ready but was: %v", name)
	case <-AsyncWaitch(ctx):
	}
}

func AssertDone(t *testing.T, name string, obj doneable) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	select {
	case <-obj.Done():
	case <-AsyncWaitch(ctx):
		assert.Fail(t, "expected to be done but wasn't: %v", name)
	}
}

func AssertNotDone(t *testing.T, name string, obj doneable) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	select {
	case <-obj.Done():
		assert.Fail(t, "expected to be not done but wasn: %v", name)
	case <-AsyncWaitch(ctx):
	}
}
