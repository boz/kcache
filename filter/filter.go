package filter

import (
	"reflect"

	"github.com/boz/kcache/nsname"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/pkg/api/v1"
)

type Filter interface {
	Accept(metav1.Object) bool
}

type ComparableFilter interface {
	Filter
	Equals(Filter) bool
}

func Null() ComparableFilter {
	return nullFilter{}
}

type nullFilter struct{}

func (nullFilter) Accept(_ metav1.Object) bool {
	return true
}

func (nullFilter) Equals(other Filter) bool {
	_, ok := other.(nullFilter)
	return ok
}

func LabelsSelector(match map[string]string) ComparableFilter {
	return &labelsSelector{labels.Set(match)}
}

type labelsSelector struct {
	filter labels.Set
}

func (f *labelsSelector) Accept(obj metav1.Object) bool {
	if len(f.filter) == 0 {
		return false
	}
	return labels.AreLabelsInWhiteList(f.filter, labels.Set(obj.GetLabels()))
}

func (f *labelsSelector) Equals(other Filter) bool {
	if other, ok := other.(*labelsSelector); ok {
		return labels.Equals(f.filter, other.filter)
	}
	return false
}

func ServiceSelector(target map[string]string) ComparableFilter {
	return &serviceSelector{labels.Set(target)}
}

type serviceSelector struct {
	target labels.Set
}

func (f *serviceSelector) Accept(obj metav1.Object) bool {
	svc, ok := obj.(*v1.Service)

	if !ok {
		return false
	}

	if len(svc.Spec.Selector) == 0 {
		return false
	}

	for k, v := range svc.Spec.Selector {
		if f.target.Get(k) != v {
			return false
		}
	}

	return false
}

func (f *serviceSelector) Equals(other Filter) bool {
	if other, ok := other.(*serviceSelector); ok {
		return labels.Equals(f.target, other.target)
	}
	return false
}

func NSNamesSelector(ids ...nsname.NSName) ComparableFilter {
	set := make(map[nsname.NSName]bool)
	for _, id := range ids {
		set[id] = true
	}
	return nsNameSelector(set)
}

type nsNameSelector map[nsname.NSName]bool

func (f nsNameSelector) Accept(obj metav1.Object) bool {
	key := nsname.ForObject(obj)
	_, ok := f[key]
	return ok
}

func (f nsNameSelector) Equals(other Filter) bool {
	return reflect.DeepEqual(f, other)
}
