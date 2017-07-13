package secret

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
)

var (
	ErrInvalidType = fmt.Errorf("invalid type")
	adapter        = _adapter{}
)

type _adapter struct{}

func (_adapter) adaptObject(obj metav1.Object) (*v1.Secret, error) {
	if obj, ok := obj.(*v1.Secret); ok {
		return obj, nil
	}
	return nil, ErrInvalidType
}

func (a _adapter) adaptList(objs []metav1.Object) ([]*v1.Secret, error) {
	var ret []*v1.Secret
	for _, orig := range objs {
		adapted, err := a.adaptObject(orig)
		if err != nil {
			continue
		}
		ret = append(ret, adapted)
	}
	return ret, nil
}
