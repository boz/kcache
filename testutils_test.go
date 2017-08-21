package kcache

import (
	"context"
	"testing"

	logutil "github.com/boz/go-logutil"
	"github.com/boz/kcache/filter"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
