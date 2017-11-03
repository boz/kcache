package main

import (
	"context"
	"fmt"

	logutil "github.com/boz/go-logutil"
	"github.com/boz/kcache"
	"github.com/boz/kcache/client"
	"github.com/boz/kcache/filter"
	"github.com/cheekybits/genny/generic"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"k8s.io/api/extensions/v1beta1"
)

var (
	ErrInvalidType = fmt.Errorf("invalid type")
	adapter        = _adapter{}
)

var _ = v1beta1.Deployment{}

type ObjectType generic.Type

type Event interface {
	Type() kcache.EventType
	Resource() ObjectType
}

type CacheReader interface {
	Get(ns string, name string) (ObjectType, error)
	List() ([]ObjectType, error)
}

type CacheController interface {
	Cache() CacheReader
	Ready() <-chan struct{}
}

type Subscription interface {
	CacheController
	Events() <-chan Event
	Close()
	Done() <-chan struct{}
}

type Publisher interface {
	Subscribe() (Subscription, error)
	SubscribeWithFilter(filter.Filter) (FilterSubscription, error)
	SubscribeForFilter() (FilterSubscription, error)
	Clone() (Controller, error)
	CloneWithFilter(filter.Filter) (FilterController, error)
	CloneForFilter() (FilterController, error)
}

type Controller interface {
	CacheController
	Publisher
	Done() <-chan struct{}
	Close()
	Error() error
}

type FilterSubscription interface {
	Subscription
	Refilter(filter.Filter) error
}

type FilterController interface {
	Controller
	Refilter(filter.Filter) error
}

type BaseHandler interface {
	OnCreate(ObjectType)
	OnUpdate(ObjectType)
	OnDelete(ObjectType)
}

type Handler interface {
	BaseHandler
	OnInitialize([]ObjectType)
}

type HandlerBuilder interface {
	OnInitialize(func([]ObjectType)) HandlerBuilder
	OnCreate(func(ObjectType)) HandlerBuilder
	OnUpdate(func(ObjectType)) HandlerBuilder
	OnDelete(func(ObjectType)) HandlerBuilder
	Create() Handler
}

type UnitaryHandler interface {
	BaseHandler
	OnInitialize(ObjectType)
}

type UnitaryHandlerBuilder interface {
	OnInitialize(func(ObjectType)) UnitaryHandlerBuilder
	OnCreate(func(ObjectType)) UnitaryHandlerBuilder
	OnUpdate(func(ObjectType)) UnitaryHandlerBuilder
	OnDelete(func(ObjectType)) UnitaryHandlerBuilder
	Create() UnitaryHandler
}

type _adapter struct{}

func (_adapter) adaptObject(obj metav1.Object) (ObjectType, error) {
	if obj, ok := obj.(ObjectType); ok {
		return obj, nil
	}
	return nil, ErrInvalidType
}

func (a _adapter) adaptList(objs []metav1.Object) ([]ObjectType, error) {
	var ret []ObjectType
	for _, orig := range objs {
		adapted, err := a.adaptObject(orig)
		if err != nil {
			continue
		}
		ret = append(ret, adapted)
	}
	return ret, nil
}

func newCache(parent kcache.CacheReader) CacheReader {
	return &cache{parent}
}

type cache struct {
	parent kcache.CacheReader
}

func (c *cache) Get(ns string, name string) (ObjectType, error) {
	obj, err := c.parent.Get(ns, name)
	switch {
	case err != nil:
		return nil, err
	case obj == nil:
		return nil, nil
	default:
		return adapter.adaptObject(obj)
	}
}

func (c *cache) List() ([]ObjectType, error) {
	objs, err := c.parent.List()
	if err != nil {
		return nil, err
	}
	return adapter.adaptList(objs)
}

type event struct {
	etype    kcache.EventType
	resource ObjectType
}

func wrapEvent(evt kcache.Event) (Event, error) {
	obj, err := adapter.adaptObject(evt.Resource())
	if err != nil {
		return nil, err
	}
	return event{evt.Type(), obj}, nil
}

func (e event) Type() kcache.EventType {
	return e.etype
}

func (e event) Resource() ObjectType {
	return e.resource
}

type subscription struct {
	parent kcache.Subscription
	cache  CacheReader
	outch  chan Event
}

func newSubscription(parent kcache.Subscription) *subscription {
	s := &subscription{
		parent: parent,
		cache:  newCache(parent.Cache()),
		outch:  make(chan Event, kcache.EventBufsiz),
	}
	go s.run()
	return s
}

func (s *subscription) run() {
	defer close(s.outch)
	for pevt := range s.parent.Events() {
		evt, err := wrapEvent(pevt)
		if err != nil {
			continue
		}
		select {
		case s.outch <- evt:
		default:
		}
	}
}

func (s *subscription) Cache() CacheReader {
	return s.cache
}

func (s *subscription) Ready() <-chan struct{} {
	return s.parent.Ready()
}

func (s *subscription) Events() <-chan Event {
	return s.outch
}

func (s *subscription) Close() {
	s.parent.Close()
}

func (s *subscription) Done() <-chan struct{} {
	return s.parent.Done()
}

func NewController(ctx context.Context, log logutil.Log, cs kubernetes.Interface, ns string) (Controller, error) {
	client := NewClient(cs, ns)
	return BuildController(ctx, log, client)
}

func BuildController(ctx context.Context, log logutil.Log, client client.Client) (Controller, error) {
	parent, err := kcache.NewController(ctx, log, client)
	if err != nil {
		return nil, err
	}
	return newController(parent), nil
}

func newController(parent kcache.Controller) *controller {
	return &controller{parent, newCache(parent.Cache())}
}

type controller struct {
	parent kcache.Controller
	cache  CacheReader
}

func (c *controller) Close() {
	c.parent.Close()
}

func (c *controller) Ready() <-chan struct{} {
	return c.parent.Ready()
}

func (c *controller) Done() <-chan struct{} {
	return c.parent.Done()
}

func (c *controller) Error() error {
	return c.parent.Error()
}

func (c *controller) Cache() CacheReader {
	return c.cache
}

func (c *controller) Subscribe() (Subscription, error) {
	parent, err := c.parent.Subscribe()
	if err != nil {
		return nil, err
	}
	return newSubscription(parent), nil
}

func (c *controller) SubscribeWithFilter(f filter.Filter) (FilterSubscription, error) {
	parent, err := c.parent.SubscribeWithFilter(f)
	if err != nil {
		return nil, err
	}
	return newFilterSubscription(parent), nil
}

func (c *controller) SubscribeForFilter() (FilterSubscription, error) {
	parent, err := c.parent.SubscribeForFilter()
	if err != nil {
		return nil, err
	}
	return newFilterSubscription(parent), nil
}

func (c *controller) Clone() (Controller, error) {
	parent, err := c.parent.Clone()
	if err != nil {
		return nil, err
	}
	return newController(parent), nil
}

func (c *controller) CloneWithFilter(f filter.Filter) (FilterController, error) {
	parent, err := c.parent.CloneWithFilter(f)
	if err != nil {
		return nil, err
	}
	return newFilterController(parent), nil
}

func (c *controller) CloneForFilter() (FilterController, error) {
	parent, err := c.parent.CloneForFilter()
	if err != nil {
		return nil, err
	}
	return newFilterController(parent), nil
}

type filterController struct {
	controller
	filterParent kcache.FilterController
}

func newFilterController(parent kcache.FilterController) FilterController {
	return &filterController{
		controller:   controller{parent, newCache(parent.Cache())},
		filterParent: parent,
	}
}

func (c *filterController) Refilter(f filter.Filter) error {
	return c.filterParent.Refilter(f)
}

type filterSubscription struct {
	subscription
	filterParent kcache.FilterSubscription
}

func newFilterSubscription(parent kcache.FilterSubscription) FilterSubscription {
	return &filterSubscription{
		subscription: *newSubscription(parent),
		filterParent: parent,
	}
}

func (s *filterSubscription) Refilter(f filter.Filter) error {
	return s.filterParent.Refilter(f)
}

func NewMonitor(publisher Publisher, handler Handler) (kcache.Monitor, error) {
	phandler := kcache.BuildHandler().
		OnInitialize(func(objs []metav1.Object) {
			aobjs, _ := adapter.adaptList(objs)
			handler.OnInitialize(aobjs)
		}).
		OnCreate(func(obj metav1.Object) {
			aobj, _ := adapter.adaptObject(obj)
			handler.OnCreate(aobj)
		}).
		OnUpdate(func(obj metav1.Object) {
			aobj, _ := adapter.adaptObject(obj)
			handler.OnUpdate(aobj)
		}).
		OnDelete(func(obj metav1.Object) {
			aobj, _ := adapter.adaptObject(obj)
			handler.OnDelete(aobj)
		}).Create()

	switch obj := publisher.(type) {
	case *controller:
		return kcache.NewMonitor(obj.parent, phandler)
	case *filterController:
		return kcache.NewMonitor(obj.parent, phandler)
	default:
		panic(fmt.Sprintf("Invalid publisher type: %T is not a *controller", publisher))
	}
}

func ToUnitary(log logutil.Log, delegate UnitaryHandler) Handler {
	return BuildHandler().
		OnInitialize(func(objs []ObjectType) {
			if count := len(objs); count > 1 {
				log.Warnf("initialized with invalid count: %v", count)
				return
			}
			if count := len(objs); count == 0 {
				log.Debugf("initialized with empty result, ignoring")
				return
			}
			delegate.OnInitialize(objs[0])
		}).
		OnCreate(func(obj ObjectType) {
			delegate.OnCreate(obj)
		}).
		OnUpdate(func(obj ObjectType) {
			delegate.OnUpdate(obj)
		}).
		OnDelete(func(obj ObjectType) {
			delegate.OnDelete(obj)
		}).Create()
}

func BuildHandler() HandlerBuilder {
	return &handlerBuilder{}
}

func BuildUnitaryHandler() UnitaryHandlerBuilder {
	return &unitaryHandlerBuilder{}
}

type baseHandler struct {
	onCreate func(ObjectType)
	onUpdate func(ObjectType)
	onDelete func(ObjectType)
}

type handler struct {
	baseHandler
	onInitialize func([]ObjectType)
}
type handlerBuilder handler

type unitaryHandler struct {
	baseHandler
	onInitialize func(ObjectType)
}
type unitaryHandlerBuilder unitaryHandler

func (hb *handlerBuilder) OnInitialize(fn func([]ObjectType)) HandlerBuilder {
	hb.onInitialize = fn
	return hb
}

func (hb *handlerBuilder) OnCreate(fn func(ObjectType)) HandlerBuilder {
	hb.onCreate = fn
	return hb
}

func (hb *handlerBuilder) OnUpdate(fn func(ObjectType)) HandlerBuilder {
	hb.onUpdate = fn
	return hb
}

func (hb *handlerBuilder) OnDelete(fn func(ObjectType)) HandlerBuilder {
	hb.onDelete = fn
	return hb
}

func (hb *handlerBuilder) Create() Handler {
	return handler(*hb)
}

func (h handler) OnInitialize(objs []ObjectType) {
	if h.onInitialize != nil {
		h.onInitialize(objs)
	}
}

func (hb *unitaryHandlerBuilder) OnInitialize(fn func(ObjectType)) UnitaryHandlerBuilder {
	hb.onInitialize = fn
	return hb
}

func (hb *unitaryHandlerBuilder) OnCreate(fn func(ObjectType)) UnitaryHandlerBuilder {
	hb.onCreate = fn
	return hb
}

func (hb *unitaryHandlerBuilder) OnUpdate(fn func(ObjectType)) UnitaryHandlerBuilder {
	hb.onUpdate = fn
	return hb
}

func (hb *unitaryHandlerBuilder) OnDelete(fn func(ObjectType)) UnitaryHandlerBuilder {
	hb.onDelete = fn
	return hb
}

func (hb *unitaryHandlerBuilder) Create() UnitaryHandler {
	return unitaryHandler(*hb)
}

func (h unitaryHandler) OnInitialize(obj ObjectType) {
	if h.onInitialize != nil {
		h.onInitialize(obj)
	}
}

func (h baseHandler) OnCreate(obj ObjectType) {
	if h.onCreate != nil {
		h.onCreate(obj)
	}
}

func (h baseHandler) OnUpdate(obj ObjectType) {
	if h.onUpdate != nil {
		h.onUpdate(obj)
	}
}

func (h baseHandler) OnDelete(obj ObjectType) {
	if h.onDelete != nil {
		h.onDelete(obj)
	}
}
