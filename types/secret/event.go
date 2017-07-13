package secret

import (
	"github.com/boz/kcache"
	"k8s.io/client-go/pkg/api/v1"
)

type event struct {
	etype    kcache.EventType
	resource *v1.Secret
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

func (e event) Resource() *v1.Secret {
	return e.resource
}
