package ingress

import (
	"github.com/boz/kcache"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

type event struct {
	etype    kcache.EventType
	resource *v1beta1.Ingress
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

func (e event) Resource() *v1beta1.Ingress {
	return e.resource
}
