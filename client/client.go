package client

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type ListFn func(context.Context, v1.ListOptions) (runtime.Object, error)
type WatchFn func(context.Context, v1.ListOptions) (watch.Interface, error)

type ListClient interface {
	List(context.Context, v1.ListOptions) (runtime.Object, error)
}

type WatchClient interface {
	Watch(context.Context, v1.ListOptions) (watch.Interface, error)
}

type Client interface {
	ListClient
	WatchClient
}

type client struct {
	list  ListFn
	watch WatchFn
}

func NewListClient(fn ListFn) ListClient {
	return &client{list: fn}
}

func NewWatchClient(fn WatchFn) WatchClient {
	return &client{watch: fn}
}

func NewClient(list ListFn, watch WatchFn) Client {
	return &client{list, watch}
}

func (c *client) List(ctx context.Context, opts v1.ListOptions) (runtime.Object, error) {
	return c.list(ctx, opts)
}

func (c *client) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.watch(ctx, opts)
}

type restRequester interface {
	Get() *rest.Request
}

func ForResource(
	c restRequester, res string, ns string, fsel fields.Selector) Client {
	return NewClient(
		makeResourceListFn(c, res, ns, fsel),
		makeResourceWatchFn(c, res, ns, fsel),
	)
}

func makeResourceListFn(
	c restRequester, res string, ns string, fsel fields.Selector) ListFn {
	return func(ctx context.Context, opts v1.ListOptions) (runtime.Object, error) {
		return c.Get().
			Context(ctx).
			Namespace(ns).
			Resource(res).
			VersionedParams(&opts, scheme.ParameterCodec).
			FieldsSelectorParam(fsel).
			Do().
			Get()
	}
}

func makeResourceWatchFn(
	c restRequester, res string, ns string, fsel fields.Selector) WatchFn {

	return func(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
		return c.Get().
			Context(ctx).
			Prefix("watch").
			Namespace(ns).
			Resource(res).
			VersionedParams(&opts, scheme.ParameterCodec).
			FieldsSelectorParam(fsel).
			Watch()
	}
}
