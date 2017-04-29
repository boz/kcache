package kcache

import (
	"context"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	lifecycle "github.com/boz/go-lifecycle"
	logutil "github.com/boz/go-logutil"
	"github.com/boz/kcache/client"
)

type watcher interface {
	reset(string)
	events() <-chan Event
}

type _watcher struct {
	version string

	client client.WatchClient

	resetch chan string
	outch   chan Event

	log logutil.Log
	lc  lifecycle.Lifecycle
	ctx context.Context
}

func newWatcher(ctx context.Context, log logutil.Log, stopch <-chan struct{}, client client.WatchClient) watcher {
	log = log.WithComponent("watcher")
	lc := lifecycle.New()

	w := &_watcher{
		client:  client,
		resetch: make(chan string),
		outch:   make(chan Event),
		log:     log,
		lc:      lc,
		ctx:     ctx,
	}

	go w.lc.WatchContext(ctx)
	go w.lc.WatchChannel(stopch)
	go w.run()
	return w
}

func (w *_watcher) reset(vsn string) {
	select {
	case w.resetch <- vsn:
	case <-w.lc.ShuttingDown():
	}
}

func (w *_watcher) events() <-chan Event {
	return w.outch
}

func (w *_watcher) run() {
	defer w.lc.ShutdownCompleted()
	defer w.lc.ShutdownInitiated()

	ctx, cancel := context.WithCancel(w.ctx)
	defer cancel()

	var watchstream watch.Interface
	var eventsch <-chan watch.Event

	var buf []Event

	for {
		var evt Event
		var outch chan Event

		if len(buf) > 0 {
			evt = buf[0]
			outch = w.outch
		}

		select {
		case <-w.lc.ShutdownRequest():
			return

		case vsn := <-w.resetch:

			if watchstream != nil {
				watchstream.Stop()
			}

			ws, err := w.startWatch(ctx, vsn)
			if err != nil {
				w.log.Err(err, "startWatch")
				return
			}

			buf = make([]Event, 0)

			watchstream = ws
			eventsch = watchstream.ResultChan()

		case evt := <-eventsch:

			obj, err := meta.Accessor(evt.Object)
			if err != nil {
				w.log.ErrWarn(err, "meta.Accessor(%T)", evt.Object)
				continue
			}

			switch evt.Type {
			case watch.Added:
				buf = append(buf, NewEvent(EventTypeCreate, obj))
			case watch.Modified:
				buf = append(buf, NewEvent(EventTypeUpdate, obj))
			case watch.Deleted:
				buf = append(buf, NewEvent(EventTypeDelete, obj))
			}

		case outch <- evt:
			buf = buf[1:]
		}
	}
}

func (w *_watcher) startWatch(ctx context.Context, version string) (watch.Interface, error) {
	response, err := w.client.Watch(ctx, metav1.ListOptions{
		ResourceVersion: version,
		Watch:           true,
	})
	return response, err
}
