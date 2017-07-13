package ingress

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

var (
	ErrInvalidType = fmt.Errorf("invalid type")
	adapter        = _adapter{}
)

type _adapter struct{}

func (_adapter) adaptObject(obj metav1.Object) (*v1beta1.Ingress, error) {
	if obj, ok := obj.(*v1beta1.Ingress); ok {
		return obj, nil
	}
	return nil, ErrInvalidType
}

func (a _adapter) adaptList(objs []metav1.Object) ([]*v1beta1.Ingress, error) {
	var ret []*v1beta1.Ingress
	for _, orig := range objs {
		adapted, err := a.adaptObject(orig)
		if err != nil {
			continue
		}
		ret = append(ret, adapted)
	}
	return ret, nil
}
