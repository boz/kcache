package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

type Client struct {
	mock.Mock
}

func (c *Client) List(ctx context.Context, opts metav1.ListOptions) (runtime.Object, error) {
	args := c.Called(ctx, opts)
	return args.Get(0).(runtime.Object), args.Error(1)
}

func (c *Client) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	args := c.Called(ctx, opts)
	return args.Get(0).(watch.Interface), args.Error(1)
}

type WatchInterface struct {
	mock.Mock
}

func (w *WatchInterface) Stop() {
	w.Called()
}

func (w *WatchInterface) ResultChan() <-chan watch.Event {
	args := w.Called()
	return args.Get(0).(chan watch.Event)
}
