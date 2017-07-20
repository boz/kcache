package event

import (
	"github.com/boz/kcache/filter"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/pkg/api/v1"
)

type Object interface {
	GetObjectKind() schema.ObjectKind
	GetNamespace() string
	GetName() string
}

func InvolvedObjectFilter(obj Object) filter.ComparableFilter {
	kind := obj.GetObjectKind().GroupVersionKind().Kind
	return InvolvedFilter(kind, obj.GetNamespace(), obj.GetName())
}

func InvolvedFilter(kind, ns, name string) filter.ComparableFilter {
	return &involvedFilter{kind, ns, name}
}

type involvedFilter struct {
	kind string
	ns   string
	name string
}

func (f *involvedFilter) Accept(obj metav1.Object) bool {
	event, ok := obj.(*v1.Event)
	if !ok {
		return false
	}
	ref := event.InvolvedObject
	return ref.Kind == f.kind &&
		ref.Namespace == f.ns &&
		ref.Name == f.name
}

func (f *involvedFilter) Equals(other filter.Filter) bool {
	if other, ok := other.(*involvedFilter); ok {
		return *f == *other
	}
	return false
}
