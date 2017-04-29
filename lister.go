package kcache

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	lifecycle "github.com/boz/go-lifecycle"
	logutil "github.com/boz/go-logutil"
	"github.com/boz/kcache/client"
)

var (
	errInvalidType       = fmt.Errorf("Invalid type")
	defaultRefreshPeriod = time.Minute
)

type lister interface {
	Result() <-chan listResult
}

type listResult struct {
	list runtime.Object
	err  error
}

type _lister struct {
	client   client.ListClient
	period   time.Duration
	resultch chan listResult

	log logutil.Log
	lc  lifecycle.Lifecycle
	ctx context.Context
}

func newLister(ctx context.Context, log logutil.Log, stopch <-chan struct{}, period time.Duration, client client.ListClient) *_lister {
	log = log.WithComponent("lister")

	l := &_lister{
		client:   client,
		period:   period,
		resultch: make(chan listResult),
		log:      log,
		lc:       lifecycle.New(),
		ctx:      ctx,
	}

	go l.lc.WatchContext(ctx)
	go l.lc.WatchChannel(stopch)

	go l.run()

	return l
}

func (l *_lister) Result() <-chan listResult {
	return l.resultch
}

func (l *_lister) run() {
	defer l.log.Un(l.log.Trace("list"))
	defer l.lc.ShutdownCompleted()

	var tickch <-chan time.Time
	var ticker *time.Ticker

	var resultch chan listResult

	var result listResult

	runch := l.list()

	for {
		select {
		case <-tickch:
			l.drainTicker(ticker)
			runch = l.list()
			tickch = nil
			ticker = nil

		case result = <-runch:
			resultch = l.resultch
			runch = nil

		case resultch <- result:
			ticker = time.NewTicker(l.period)
			tickch = ticker.C
			resultch = nil

		case <-l.lc.ShutdownRequest():
			l.lc.ShutdownInitiated()
			l.drainTicker(ticker)
			return
		}
	}
}

func (l *_lister) list() <-chan listResult {
	defer l.log.Un(l.log.Trace("list"))
	runch := make(chan listResult, 1)

	go func() {
		ctx, cancel := context.WithCancel(l.ctx)
		defer cancel()
		select {
		case runch <- l.executeList(ctx):
		case <-l.lc.ShuttingDown():
		}
	}()

	return runch
}

func (l *_lister) executeList(ctx context.Context) listResult {
	defer l.log.Un(l.log.Trace("executeList"))

	list, err := l.client.List(ctx, metav1.ListOptions{})
	if err != nil {
		l.log.ErrWarn(err, "client.List()")
		return listResult{nil, err}
	}

	if _, ok := list.(meta.List); !ok {
		l.log.Warnf("invalid type: %T", list)
		return listResult{nil, errInvalidType}
	}

	return listResult{list, nil}
}

func (l *_lister) drainTicker(ticker *time.Ticker) {
	if ticker == nil {
		return
	}

	ticker.Stop()

	select {
	case <-ticker.C:
	default:
	}
}
